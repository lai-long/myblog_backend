-- 评论表
-- status: 0=待审核 1=通过 2=垃圾
-- parent_id: 0 表示顶层评论，非 0 为回复的父评论 id（用 0 不用 NULL：INTEGER 列存 NULL，Scan 到 int64 会报错）
CREATE TABLE IF NOT EXISTS comment (
                                     id         INTEGER PRIMARY KEY AUTOINCREMENT,
                                     article_id INTEGER NOT NULL,
                                     parent_id  INTEGER NOT NULL DEFAULT 0,
                                     nickname   TEXT NOT NULL,
                                     avatar_url TEXT NOT NULL DEFAULT '',
                                     content    TEXT NOT NULL,
                                     status     INTEGER NOT NULL DEFAULT 0,
                                     created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                                     deleted_at DATETIME
);

-- 公开列表：按文章查已通过的，按时间正序
CREATE INDEX IF NOT EXISTS idx_comment_article
    ON comment (article_id, status, created_at);
