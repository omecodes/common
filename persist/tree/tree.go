package tree

import (
	"github.com/omecodes/common/codec"
	"github.com/omecodes/common/dao"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/jcon"
	"path"
)

type Tree interface {
	CreateRoot() error
	CreateNode(nodePath string, o interface{}, isLeaf bool) error
	MoveNode(nodePath string, newPath string) error
	RenameNode(nodePath string, newName string) error
	DeleteNode(nodePath string) error
	UpdateNode(nodePath string, o interface{}) error
	ReadNode(nodePath string, o interface{}) error
	Children(nodePath string) (Cursor, error)
	Leaves(nodePath string) (Cursor, error)
	LeavesRange(nodePath string, offset, count int) (Cursor, error)
	Subtrees(nodePath string) (Cursor, error)
	SubtreesRange(nodePath string, offset, count int) (Cursor, error)
	CountChildren(nodePath string) (int, error)
	CountLeaves(nodePath string) (int, error)
	CountSubtrees(nodePath string) (int, error)
	Clear() error
}

type sqlTree struct {
	dao.SQL
	codec codec.Codec
}

func NewSQL(dbCfg jcon.Map, prefix string, codec codec.Codec) (*sqlTree, error) {
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
				is_leaf int(1),
				foreign key (parent) references $prefix$_trees(id) on delete cascade
			);`).
		AddStatement("clear", `DELETE FROM $prefix$_trees;`).
		AddStatement("insert_tree_path", `INSERT INTO $prefix$_trees (path) VALUES(?);`).
		AddStatement("insert_encoded", `INSERT INTO $prefix$_encoded VALUES(?, ?, ?, ?);`).
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
		AddStatement("select_encoded", `SELECT * FROM $prefix$_encoded WHERE parent=? AND node_name=?;`).
		AddStatement("select_all_encoded", `SELECT * FROM $prefix$_encoded WHERE parent=? ORDER BY node_name;`).
		AddStatement("select_subtrees", `SELECT * FROM $prefix$_encoded WHERE parent=? AND is_leaf=0 ORDER BY node_name;`).
		AddStatement("select_subtrees_range", `SELECT * FROM $prefix$_encoded WHERE parent=? AND is_leaf=0 ORDER BY node_name limit ? offset ?;`).
		AddStatement("select_leaves", `SELECT * FROM $prefix$_encoded WHERE parent=? AND is_leaf=1  ORDER BY node_name limit ? offset ?;`).
		AddStatement("select_leaves_range", `SELECT * FROM $prefix$_encoded WHERE parent=? AND is_leaf=1  ORDER BY node_name;`).
		AddStatement("children_count", `SELECT count(*) FROM $prefix$_encoded WHERE parent=?;`).
		AddStatement("subtrees_count", `SELECT count(*) FROM $prefix$_encoded WHERE parent=? AND is_leaf=0;`).
		AddStatement("leaves_count", `SELECT count(*) FROM $prefix$_encoded WHERE parent=? AND is_leaf=1;`).
		RegisterScanner("tree", dao.NewScannerFunc(scanTreeRow)).
		RegisterScanner("count", dao.NewScannerFunc(func(row dao.Row) (interface{}, error) {
			var count int64
			return count, row.Scan(&count)
		})).
		RegisterScanner("encoded", dao.NewScannerFunc(scanEncodedRow))

	err := db.Init(dbCfg)
	if err != nil {
		return nil, err
	}

	_ = db.AddUniqueIndex(dao.SQLIndex{Name: "unique_parent_child", Table: "$prefix$_encoded", Fields: []string{
		"parent", "node_name",
	}}, false)
	_ = db.AddUniqueIndex(dao.SQLIndex{Name: "unique_parent_path", Table: "$prefix$_trees", Fields: []string{
		"path",
	}}, false)

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

func (t *sqlTree) CreateRoot() error {
	return t.init()
}

func (t *sqlTree) CreateNode(nodePath string, o interface{}, isLeaf bool) error {
	parent := path.Dir(nodePath)
	if parent == "" {
		parent = "/"
	}
	name := path.Base(nodePath)

	var e error
	parentRow, e := t.getTreeByPath(parent)
	if e != nil {
		return e
	}

	encoded, err := t.codec.Encode(o)
	if err != nil {
		return err
	}

	isLeafIntValue := 0
	if isLeaf {
		isLeafIntValue = 1
	}

	err = t.Exec("insert_encoded", parentRow.id, name, encoded, isLeafIntValue).Error
	if err != nil {
		return err
	}

	if !isLeaf && parent != "/" && parent != "" {
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

	encoded := item.(*EncodedRow)
	return t.codec.Decode([]byte(encoded.encoded), o)
}

func (t *sqlTree) Children(nodePath string) (Cursor, error) {
	dirInfo, e := t.getTreeByPath(nodePath)
	if e != nil {
		return nil, e
	}

	c, err := t.Query("select_all_encoded", "encoded", dirInfo.id)
	if err != nil {
		return nil, err
	}

	return newCursor(c, t.codec), nil
}

func (t *sqlTree) Leaves(nodePath string) (Cursor, error) {
	dirInfo, e := t.getTreeByPath(nodePath)
	if e != nil {
		return nil, e
	}

	c, err := t.Query("select_leaves", "encoded", dirInfo.id)
	if err != nil {
		return nil, err
	}

	return newCursor(c, t.codec), nil
}

func (t *sqlTree) LeavesRange(nodePath string, offset, count int) (Cursor, error) {
	if offset+count <= offset {
		return nil, errors.BadInput
	}

	dirInfo, e := t.getTreeByPath(nodePath)
	if e != nil {
		return nil, e
	}

	c, err := t.Query("select_leaves_range", "encoded", dirInfo.id, count, offset)
	if err != nil {
		return nil, err
	}

	return newCursor(c, t.codec), nil
}

func (t *sqlTree) Subtrees(nodePath string) (Cursor, error) {
	dirInfo, e := t.getTreeByPath(nodePath)
	if e != nil {
		return nil, e
	}

	c, err := t.Query("select_subtrees", "encoded", dirInfo.id)
	if err != nil {
		return nil, err
	}

	return newCursor(c, t.codec), nil
}

func (t *sqlTree) SubtreesRange(nodePath string, offset, count int) (Cursor, error) {
	if offset+count <= offset {
		return nil, errors.BadInput
	}

	dirInfo, e := t.getTreeByPath(nodePath)
	if e != nil {
		return nil, e
	}

	c, err := t.Query("select_subtrees_range", "encoded", dirInfo.id, count, offset)
	if err != nil {
		return nil, err
	}

	return newCursor(c, t.codec), nil
}

func (t *sqlTree) CountChildren(nodePath string) (int, error) {
	var e error
	parentRow, e := t.getTreeByPath(nodePath)
	if e != nil {
		return 0, e
	}

	o, err := t.QueryOne("children_count", "count", parentRow.id)
	if err != nil {
		return 0, err
	}

	return o.(int), nil
}

func (t *sqlTree) CountLeaves(nodePath string) (int, error) {
	var e error
	parentRow, e := t.getTreeByPath(nodePath)
	if e != nil {
		return 0, e
	}

	o, err := t.QueryOne("leaves_count", "count", parentRow.id)
	if err != nil {
		return 0, err
	}

	return o.(int), nil
}

func (t *sqlTree) CountSubtrees(nodePath string) (int, error) {
	var e error
	parentRow, e := t.getTreeByPath(nodePath)
	if e != nil {
		return 0, e
	}

	o, err := t.QueryOne("subtrees_count", "count", parentRow.id)
	if err != nil {
		return 0, err
	}

	return o.(int), nil
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
