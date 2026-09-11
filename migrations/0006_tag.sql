-- 标签：tag 表 + article_tag 多对多关联表
-- slug 用于 URL（/tags/:slug）；name 用于展示。中文标签的 slug 直接用原名（URL 里会 percent-encode，可正常用）
CREATE TABLE IF NOT EXISTS tag (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT UNIQUE NOT NULL,
    slug       TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS article_tag (
    article_id INTEGER NOT NULL REFERENCES article (id),
    tag_id     INTEGER NOT NULL REFERENCES tag (id),
    PRIMARY KEY (article_id, tag_id)
);

-- 按标签筛文章：从 tag_id 反查关联记录
CREATE INDEX IF NOT EXISTS idx_article_tag_tag ON article_tag (tag_id);
