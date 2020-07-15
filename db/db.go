package db

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/iancoleman/strcase"
	"github.com/tkuchiki/mysql-parser/query"
)

type Client struct {
	db      *sql.DB
	columns []string
}

func New(dbuser, dbpass, dbhost, dbname, socket string, port int) (*Client, error) {
	userpass := fmt.Sprintf("%s:%s", dbuser, dbpass)
	var conn string
	if socket != "" {
		conn = fmt.Sprintf("unix(%s)", socket)
	} else {
		conn = fmt.Sprintf("tcp(%s:%d)", dbhost, port)
	}

	s, err := sql.Open("mysql", fmt.Sprintf("%s@%s/%s", userpass, conn, dbname))
	if err != nil {
		return nil, err
	}

	columns := []string{
		"Field",
		"Type",
		"Collation",
		"Null",
		"Key",
		"Default",
		"Extra",
		"Privileges",
		"Comment",
	}

	return &Client{
		db:      s,
		columns: columns,
	}, nil
}

type Columns map[string]string

func NewColumns(col string) Columns {
	c := Columns{}
	c["Field"] = col

	return c
}

func (c Columns) Column() string {
	return strcase.ToCamel(c["Field"])
}

func (c Columns) Type() string {
	lowert := strings.ToLower(c["Type"])
	isUnsigned := strings.Index(lowert, "unsigned") != -1
	rep := strings.NewReplacer("unsigned", "", "signed", "", " ", "")
	t := rep.Replace(lowert)
	if t == "tinyint(1)" || strings.HasPrefix(t, "bool") {
		return "bool"
	} else if strings.Index(t, "int") != -1 {
		if isUnsigned {
			return "uint64"
		} else {
			return "int64"
		}
	} else if strings.HasPrefix(t, "double") || strings.HasPrefix(t, "float") || strings.HasPrefix(t, "decimal") {
		return "float64"
	} else if t == "" {
		return "interface{}"
	}

	return "string"
}

func (c Columns) StructTag() string {
	return fmt.Sprintf("`json:\"%s\" db:\"%s\"`", c["Field"], c["Field"])
}

func contains(s []string, v string) bool {
	for _, str := range s {
		if str == v {
			return true
		}
	}
	return false
}

func (c *Client) TableDefinitions(table string, cols []string) ([]Columns, error) {
	sql := fmt.Sprintf("SHOW FULL COLUMNS FROM `%s`", table)

	rows, err := c.db.Query(sql)
	if err != nil {
		return []Columns{}, err
	}

	isAllColumns := false
	if len(cols) == 1 && cols[0] == "*" {
		isAllColumns = true
	}

	_ = isAllColumns

	values := make([][]byte, len(c.columns))
	row := make([]interface{}, len(c.columns))
	for i, _ := range values {
		row[i] = &values[i]
	}

	data := make([]Columns, 0)

	for rows.Next() {
		if err := rows.Scan(row...); err != nil {
			return []Columns{}, err
		}

		r := make(Columns)
		for i, val := range values {
			v := string(val)
			r[c.columns[i]] = v
		}

		if !isAllColumns && !contains(cols, r["Field"]) {
			continue
		}

		data = append(data, r)
	}

	return data, nil
}

func (c *Client) GenStruct(q *query.Query) (string, error) {
	tableNames := q.Table.GetNames()
	columns := make([]Columns, 0)
	for _, table := range q.Table.GetNames() {
		cols, err := c.TableDefinitions(table, q.Table.Columns[table])
		columns = append(columns, cols...)

		if err != nil {
			return "", err
		}
	}

	aliases, ok := q.Table.Columns["*aliases_functions*"]
	if ok {
		for _, a := range aliases {
			col := NewColumns(a)
			columns = append(columns, col)
		}
	}

	var tableName string
	for _, t := range tableNames {
		tableName += strcase.ToCamel(strings.ToLower(t))
	}

	tmpl := `type {{ .Name }} struct {
{{- range $i, $col := .Columns }}
    {{ $col.Column }} {{ $col.Type }} {{ $col.StructTag }} 
{{- end }}
}`
	data := struct {
		Name    string
		Columns []Columns
	}{
		Name:    tableName,
		Columns: columns,
	}
	t, err := template.New("struct").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (c *Client) Close() {
	c.db.Close()
}
