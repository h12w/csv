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
	Tag       string
	LineBreak string
	written   bool
	w         io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		Delimiter: ',',
		Tag:       "csv",
		LineBreak: "\n",
		w:         w,
	}
}

func (enc *Encoder) SetDelimeter(delim rune) *Encoder       { enc.Delimiter = delim; return enc }
func (enc *Encoder) SetTag(tag string) *Encoder             { enc.Tag = tag; return enc }
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
		tag := field.Tag.Get(enc.Tag)
		if tag == "-" || (tag == "" && field.Type.Kind() != reflect.Struct) {
			continue
		}
		if err := enc.encode(v.Field(i)); err != nil {
			return err
		}
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
