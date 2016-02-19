package csv

import (
	"bytes"
	"io"
	"reflect"
)

func expand(w io.Writer, v reflect.Value, path [][]string) error {
	fields, err := unmarshal(v, ',', "csv2")
	if err != nil {
		return err
	}
	its := []*iter{newIter(v, path, 0)}
	buf := [][]byte{fields}
	for {
		it := its[len(its)-1]
		v = it.New()
		if !it.Next(v) {
			buf = buf[:len(buf)-1]
			its = its[:len(its)-1]
			break
		}
		fields, err := unmarshal(v, ',', "csv2")
		if err != nil {
			return err
		}
		if it.level+1 < len(path) {
			its = append(its, newIter(v, path, it.level+1))
			buf = append(buf, fields)
		} else {
			buf = append(buf, fields)
			if _, err := w.Write(append(bytes.Join(buf, []byte{','}), '\n')); err != nil {
				return err
			}
			buf = buf[:len(buf)-1]
		}
	}
	return nil
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
