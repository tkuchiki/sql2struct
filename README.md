# sql2struct

## Installation

Download from https://github.com/tkuchiki/sql2struct/releases

## Usage

```console
$ sql2struct --help
Usage: sql2struct --db-user=STRING

Flags:
  -h, --help                   Show context-sensitive help.
      --db-user=STRING         database uesr
      --db-pass=STRING         database password
      --db-host="localhost"    database host
      --db-port=3306           database port
      --db-sock=STRING         database socket
      --db-name=STRING         database name
      --sql=STRING             sql
```

## Examples

```console
$ sql2struct --db-user=root --dbname=testdb --sql "SELECT t1.*, t2.name FROM t1 JOIN t1.id = t2.t1_id"

$ echo "SELECT t1.*, t2.name FROM t1 JOIN t1.id = t2.t1_id" | sql2struct --db-user=root --dbname=testdb
```
