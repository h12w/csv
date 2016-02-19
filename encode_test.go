package csv

import (
	"bytes"
	"reflect"
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

type iter struct {
	a     reflect.Value
	i     int
	level int
}

func newIter(v reflect.Value, path [][]string, level int) iter {
	return iter{
		a:     findFieldByPath(v, path[level]),
		level: level,
	}
}

func (it *iter) New() reflect.Value {
	return reflect.New(it.a.Type().Elem()).Elem()
}

func (it *iter) Next(v reflect.Value) bool {
	if it.i >= it.a.Len() {
		return false
	}
	v.Set(it.a.Index(it.i))
	it.i++
	return it.i <= it.a.Len()
}

func findFieldByPath(v reflect.Value, path []string) reflect.Value {
	for _, name := range path {
		v = v.FieldByName(name)
	}
	return v
}

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
	v := Struct{
		V1: "a",
		Level1: Level1{[]Level2{
			{1, []Leaf{{3, 4}, {5, 6}}, 2},
		}},
		V3: "b",
		V4: "c",
	}

	path := [][]string{[]string{"Level1", "Level2"}, []string{"Leaf"}}

	w := new(bytes.Buffer)
	enc := Encoder{w: w, Delimiter: ',', Tag: "csv2"}
	if err := enc.Encode(v); err != nil {
		t.Fatal(err)
	}
	prefix1 := w.Bytes()

	actual := ""
	it := newIter(reflect.ValueOf(v), path, 0)
	level2 := it.New()
	for it.Next(level2) {
		w := new(bytes.Buffer)
		enc := Encoder{w: w, Delimiter: ',', Tag: "csv2"}
		if err := enc.encode(level2); err != nil {
			t.Fatal(err)
		}
		prefix2 := w.Bytes()

		it2 := newIter(level2, path, it.level+1)
		leaf := it2.New()
		for it2.Next(leaf) {
			w := new(bytes.Buffer)
			enc := Encoder{w: w, Delimiter: ',', Tag: "csv2"}
			if err := enc.encode(leaf); err != nil {
				t.Fatal(err)
			}
			actual += string(bytes.Join([][]byte{prefix1, prefix2, w.Bytes()}, []byte{','})) + "\n"
		}
	}

	expected := "a,b,1,2,3,4\na,b,1,2,5,6\n"
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
