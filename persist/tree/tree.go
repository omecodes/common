package tree

import (
	"github.com/zoenion/common/codec"
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/dao"
	"github.com/zoenion/common/errors"
	"path"
)

type Tree interface {
	CreateNode(nodePath string, o interface{}, isLeaf bool) error
	MoveNode(nodePath string, newPath string) error
	RenameNode(nodePath string, newName string) error
	DeleteNode(nodePath string) error
	UpdateNode(nodePath string, o interface{}) error
	ReadNode(nodePath string, o interface{}) error
	Children(nodePath string) (Cursor, error)
	Clear() error
}

type sqlTree struct {
	dao.SQL
	codec codec.Codec
}

func NewSQL(dbCfg conf.Map, prefix string, codec codec.Codec) (*sqlTree, error) {
	db := new(sqlTree)
	db.codec = codec

	db.SetTablePrefix(prefix).
		AddTableDefinition("directories", `CREATE TABLE IF NOT EXISTS $prefix$_trees (
				id INTEGER NOT NULL PRIMARY KEY $auto_increment$,
				path VARCHAR(2048) NOT NULL
			);
		`).
		AddTableDefinition("encoded", `CREATE TABLE IF NOT EXISTS $prefix$_encoded (
				parent INT NOT NULL,
				node_name VARCHAR(255) NOT NULL,
				encoded longblob NOT NULL,
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
		AddStatement("select_tree_path_id", `SELECT * FROM $prefix$_trees WHERE path=?;`).
		AddStatement("select_encoded", `SELECT encoded FROM $prefix$_encoded WHERE parent=? AND node_name=?;`).
		AddStatement("select_all_encoded", `SELECT encoded FROM $prefix$_encoded WHERE parent=? ORDER BY node_name;`).
		RegisterScanner("tree", dao.NewScannerFunc(scanTreeRow)).
		RegisterScanner("encoded", dao.NewScannerFunc(scanEncodedRow))

	err := db.Init(dbCfg)
	if err != nil {
		return nil, err
	}

	_ = db.AddUniqueIndex(dao.SQLIndex{}, false)

	return db, db.init()
}

func (t *sqlTree) init() error {
	_, err := t.getTreeByPath("/")
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

func (t *sqlTree) getTreeByPath(p string) (*TreeRow, error) {
	i, err := t.QueryOne("select_tree_path_id", "tree", p)
	if err != nil {
		return nil, err
	}
	return i.(*TreeRow), err
}

func (t *sqlTree) CreateNode(nodePath string, o interface{}, isLeaf bool) error {
	parent, name := path.Split(nodePath)
	var e error
	parentRow, e := t.getTreeByPath(parent)
	if e != nil {
		return e
	}

	encoded, err := t.codec.Encode(o)
	if err != nil {
		return err
	}

	err = t.Exec("insert_encoded", parentRow.id, name, encoded).Error
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
	pattern := src + "%"

	srcBase := path.Base(src)
	srcParent := path.Dir(src)
	srcParentInfo, e := t.getTreeByPath(srcParent)
	if e != nil {
		return e
	}

	dstBase := path.Base(dst)
	dstParent := path.Dir(dst)
	dstParentInfo, e := t.getTreeByPath(dstParent)
	if e != nil {
		return e
	}

	err := t.Exec("update_tree_path", src, dst, pattern).Error
	if err != nil {
		return err
	}

	err = t.Exec("update_name", dstBase, srcBase, srcParentInfo.id).Error
	if err != nil {
		return err
	}

	return t.Exec("update_parent", dstParentInfo.id, srcParentInfo.id, path.Base(src)).Error
}

func (t *sqlTree) RenameNode(src string, newName string) error {
	parent := path.Dir(src)
	newPath := path.Join(parent, newName)
	return t.MoveNode(src, newPath)
}

func (t *sqlTree) DeleteNode(nodePath string) error {
	name := path.Base(nodePath)
	parent := path.Dir(nodePath)

	nodeInfo, err := t.getTreeByPath(nodePath)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}

	isLeaf := nodeInfo == nil

	parentInfo, e := t.getTreeByPath(parent)
	if e != nil {
		return e
	}

	if err := t.Exec("delete_encoded", parentInfo.id, name).Error; err != nil {
		return err
	}

	if !isLeaf {
		return t.Exec("delete_tree_path", nodePath+"/%").Error
	}
	return nil
}

func (t *sqlTree) UpdateNode(nodePath string, o interface{}) error {
	parentPath := path.Dir(nodePath)

	parentInfo, e := t.getTreeByPath(parentPath)
	if e != nil {
		return e
	}

	encoded, err := t.codec.Encode(o)
	if err != nil {
		return err
	}

	return t.Exec("update_encoded", encoded, parentInfo.id, parentPath).Error
}

func (t *sqlTree) ReadNode(nodePath string, o interface{}) error {
	name := path.Base(nodePath)
	parent := path.Dir(nodePath)

	parentInfo, e := t.getTreeByPath(parent)
	if e != nil {
		return e
	}

	item, err := t.QueryOne("select_encoded", "encoded", parentInfo.id, name)
	if err != nil {
		return err
	}

	data := []byte(item.(string))
	return t.codec.Decode(data, o)
}

func (t *sqlTree) Children(dir string) (Cursor, error) {
	dirInfo, e := t.getTreeByPath(dir)
	if e != nil {
		return nil, e
	}

	c, err := t.Query("select_all_encoded", "encoded", dirInfo.id)
	if err != nil {
		return nil, err
	}

	return newCursor(c, t.codec), nil
}

func (t *sqlTree) deleteEncoded(nodePath string) error {
	name := path.Base(nodePath)
	parent := path.Dir(nodePath)

	parentInfo, err := t.getTreeByPath(parent)
	if err != nil {
		return err
	}

	return t.Exec("delete_encoded", parentInfo.id, name).Error
}
