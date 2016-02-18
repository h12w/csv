package csv

import (
	"testing"
	"time"
)

func TestMarshalNested(t *testing.T) {
	type Nested struct {
		V1 string `csv:"v1"`
		A1 int    `csv:"-"`
		A2 int
		S1 struct {
			V2 string `csv:"v2"`
			S2 struct {
				V3 int `csv:"v3"`
			}
		}
	}

	var v Nested
	v.V1 = "a"
	v.S1.V2 = "b"
	v.S1.S2.V3 = 1
	buf, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	expected := "a,b,1\n"
	actual := string(buf)
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}

// TODO: use iterators!
func TestMarshalSlice(t *testing.T) {
	type Inner struct {
		V2 int `csv:"v2"`
	}
	type Struct struct {
		V1    string  `csv:"v1"`
		Slice []Inner `csv:",expand"`
		V3    string  `csv:"v3"`
	}
	v := Struct{V1: "a", Slice: []Inner{{1}, {2}}, V3: "b"}
	buf, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	expected := "a,1,b\na,2,b\n"
	actual := string(buf)
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}

type Types struct {
	V1 string    `csv:"v1"`
	V2 int       `csv:"v2"`
	V3 float64   `csv:"v3"`
	V4 time.Time `csv:"v4"`
}
