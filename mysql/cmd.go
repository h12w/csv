package mysql

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"h12.me/csv"
)

type Cmd struct {
	Value      interface{}
	TagKey     string
	ExpandPath []string
	DB         string
	Engine     string
	Table      string
}

func (cmd Cmd) CreateDB() string {
	return fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s DEFAULT CHARACTER SET utf8;\n", cmd.DB)
}

func (cmd Cmd) CreateTable() (string, error) {
	w := new(bytes.Buffer)
	fmt.Fprintf(w, "CREATE TABLE IF NOT EXISTS %s (\n", cmd.Table)
	fields, err := cmd.Fields()
	if err != nil {
		return "", err
	}
	var pks []string
	sort.Sort(fields)
	for _, field := range fields {
		if field.Tag.Get("PK") == "yes" {
			pks = append(pks, field.Name)
			fmt.Fprintf(w, "\t%s %s,\n", field.Name, field.Tag.Get("TYPE"))
		}
	}
	for _, field := range fields {
		if field.Tag.Get("PK") != "yes" {
			fmt.Fprintf(w, "\t%s %s,\n", field.Name, field.Tag.Get("TYPE"))
		}
	}
	fmt.Fprintf(w, "\tPRIMARY KEY (%s)\n", strings.Join(pks, ","))
	fmt.Fprintf(w, ") ENGINE=%s DEFAULT CHARSET=UTF8;\n", cmd.Engine)
	return w.String(), nil
}

func (cmd Cmd) LoadData() (string, error) {
	fields, err := cmd.Fields()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("LOAD DATA LOCAL INFILE 'Reader::%%[1]s' REPLACE INTO TABLE %%[1]s (%s);\n", strings.Join(fields.Names(), ", ")), nil
}

func (cmd Cmd) Fields() (csv.Fields, error) {
	enc := csv.NewEncoder(ioutil.Discard).SetTagKey(cmd.TagKey).SetExpandPath(cmd.ExpandPath...)
	if err := enc.Encode(cmd.Value); err != nil {
		return nil, err
	}
	return enc.Fields(), nil
}

func (cmd Cmd) FullTableName() string {
	return cmd.DB + "." + cmd.Table
}
