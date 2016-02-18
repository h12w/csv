package csv

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
)

const (
	csvTag = "csv"
	delim  = ','
)

type Encoder struct {
	written bool
	w       io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (enc *Encoder) Encode(v interface{}) error {
	return enc.encode(reflect.ValueOf(v))
}

func (enc *Encoder) encode(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Struct:
		return enc.encodeStruct(v)
	default:
		if enc.written {
			_, err := enc.w.Write([]byte{delim})
			if err != nil {
				return err
			}
		}
		_, err := fmt.Fprint(enc.w, v.Interface())
		if err != nil {
			return err
		}
		enc.written = true
		return nil
	}
	return nil
}

func (enc *Encoder) encodeStruct(v reflect.Value) error {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if t.Field(i).Tag.Get(csvTag) == "-" {
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
