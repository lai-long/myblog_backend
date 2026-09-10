package migrations

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// Apply 按文件名顺序幂等执行内嵌的迁移 SQL，已记录在 schema_migrations 的跳过。
// 应用启动时与 tools/migrate 共用同一套逻辑，行为一致。
func Apply(db *sql.DB) error {
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("pragma foreign_keys: %w", err)
	}

	// 记录已执行的迁移，保证可重复运行、不会因重复建表报错
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		name       TEXT NOT NULL UNIQUE,
		applied_at DATETIME NOT NULL
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := FS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		applied, err := alreadyApplied(db, name)
		if err != nil {
			return fmt.Errorf("check %s: %w", name, err)
		}
		if applied {
			fmt.Printf("skip  %s\n", name)
			continue
		}
		text, err := FS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		if err := applyOne(db, name, string(text)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
		fmt.Printf("apply %s\n", name)
	}
	return nil
}

func alreadyApplied(db *sql.DB, name string) (bool, error) {
	var cnt int
	err := db.QueryRow("SELECT COUNT(1) FROM schema_migrations WHERE name = ?", name).Scan(&cnt)
	return cnt > 0, err
}

func applyOne(db *sql.DB, name, text string) error {
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
