package csv

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type Encoder struct {
	basicEncoder
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		basicEncoder: basicEncoder{
			Delimiter: ',',
			TagKey:    "csv",
			LineBreak: "\n",
			w:         w,
		},
	}
}

func (e *Encoder) SetDelimeter(delim rune) *Encoder       { e.Delimiter = delim; return e }
func (e *Encoder) SetTagKey(tagKey string) *Encoder       { e.TagKey = tagKey; return e }
func (e *Encoder) SetLineBreak(lineBreak string) *Encoder { e.LineBreak = lineBreak; return e }

func (e *Encoder) Encode(value interface{}, path ...string) error {
	if len(path) == 0 {
		return e.basicEncoder.encodeLine(value)
	}
	ps := make([][]string, len(path))
	for i := range ps {
		ps[i] = strings.Split(path[i], ".")
	}
	v := reflect.ValueOf(value)
	return e.expand(v, ps)
}

func (e *Encoder) Tags() []Tag     { return e.tags }
func (e *Encoder) Names() []string { return e.names }

func (e *Encoder) expand(v reflect.Value, path [][]string) error {
	fields, err := e.marshal(v)
	if err != nil {
		return err
	}
	slice, err := getSlice(v, path[0])
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
		if it.level+1 < len(path) {
			slice, err := getSlice(v, path[it.level+1])
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
		e.written = true
		if _, err := e.w.Write([]byte(e.LineBreak)); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) marshal(v reflect.Value) ([]byte, error) {
	w := new(bytes.Buffer)
	enc := basicEncoder{w: w, Delimiter: e.Delimiter, TagKey: e.TagKey}
	if err := enc.encodeValue(v); err != nil {
		return nil, err
	}
	if !e.written { // only accumulate names for the first record
		e.tags = append(e.tags, enc.tags...)
		e.names = append(e.names, enc.names...)
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
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("expect struct type but got %v in path %v", v.Kind(), path)
		}
		v = v.FieldByName(name)
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
	default:
		return reflect.Value{}, fmt.Errorf("expect slice or array type but got %v for path %v", v.Kind(), path)
	}
	return v, nil
}
