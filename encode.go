package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

const (
	sqlTime = "2006-01-02 15:04:05.000"
)

type basicEncoder struct {
	delimiter rune
	tagKey    string
	lineBreak string
	w         io.Writer
	fields    Fields
	tagStack  Tags
}

func (enc *basicEncoder) encodeLine(v interface{}) error {
	enc.fields = nil
	if err := enc.encodeValue(reflect.ValueOf(v)); err != nil {
		return err
	}
	return enc.encodeFields(enc.fields.Values())
}

func (enc *basicEncoder) encodeFields(values []string) error {
	for i := range values {
		values[i] = escapeField(values[i])
	}
	w := csv.NewWriter(enc.w)
	w.Comma = enc.delimiter
	if enc.lineBreak == "\r\n" {
		w.UseCRLF = true
	}
	if err := w.Write(values); err != nil {
		return err
	}
	w.Flush()
	return nil
}
func escapeField(s string) string {
	s = strings.Replace(s, "\t", `\t`, -1)
	s = strings.Replace(s, "\n", `\n`, -1)
	return s
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
	case reflect.Bool:
		if v.Bool() {
			return enc.write([]byte{'1'})
		} else {
			return enc.write([]byte{'0'})
		}
	default:
		return enc.write([]byte(fmt.Sprint(v.Interface())))
	}
	return nil
}

func (enc *basicEncoder) write(bs []byte) error {
	enc.fields = append(enc.fields, Field{
		Name:  enc.tagStack.join(enc.tagKey),
		Value: string(bs),
		Tag:   enc.tagStack.top()})
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
