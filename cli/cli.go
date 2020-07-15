package cli

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/alecthomas/kong"
	"github.com/tkuchiki/mysql-parser/parser"
	"github.com/tkuchiki/sql2struct/db"
)

type Cli struct {
	out    io.Writer
	in     io.Reader
	DbUser string `help:"database uesr" required`
	DbPass string `help:"database password"`
	DbHost string `help:"database host" default:"localhost"`
	DbPort int    `help:"database port" default:"3306"`
	DbSock string `help:"database socket"`
	DbName string `help:"database name"`
	Sql    string `help:"sql"`
}

func New(out io.Writer, in io.Reader) *Cli {
	return &Cli{
		out: out,
		in:  in,
	}
}

func (c *Cli) Run() error {
	kong.Parse(c)

	sql := c.Sql
	if sql == "" {
		b, err := ioutil.ReadAll(c.in)
		if err != nil {
			return err
		}
		sql = string(b)
	}

	p := parser.New()
	err := p.Parse(sql)
	if err != nil {
		return err
	}

	client, err := db.New(c.DbUser, c.DbPass, c.DbHost, c.DbName, c.DbSock, c.DbPort)
	if err != nil {
		return err
	}
	defer client.Close()

	q := p.Query()
	columns := make([]db.Columns, 0)
	for _, table := range q.Table.GetNames() {
		cols, err := client.TableDefinitions(table, q.Table.Columns[table])
		columns = append(columns, cols...)

		if err != nil {
			return err
		}
	}

	st, err := client.GenStruct(q)
	fmt.Fprintln(c.out, st)

	return nil
}
