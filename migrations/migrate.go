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

// splitStmts 按分号切分，但触发器 BEGIN...END 块内部的分号不切（块里是多条语句）。
// 规则：逐段扫描，BEGIN 计数大于 END 计数时说明还在块内，继续累积。
func splitStmts(text string) []string {
	out := make([]string, 0)
	depth := 0
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] != ';' {
			continue
		}
		// 数一下 start..i 之间新出现的 BEGIN / END（按整词匹配，注释里不会有）
		depth += countWord(text[start:i], "BEGIN") - countWord(text[start:i], "END")
		if depth > 0 {
			continue // 在触发器块内，分号不切
		}
		if s := strings.TrimSpace(text[start:i]); s != "" {
			out = append(out, s)
		}
		start = i + 1
	}
	if s := strings.TrimSpace(text[start:]); s != "" {
		out = append(out, s)
	}
	return out
}

// countWord 统计整词出现次数（大小写不敏感）
func countWord(s, word string) int {
	n := 0
	for _, tok := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '(' || r == ')'
	}) {
		if strings.EqualFold(tok, word) {
			n++
		}
	}
	return n
}
