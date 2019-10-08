package persist

import (
	"github.com/zoenion/common/dao"
)

type Triplet struct {
	First  string
	Second string
	Third  string
}

type Pair struct {
	First  string
	Second string
}

func scanData(row dao.Row) (interface{}, error) {
	var data string
	return data, row.Scan(&data)
}

func scanPair(row dao.Row) (interface{}, error) {
	p := new(Pair)
	return p, row.Scan(&p.First, &p.Second)
}

func scanTriplet(row dao.Row) (interface{}, error) {
	t := new(Triplet)
	return t, row.Scan(&t.First, &t.Second, &t.Third)
}
