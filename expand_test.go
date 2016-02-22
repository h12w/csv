package csv

import (
	"bytes"
	"testing"
)

func TestMarshalSlice(t *testing.T) {
	type (
		Leaf struct {
			V1 int `csv2:"v1"`
			V2 int `csv2:"v2"`
		}
		Level2 struct {
			V3   int `csv2:"v3"`
			Leaf []Leaf
			V4   int `csv2:"v4"`
		}
		Level1 struct {
			Level2 []Level2
		}
		Struct struct {
			V1     string `csv:"v1" csv2:"v1"`
			Level1 Level1 `csv:"-"`
			V3     string `csv:"v3" csv2:"v3"`
			V4     string `csv:"v4"`
		}
	)
	st := Struct{
		V1: "a",
		Level1: Level1{[]Level2{
			{1, []Leaf{{3, 4}, {5, 6}}, 2},
		}},
		V3: "b",
		V4: "c",
	}

	w := new(bytes.Buffer)
	expander := NewEncoder(w).SetTagKey("csv2").SetExpandPath("Level1.Level2", "Leaf")
	if err := expander.Encode(st); err != nil {
		t.Fatal(err)
	}
	actual := w.String()

	expected := "a,b,1,2,3,4\na,b,1,2,5,6\n"
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}

}
