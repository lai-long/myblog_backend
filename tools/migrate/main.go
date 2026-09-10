// migrate 是一个极简的 SQLite 迁移工具，按文件名顺序幂等执行 migrations/*.sql。
// 它会被打进镜像，在容器启动时先于 blog 运行，保证空 volume 也能自动建表。
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

func main() {
	dir := flag.String("dir", "migrations", "迁移 SQL 所在目录")
	// 默认 DSN 与应用一致：应用 WORKDIR 为 /app/blog，../data -> /app/data
	dsn := flag.String("dsn", "../data/blog.db", "SQLite DSN（需与应用一致）")
	flag.Parse()

	db, err := sql.Open("sqlite", *dsn)
	if err != nil {
		fatal("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		fatal("pragma foreign_keys: %v", err)
	}

	// 记录已执行的迁移，保证可重复运行、不会因重复建表报错
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		name       TEXT NOT NULL UNIQUE,
		applied_at DATETIME NOT NULL
	)`); err != nil {
		fatal("create schema_migrations: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(*dir, "*.sql"))
	if err != nil {
		fatal("glob migrations: %v", err)
	}
	if len(files) == 0 {
		fmt.Printf("warn: no .sql found in %s\n", *dir)
	}

	for _, f := range files {
		name := filepath.Base(f)
		done, err := alreadyApplied(db, name)
		if err != nil {
			fatal("check %s: %v", name, err)
		}
		if done {
			fmt.Printf("skip  %s\n", name)
			continue
		}
		sqlText, err := os.ReadFile(f)
		if err != nil {
			fatal("read %s: %v", name, err)
		}
		if err := apply(db, name, string(sqlText)); err != nil {
			fatal("apply %s: %v", name, err)
		}
		fmt.Printf("apply %s\n", name)
	}
	fmt.Println("migrations done")
}

func alreadyApplied(db *sql.DB, name string) (bool, error) {
	var cnt int
	err := db.QueryRow("SELECT COUNT(1) FROM schema_migrations WHERE name = ?", name).Scan(&cnt)
	return cnt > 0, err
}

func apply(db *sql.DB, name, text string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, stmt := range splitStmts(text) {
		if _, err := tx.Exec(stmt); err != nil {
			tx.Rollback()
			return fmt.Errorf("stmt [%s]: %w", stmt, err)
		}
	}
	if _, err := tx.Exec("INSERT INTO schema_migrations(name, applied_at) VALUES(?, datetime('now'))", name); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// splitStmts 按分号切分。迁移文件都是简单 DDL，字符串里不含分号，按 ; 切足够。
func splitStmts(text string) []string {
	out := make([]string, 0)
	for _, p := range strings.Split(text, ";") {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
