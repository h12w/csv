package csv

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type Expander struct {
	Delimiter rune
	Tag       string
	LineBreak string
	w         io.Writer
}

func NewExpander(w io.Writer) *Expander {
	return &Expander{
		Delimiter: ',',
		Tag:       "csv",
		LineBreak: "\n",
		w:         w,
	}
}

func (e *Expander) SetDelimeter(delim rune) *Expander       { e.Delimiter = delim; return e }
func (e *Expander) SetTag(tag string) *Expander             { e.Tag = tag; return e }
func (e *Expander) SetLineBreak(lineBreak string) *Expander { e.LineBreak = lineBreak; return e }

func (e *Expander) Expand(value interface{}, path ...string) error {
	ps := make([][]string, len(path))
	for i := range ps {
		ps[i] = strings.Split(path[i], ".")
	}
	v := reflect.ValueOf(value)
	fields, err := e.marshal(v)
	if err != nil {
		return err
	}
	slice, err := getSlice(v, ps[0])
	if err != nil {
		return err
	}
	its := its{{
		slice:  slice,
		prefix: fields,
	}}
	for {
		it := its.top()
		v, ok := it.next()
		if !ok {
			its.pop()
			break
		}
		fields, err := e.marshal(v)
		if err != nil {
			return err
		}
		if it.level+1 < len(ps) {
			slice, err := getSlice(v, ps[it.level+1])
			if err != nil {
				return err
			}
			its.push(&iter{
				slice:  slice,
				level:  it.level + 1,
				prefix: fields,
			})
			continue
		}
		if _, err := e.w.Write(bytes.Join(append(its.prefixes(), fields), []byte(string(e.Delimiter)))); err != nil {
			return err
		}
		if _, err := e.w.Write([]byte(e.LineBreak)); err != nil {
			return err
		}
	}
	return nil
}

func (e *Expander) marshal(v reflect.Value) ([]byte, error) {
	w := new(bytes.Buffer)
	enc := Encoder{w: w, Delimiter: e.Delimiter, Tag: e.Tag}
	if err := enc.encode(v); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

type its []*iter

func (s *its) top() *iter {
	return (*s)[len(*s)-1]
}

func (s *its) push(it *iter) {
	*s = append(*s, it)
}

func (s *its) pop() (res *iter) {
	*s, res = (*s)[:len(*s)-1], (*s)[len(*s)-1]
	return
}

func (s *its) prefixes() [][]byte {
	bufs := make([][]byte, len(*s))
	for i := range *s {
		bufs[i] = (*s)[i].prefix
	}
	return bufs
}

type iter struct {
	slice  reflect.Value
	i      int
	level  int
	prefix []byte
}

func (it *iter) next() (reflect.Value, bool) {
	v := reflect.New(it.slice.Type().Elem()).Elem()
	if it.i >= it.slice.Len() {
		return reflect.Value{}, false
	}
	v.Set(it.slice.Index(it.i))
	it.i++
	return v, it.i <= it.slice.Len()
}

func getSlice(v reflect.Value, path []string) (reflect.Value, error) {
	for _, name := range path {
		if v.Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("expect struct type in path %v", path)
		}
		v = v.FieldByName(name)
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
	default:
		return reflect.Value{}, fmt.Errorf("expect slice or array type for path %v", path)
	}
	return v, nil
}
