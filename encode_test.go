package csv

import (
	"testing"
	"time"
)

func TestMarshalNested(t *testing.T) {
	type (
		S2Type *struct {
			V3 int `csv:"v3"`
		}
		Nested struct {
			V1 string `csv:"v1"`
			A1 int    `csv:"-"`
			A2 int
			S1 struct {
				V2 string `csv:"v2"`
				S2 S2Type `csv:""`
				S3 struct {
					V4 int `csv:"v4"`
				} `csv:""`
			} `csv:""`
		}
	)

	var v Nested
	v.V1 = "a"
	v.S1.V2 = "b"
	v.S1.S3.V4 = 1
	buf, err := Marshal(&v)
	if err != nil {
		t.Fatal(err)
	}
	expected := "a,b,0,1\n"
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
