package main

import "testing"
import "os"

func TestPaserLog(t *testing.T) {
	path := "./config.json"
	s := `
[
	{
		"LoggingPath": "/aaa",
		"LoggingName": "a",
		"FlushTime": 2.0,
		"BufferSize": 10000,
		"RotateType": "H"
	},
	{
		"LoggingPath": "/bbb",
		"LoggingName": "b"
	}
]
	`
	f, err := os.Create(path)
	f.Write([]byte(s))
	f.Close()
	f = nil
	defer os.Remove(path)

	a, err := LoadJSONConfig(path)
	if err != nil {
		t.Fatalf("err %v", err.Error())
	}

	if len(a) != 2 {
		t.Fatalf("err length of a : %d", len(a))
	}

	a1 := a[0]
	if a1.LoggingName != "a" ||
		a1.LoggingPath != "/aaa" ||
		a1.FlushTime != 2.0 ||
		a1.BufferSize != 10000 ||
		a1.RotateType != "H" {
		t.Fatalf("Unexpect item %v", a1)
	}

}
