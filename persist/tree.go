package persist

import (
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/dao"
	"github.com/zoenion/common/errors"
	"path"
)

type Tree interface {
	CreateNode(nodePath string, content string, isLeaf bool) error
	MoveNode(nodePath string, newPath string) error
	RenameNode(nodePath string, newName string) error
	DeleteNode(nodePath string) error
	UpdateNode(nodePath string, content string) error
	Content(nodePath string) (string, error)
	Children(nodePath string) (dao.Cursor, error)
	Clear() error
}

type sqlTree struct {
	dao.SQL
}

func NewSQLTree(dbCfg conf.Map, prefix string) (*sqlTree, error) {
	db := new(sqlTree)

	db.SetTablePrefix(prefix).
		AddTableDefinition("directories", `CREATE TABLE IF NOT EXISTS $prefix$_trees (
				id INTEGER NOT NULL PRIMARY KEY $auto_increment$,
				path VARCHAR(2048) NOT NULL
			);
		`).
		AddTableDefinition("files", `CREATE TABLE IF NOT EXISTS $prefix$_encoded (
				parent INT NOT NULL,
				node_name VARCHAR(255) NOT NULL,
				encoded TEXT NOT NULL,
				foreign key (parent) references $prefix$_trees(id) on delete cascade
			);`).
		AddStatement("clear", `DELETE FROM $prefix$_trees;`).
		AddStatement("insert_tree_path", `INSERT INTO $prefix$_trees (path) VALUES(?);`).
		AddStatement("insert_encoded", `INSERT INTO $prefix$_encoded VALUES(?, ?, ?);`).
		AddStatement("update_encoded", `UPDATE $prefix$_encoded SET encoded=? WHERE parent=? and node_name=?;`).
		AddStatement("update_tree_path", `UPDATE $prefix$_trees SET path=replace(path, ?, ?) WHERE path LIKE ?;`).
		//AddStatement("update_parent", `UPDATE $prefix$_encoded SET parent=? WHERE parent=?;`).
		AddStatement("update_parent", `UPDATE $prefix$_encoded SET parent=? WHERE parent=? AND node_name=?;`).
		AddStatement("update_name", `UPDATE $prefix$_encoded SET node_name=? WHERE node_name=? AND parent=?;`).
		AddStatement("update_tree_path", `UPDATE $prefix$_trees SET path=? WHERE path=?;`).
		AddStatement("delete_tree_path", `DELETE FROM $prefix$_trees WHERE path LIKE ?;`).
		AddStatement("delete_encoded", `DELETE FROM $prefix$_encoded WHERE parent=? AND node_name=?;`).
		AddStatement("select_tree_path", `SELECT * FROM $prefix$_trees WHERE path LIKE ?;`).
		AddStatement("select_tree_path_id", `SELECT id FROM $prefix$_trees WHERE path=?;`).
		AddStatement("select_encoded", `SELECT encoded FROM $prefix$_encoded WHERE parent=? AND node_name=?;`).
		AddStatement("select_all_encoded", `SELECT encoded FROM $prefix$_encoded WHERE parent=? ORDER BY node_name;`).
		RegisterScanner("tree_path_scanner", dao.NewScannerFunc(db.intScanner)).
		RegisterScanner("encoded_scanner", dao.NewScannerFunc(db.stringScanner))

	err := db.Init(dbCfg)
	if err != nil {
		return nil, err
	}

	_ = db.AddUniqueIndex(dao.SQLIndex{}, false)

	return db, db.init()
}

func (t *sqlTree) init() error {
	_, err := t.getNodePathID("/")
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		return t.Exec("insert_tree_path", "/").Error
	}
	return nil
}

func (t *sqlTree) Clear() error {
	res := t.Exec("clear")
	return res.Error
}

func (t *sqlTree) getNodePathID(p string) (int, error) {
	i, err := t.QueryOne("select_tree_path_id", "tree_path_scanner", p)
	if err != nil {
		return 0, err
	}
	return i.(int), err
}

func (t *sqlTree) CreateNode(nodePath string, encoded string, isLeaf bool) error {
	parent, name := path.Split(nodePath)
	var e error
	parentID, e := t.getNodePathID(parent)
	if e != nil {
		return e
	}

	err := t.Exec("insert_encoded", parentID, name, encoded).Error
	if err != nil {
		return err
	}

	if !isLeaf {
		fullPath := path.Join(parent, name)
		return t.Exec("insert_tree_path", fullPath).Error
	}
	return nil
}

func (t *sqlTree) MoveNode(src string, dst string) error {
	if src == "" || dst == "" {
		return errors.New("bad input")
	}

	pattern := src + "%"

	srcBase := path.Base(src)
	srcParent := path.Dir(src)
	srcParentID, e := t.getNodePathID(srcParent)
	if e != nil {
		return e
	}

	dstBase := path.Base(dst)
	dstParent := path.Dir(dst)
	dstParentID, e := t.getNodePathID(dstParent)
	if e != nil {
		return e
	}

	err := t.Exec("update_tree_path", src, dst, pattern).Error
	if err != nil {
		return err
	}

	if dstBase != srcBase {
		err = t.Exec("update_name", path.Base(dst), path.Base(src), srcParentID).Error
		if err != nil {
			return err
		}
	}

	return t.Exec("update_parent", dstParentID, srcParentID, path.Base(src)).Error
}

func (t *sqlTree) RenameNode(src string, newName string) error {
	parent := path.Dir(src)
	newPath := path.Join(parent, newName)
	return t.MoveNode(src, newPath)
}

func (t *sqlTree) DeleteNode(nodePath string) error {
	name := path.Base(nodePath)
	parent := path.Dir(nodePath)

	nodePathID, _ := t.getNodePathID(nodePath)
	isLeaf := nodePathID == 0

	parentID, e := t.getNodePathID(parent)
	if e != nil {
		return e
	}

	if err := t.Exec("delete_encoded", parentID, name).Error; err != nil {
		return err
	}

	if !isLeaf {
		return t.Exec("delete_tree_path", nodePath+"/%").Error
	}
	return nil
}

func (t *sqlTree) UpdateNode(nodePath string, encoded string) error {
	parentID, e := t.getNodePathID(path.Dir(nodePath))
	if e != nil {
		return e
	}
	return t.Exec("update_encoded", encoded, parentID, path.Base(nodePath)).Error
}

func (t *sqlTree) Content(nodePath string) (string, error) {

	name := path.Base(nodePath)
	parent := path.Dir(nodePath)

	parentID, e := t.getNodePathID(parent)
	if e != nil {
		return "", e
	}

	item, err := t.QueryOne("select_encoded", "encoded_scanner", parentID, name)
	if err != nil {
		return "", err
	}
	return item.(string), err
}

func (t *sqlTree) Children(dir string) (dao.Cursor, error) {
	parentID, e := t.getNodePathID(dir)
	if e != nil {
		return nil, e
	}
	return t.Query("select_all_encoded", "encoded_scanner", parentID)
}

func (t *sqlTree) intScanner(row dao.Row) (interface{}, error) {
	var id int
	err := row.Scan(&id)
	return id, err
}

func (t *sqlTree) stringScanner(row dao.Row) (interface{}, error) {
	var content string
	err := row.Scan(&content)
	return content, err
}
