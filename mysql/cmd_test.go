package mysql

import (
	"testing"
)

func TestCmd(t *testing.T) {
	type S struct {
		ID     string `table:"id" TYPE:"VARCHAR(10)" PK:"yes"`
		V1     int    `table:"v1" TYPE:"INT(11)"`
		Nested struct {
			V2 float64 `table:"v2" TYPE:"FLOAT"`
		}
	}
	cmd := Cmd{
		Value:  S{},
		TagKey: "table",
		DB:     "db1",
		Engine: "InnoDB",
		Table:  "db1.table",
	}
	{
		createTable, err := cmd.CreateTable()
		if err != nil {
			t.Fatal(err)
		}
		expected := `CREATE TABLE IF NOT EXISTS db1.table (
	id VARCHAR(10),
	v1 INT(11),
	v2 FLOAT,
	PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=UTF8;
`
		if createTable != expected {
			t.Fatalf("expect\n%s\ngot\n%s", expected, createTable)
		}
	}
	{
		loadData, err := cmd.LoadData()
		if err != nil {
			t.Fatal(err)
		}
		expected := "LOAD DATA LOCAL INFILE 'Reader::%[1]s' REPLACE INTO TABLE %[1]s (id, v1, v2);\n"
		if loadData != expected {
			t.Fatalf("expect\n%s\ngot\n%s", expected, loadData)
		}
	}
	{
		createDB := cmd.CreateDB()
		expected := "CREATE DATABASE IF NOT EXISTS db1 DEFAULT CHARACTER SET utf8;\n"
		if createDB != expected {
			t.Fatalf("expect\n%s\ngot\n%s", expected, createDB)
		}
	}
}
