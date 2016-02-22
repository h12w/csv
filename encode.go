package csv

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"time"
)

const (
	sqlTime = "2006-01-02 15:04:05.000"
)

type Encoder struct {
	Delimiter rune
	TagKey    string
	LineBreak string
	written   bool
	w         io.Writer
	tags      []Tag
	names     []string
	tagStack  Tags
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		Delimiter: ',',
		TagKey:    "csv",
		LineBreak: "\n",
		w:         w,
	}
}

func (enc *Encoder) SetDelimeter(delim rune) *Encoder       { enc.Delimiter = delim; return enc }
func (enc *Encoder) SetTagKey(tagKey string) *Encoder       { enc.TagKey = tagKey; return enc }
func (enc *Encoder) SetLineBreak(lineBreak string) *Encoder { enc.LineBreak = lineBreak; return enc }

func (enc *Encoder) Encode(v interface{}) error {
	if err := enc.encode(reflect.ValueOf(v)); err != nil {
		return err
	}
	if _, err := enc.w.Write([]byte(enc.LineBreak)); err != nil {
		return err
	}
	return nil
}

func (enc *Encoder) Tags() []Tag     { return enc.tags }
func (enc *Encoder) Names() []string { return enc.names }

func (enc *Encoder) encode(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Struct:
		if t, ok := v.Interface().(time.Time); ok {
			return enc.write([]byte(t.Format(sqlTime)))
		}
		return enc.encodeStruct(v)
	case reflect.Slice:
		return nil
	case reflect.Ptr:
		if !v.IsNil() {
			return enc.encode(v.Elem())
		}
	default:
		return enc.write([]byte(fmt.Sprint(v.Interface())))
	}
	return nil
}

func (enc *Encoder) write(bs []byte) error {
	if enc.written {
		_, err := enc.w.Write([]byte(string(enc.Delimiter)))
		if err != nil {
			return err
		}
	}
	enc.tags = append(enc.tags, enc.tagStack.top())
	enc.names = append(enc.names, enc.tagStack.join(enc.TagKey))
	_, err := enc.w.Write(bs)
	if err != nil {
		return err
	}
	enc.written = true
	return nil
}

func (enc *Encoder) encodeStruct(v reflect.Value) error {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := Tag(field.Tag)
		if !tag.Has(enc.TagKey) && (field.Type.Kind() != reflect.Struct || field.Type == reflect.TypeOf(time.Time{})) {
			continue
		}
		if tag.Get(enc.TagKey) == "-" {
			continue
		}
		enc.tagStack.push(tag)
		if err := enc.encode(v.Field(i)); err != nil {
			return err
		}
		enc.tagStack.pop()
	}
	return nil
}

func Marshal(v interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	if err := NewEncoder(w).Encode(v); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
