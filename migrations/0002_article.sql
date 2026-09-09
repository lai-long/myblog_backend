-- 文章表
-- status: 0=草稿 1=发布 2=隐藏
-- 字符串列一律 NOT NULL DEFAULT ''，避免出现 NULL 导致 Scan 报错
CREATE TABLE IF NOT EXISTS article (
                                     id           INTEGER PRIMARY KEY AUTOINCREMENT,
                                     title        TEXT NOT NULL,
                                     slug         TEXT UNIQUE NOT NULL,
                                     summary      TEXT NOT NULL DEFAULT '',
                                     content      TEXT NOT NULL DEFAULT '',
                                     cover_url    TEXT NOT NULL DEFAULT '',
                                     status       INTEGER NOT NULL DEFAULT 0,
                                     views        INTEGER NOT NULL DEFAULT 0,
                                     published_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                     created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
                                     updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 公开列表：按 status 过滤 + published_at 倒序
CREATE INDEX IF NOT EXISTS idx_article_status_published
    ON article (status, published_at DESC);
