package csv

import (
	"bytes"
	"io"
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

func newIter(v reflect.Value, path [][]string, level int) *iter {
	return &iter{
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

func unmarshal(v reflect.Value, delimiter rune, tag string) ([]byte, error) {
	w := new(bytes.Buffer)
	enc := Encoder{w: w, Delimiter: ',', Tag: "csv2"}
	if err := enc.encode(v); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
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
	st := Struct{
		V1: "a",
		Level1: Level1{[]Level2{
			{1, []Leaf{{3, 4}, {5, 6}}, 2},
		}},
		V3: "b",
		V4: "c",
	}

	path := [][]string{[]string{"Level1", "Level2"}, []string{"Leaf"}}
	w := new(bytes.Buffer)
	if err := expand(w, reflect.ValueOf(st), path); err != nil {
		t.Fatal(err)
	}
	actual := w.String()

	expected := "a,b,1,2,3,4\na,b,1,2,5,6\n"
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}

}

func expand(w io.Writer, v reflect.Value, path [][]string) error {
	var buf [][]byte

	fields, err := unmarshal(v, ',', "csv2")
	if err != nil {
		return err
	}
	buf = append(buf, fields)
	its := []*iter{newIter(v, path, 0)}
	for {
		it := its[len(its)-1]
		v = it.New()
		if !it.Next(v) {
			break
		}
		fields, err := unmarshal(v, ',', "csv2")
		if err != nil {
			return err
		}
		buf = append(buf, fields)
		its = append(its, newIter(v, path, it.level+1))
		for {
			it := its[len(its)-1]
			v := it.New()
			if !it.Next(v) {
				break
			}
			fields, err := unmarshal(v, ',', "csv2")
			if err != nil {
				return err
			}
			buf = append(buf, fields)
			if it.level+1 == len(path) {
				if _, err := w.Write(append(bytes.Join(buf, []byte{','}), '\n')); err != nil {
					return err
				}
			}
			buf = buf[:len(buf)-1]
		}
		buf = buf[:len(buf)-1]
		its = its[:len(its)-1]
	}
	buf = buf[:len(buf)-1]
	its = its[:len(its)-1]
	return nil
}

type Types struct {
	V1 string    `csv:"v1"`
	V2 int       `csv:"v2"`
	V3 float64   `csv:"v3"`
	V4 time.Time `csv:"v4"`
}
