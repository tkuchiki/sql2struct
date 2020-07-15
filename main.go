package main

import (
	"log"
	"os"

	"github.com/tkuchiki/sql2struct/cli"
)

func main() {
	cli := cli.New(os.Stdout, os.Stdin)
	if err := cli.Run(); err != nil {
		log.Fatal(err)
	}
}
