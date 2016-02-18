package csv

import (
	"testing"
	"time"
)

type Nested struct {
	V1 string `csv:"v1"`
	A1 int    `csv:"-"`
	S1 struct {
		V2 string `csv:"v2"`
		S2 struct {
			V3 int `csv:"v3"`
		}
	}
}

type Types struct {
	V1 string    `csv:"v1"`
	V2 int       `csv:"v2`
	V3 float64   `csv:"v3"`
	V4 time.Time `csv:"v4"`
}

func TestMarshal(t *testing.T) {
	var v Nested
	v.V1 = "a"
	v.S1.V2 = "b"
	v.S1.S2.V3 = 1
	buf, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	expected := "a,b,1"
	actual := string(buf)
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}
