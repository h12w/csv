package csv

import (
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

func TestMarshalNoTime(t *testing.T) {
	type S struct {
		I int `csv:"i"`
		T time.Time
	}
	var v S
	v.I = 1
	buf, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	expected := "1\n"
	actual := string(buf)
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}
