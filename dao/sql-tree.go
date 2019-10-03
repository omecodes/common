package dao

import (
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/errors"
	"path"
)

type SQLTree struct {
	SQL
}

func NewSQLTree(dbCfg conf.Map, prefix string) (*SQLTree, error) {
	db := new(SQLTree)

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
				foreign key (parent) references $prefix$_directories(id) on delete cascade
			);`).
		AddStatement("clear", `DELETE FROM $prefix$_trees;`).
		AddStatement("insert_tree_path", `INSERT INTO $prefix$_trees (path) VALUES(?);`).
		AddStatement("insert_encoded", `INSERT INTO $prefix$_encoded VALUES(?, ?, ?, ?, ?, ?, ?, ?);`).
		AddStatement("update_encoded", `UPDATE $prefix$_encoded SET encoded=? WHERE parent=? and node_name=?;`).
		AddStatement("update_tree_path", `UPDATE $prefix$_trees SET path=replace(path, ?, ?) WHERE path LIKE ?;`).
		//AddStatement("update_parent", `UPDATE $prefix$_files SET parent=? WHERE parent=?;`).
		AddStatement("update_parent", `UPDATE $prefix$_encoded SET parent=? WHERE parent=? AND node_name=?;`).
		AddStatement("update_name", `UPDATE $prefix$_files SET node_name=? WHERE node_name=? AND parent=?;`).
		AddStatement("update_tree_path", `UPDATE $prefix$_trees SET path=? WHERE path=?;`).
		AddStatement("delete_tree_path", `DELETE FROM $prefix$_trees WHERE path LIKE ?;`).
		AddStatement("delete_encoded", `DELETE FROM $prefix$_encoded WHERE parent=? AND node_name=?;`).
		AddStatement("select_tree_path", `SELECT * FROM $prefix$_trees WHERE path LIKE ?;`).
		AddStatement("select_tree_path_id", `SELECT id FROM $prefix$_trees WHERE path=?;`).
		AddStatement("select_encoded", `SELECT encoded FROM $prefix$_encoded WHERE parent=? AND node_name=?;`).
		AddStatement("select_all_encoded", `SELECT encoded FROM $prefix$_encoded WHERE parent=? ORDER BY node_name OFFSET ? ROWS;`).
		RegisterScanner("tree_path_scanner", NewScannerFunc(db.scanNodePathID)).
		RegisterScanner("encoded_scanner", NewScannerFunc(db.scanEncoded))

	err := db.Init(dbCfg)
	if err != nil {
		return nil, err
	}

	_ = db.AddUniqueIndex(SQLIndex{}, false)

	return db, db.init()
}

func (dao *SQLTree) init() error {
	_, err := dao.getNodePathID("/")
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		return dao.Exec("insert_tree_path", "/").Error
	}
	return nil
}

func (dao *SQLTree) Clear() error {
	res := dao.Exec("clear")
	return res.Error
}

func (dao *SQLTree) getNodePathID(p string) (int, error) {
	i, err := dao.QueryOne("get_tree_path_id", "tree_path_scanner", p)
	if err != nil {
		return 0, err
	}
	return i.(int), err
}

func (dao *SQLTree) CreateNode(nodePath string, encoded string, isLeaf bool) error {
	parent, name := path.Split(nodePath)
	var e error
	parentID, e := dao.getNodePathID(parent)
	if e != nil {
		return e
	}

	err := dao.Exec("insert_encoded", parentID, name, encoded).Error
	if err != nil {
		return err
	}

	if !isLeaf {
		fullPath := path.Join(parent, name)
		return dao.Exec("insert_tree_path", fullPath).Error
	}
	return nil
}

func (dao *SQLTree) MoveNode(src string, dst string) error {
	if src == "" || dst == "" {
		return errors.New("bad input")
	}

	pattern := src + "%"

	srcBase := path.Base(src)
	srcParent := path.Dir(src)
	srcParentID, e := dao.getNodePathID(srcParent)
	if e != nil {
		return e
	}

	dstBase := path.Base(dst)
	dstParent := path.Dir(dst)
	dstParentID, e := dao.getNodePathID(dstParent)
	if e != nil {
		return e
	}

	err := dao.Exec("update_tree_path", src, dst, pattern).Error
	if err != nil {
		return err
	}

	if dstBase != srcBase {
		err = dao.Exec("update_name", path.Base(dst), path.Base(src), srcParentID).Error
		if err != nil {
			return err
		}
	}

	return dao.Exec("update_parent", dstParentID, srcParentID, path.Base(src)).Error
}

func (dao *SQLTree) RenameNode(src string, newName string) error {
	parent := path.Dir(src)
	newPath := path.Join(parent, newName)
	return dao.MoveNode(src, newPath)
}

func (dao *SQLTree) DeleteNode(nodePath string) error {
	name := path.Base(nodePath)
	parent := path.Dir(nodePath)

	nodePathID, _ := dao.getNodePathID(nodePath)
	isLeaf := nodePathID == 0

	parentID, e := dao.getNodePathID(parent)
	if e != nil {
		return e
	}

	if err := dao.Exec("delete_encoded", parentID, name).Error; err != nil {
		return err
	}

	if !isLeaf {
		return dao.Exec("delete_tree_path", nodePath+"/%").Error
	}
	return nil
}

func (dao *SQLTree) UpdateNode(nodePath string, encoded string) error {
	parentID, e := dao.getNodePathID(path.Dir(nodePath))
	if e != nil {
		return e
	}
	return dao.Exec("update_encoded", encoded, parentID, path.Base(nodePath)).Error
}

func (dao *SQLTree) Encoded(nodePath string) (string, error) {

	name := path.Base(nodePath)
	parent := path.Dir(nodePath)

	parentID, e := dao.getNodePathID(parent)
	if e != nil {
		return "", e
	}

	item, err := dao.QueryOne("get_encoded", "encoded_scanner", parentID, name)
	if err != nil {
		return "", err
	}
	return item.(string), err
}

func (dao *SQLTree) List(dir string, offset int) (Cursor, error) {
	parentID, e := dao.getNodePathID(dir)
	if e != nil {
		return nil, e
	}
	return dao.Query("ls", "encoded_scanner", parentID, offset)
}

func (dao *SQLTree) scanNodePathID(row Row) (interface{}, error) {
	var id int
	err := row.Scan(&id)
	return id, err
}

func (dao *SQLTree) scanEncoded(row Row) (interface{}, error) {
	var encoded string
	err := row.Scan(&encoded)
	return encoded, err
}
