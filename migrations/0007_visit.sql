-- 访问统计：每次页面浏览一行（PV），UV 按 天+IP哈希 去重。
-- 存 IP 的 SHA-256 哈希而不是原始 IP，兼顾去重与访客隐私。
CREATE TABLE IF NOT EXISTS visit (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    date       TEXT NOT NULL,        -- YYYY-MM-DD（本地时区），按天统计的分桶键
    ip_hash    TEXT NOT NULL,        -- sha256(盐 + 客户端 IP)
    path       TEXT NOT NULL,        -- 访问路径，如 /articles/hello-world
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 统计查询都按 date 过滤/分组
CREATE INDEX IF NOT EXISTS idx_visit_date ON visit (date);
