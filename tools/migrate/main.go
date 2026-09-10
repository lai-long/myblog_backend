// migrate 是一个极简的 SQLite 迁移工具，按文件名顺序幂等执行内嵌的迁移 SQL。
// 应用启动时也会自动执行同一套迁移（见 blog.go），本工具仅用于手工操作。
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"myblog_backend/migrations"

	_ "modernc.org/sqlite"
)

func main() {
	// 默认 DSN 与应用一致：应用 WORKDIR 为 /app/blog，../data -> /app/data
	dsn := flag.String("dsn", "../data/blog.db", "SQLite DSN（需与应用一致）")
	flag.Parse()

	db, err := sql.Open("sqlite", *dsn)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer db.Close()

	if err := migrations.Apply(db); err != nil {
		fatal("%v", err)
	}
	fmt.Println("migrations done")
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
