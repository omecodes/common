package tree

type TreeRow struct {
	id       int64
	nodePath string
}

type EncodedRow struct {
	parent   int64
	nodeName string
	encoded  string
}
