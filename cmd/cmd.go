package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"h12.me/csv"
)

type MySQLCmd struct {
	Value      interface{}
	TagKey     string
	ExpandPath []string
	Engine     string
	Replace    bool
}

func (cmd MySQLCmd) CreateDB(name string) string {
	return fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s DEFAULT CHARACTER SET utf8;\n", name)
}

func (cmd MySQLCmd) CreateTable(fullTableName string) (string, error) {
	w := new(bytes.Buffer)
	fmt.Fprintf(w, "CREATE TABLE IF NOT EXISTS %s (\n", fullTableName)
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

func (cmd MySQLCmd) LoadDataTemplate() (string, error) {
	fields, err := cmd.Fields()
	if err != nil {
		return "", err
	}

	replaceOrIgnore := "IGNORE"
	if cmd.Replace {
		replaceOrIgnore = "REPLACE"
	}

	return fmt.Sprintf("LOAD DATA LOCAL INFILE 'Reader::%%s' "+
		"%s INTO TABLE %%s "+
		"CHARACTER SET UTF8 "+
		`FIELDS OPTIONALLY ENCLOSED BY '"' `+
		"(%s);\n", replaceOrIgnore, strings.Join(fields.Names(), ", ")), nil
}

func (cmd MySQLCmd) Fields() (csv.Fields, error) {
	enc := csv.NewEncoder(ioutil.Discard).SetTagKey(cmd.TagKey).SetExpandPath(cmd.ExpandPath...)
	if err := enc.Encode(cmd.Value); err != nil {
		return nil, err
	}
	return enc.Fields(), nil
}
