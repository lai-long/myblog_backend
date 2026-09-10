-- 文章软删除：deleted_at 为 NULL 表示未删除
-- 用 NULL 而不是 '' —— 空字符串不是合法时间值，Scan 到 time.Time 会报错
ALTER TABLE article ADD COLUMN deleted_at DATETIME;
