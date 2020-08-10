package jcon

import (
	"log"
	"testing"
)

func TestGet(t *testing.T) {
	var data = map[string]interface{}{
		"names": map[string]interface{}{
			"first": "jabar",
			"last":  "Oman",
			"other": "Ballack",
		},
	}

	cfg := Map(data)
	first, ok := cfg.GetString("names/first")
	log.Println("first => ", first, "ok => ", ok)
}
