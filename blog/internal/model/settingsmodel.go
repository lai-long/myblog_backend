package model

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// SettingsModel 站点配置 KV 表（settings）。一共没几行数据，不值得动用 goctl 生成，手写两个小方法
type SettingsModel struct {
	conn sqlx.SqlConn
}

func NewSettingsModel(conn sqlx.SqlConn) *SettingsModel {
	return &SettingsModel{conn: conn}
}

// Get 读单个 key；不存在返回空字符串而不是报错（配置项允许没配过）
func (m *SettingsModel) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := m.conn.QueryRowCtx(ctx, &value, "SELECT `value` FROM `settings` WHERE `key` = ?", key)
	if errors.Is(err, sqlx.ErrNotFound) {
		return "", nil
	}
	return value, err
}

// Set 写入：已存在则覆盖（SQLite 的 upsert 语法）
func (m *SettingsModel) Set(ctx context.Context, key, value string) error {
	_, err := m.conn.ExecCtx(ctx,
		"INSERT INTO `settings` (`key`, `value`, `updated_at`) VALUES (?, ?, CURRENT_TIMESTAMP) "+
			"ON CONFLICT(`key`) DO UPDATE SET `value` = excluded.`value`, `updated_at` = excluded.`updated_at`",
		key, value)
	return err
}
