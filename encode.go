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

type basicEncoder struct {
	delimiter rune
	tagKey    string
	lineBreak string
	written   bool
	w         io.Writer
	tags      []Tag
	names     []string
	tagStack  Tags
}

func (enc *basicEncoder) encodeLine(v interface{}) error {
	if err := enc.encodeValue(reflect.ValueOf(v)); err != nil {
		return err
	}
	if _, err := enc.w.Write([]byte(enc.lineBreak)); err != nil {
		return err
	}
	return nil
}

func (enc *basicEncoder) encodeValue(v reflect.Value) error {
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
			return enc.encodeValue(v.Elem())
		}
	default:
		return enc.write([]byte(fmt.Sprint(v.Interface())))
	}
	return nil
}

func (enc *basicEncoder) write(bs []byte) error {
	if enc.written {
		_, err := enc.w.Write([]byte(string(enc.delimiter)))
		if err != nil {
			return err
		}
	}
	enc.tags = append(enc.tags, enc.tagStack.top())
	enc.names = append(enc.names, enc.tagStack.join(enc.tagKey))
	_, err := enc.w.Write(bs)
	if err != nil {
		return err
	}
	enc.written = true
	return nil
}

func (enc *basicEncoder) encodeStruct(v reflect.Value) error {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := Tag(field.Tag)
		if !tag.Has(enc.tagKey) && (field.Type.Kind() != reflect.Struct || field.Type == reflect.TypeOf(time.Time{})) {
			continue
		}
		if tag.Get(enc.tagKey) == "-" {
			continue
		}
		enc.tagStack.push(tag)
		if err := enc.encodeValue(v.Field(i)); err != nil {
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
