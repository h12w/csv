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
			V5     string `csv:"v5" csv2:"v5"`
			Level1 Level1 `csv:"-"`
			V6     string `csv:"v6" csv2:"v6"`
			V7     string `csv:"v7"`
		}
	)
	st := Struct{
		V5: "a",
		Level1: Level1{[]Level2{
			{1, []Leaf{{3, 4}, {5, 6}}, 2},
		}},
		V6: "b",
		V7: "c",
	}

	{
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

	{
		w := new(bytes.Buffer)
		expander := NewEncoder(w).SetTagKey("csv2").SetExpandPath("Level1.Level2", "Leaf")
		if err := expander.EncodeHeader(st); err != nil {
			t.Fatal(err)
		}
		actual := w.String()

		expected := "v5,v6,v3,v4,v1,v2\n"
		if actual != expected {
			t.Fatalf("expected %s, got %s", expected, actual)
		}

	}

}
