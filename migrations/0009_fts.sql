-- 全文搜索：FTS5 索引表 + 触发器自动同步。
-- trigram 分词器：按连续 3 字符切片建索引，中文没有空格分词也能做子串匹配
-- （查询词不足 3 个字符时走 LIKE 兜底，见 articlemodel.go）。
CREATE VIRTUAL TABLE IF NOT EXISTS article_fts USING fts5(
    title, summary, content,
    tokenize = 'trigram'
);

-- 存量文章回填索引
INSERT INTO article_fts (rowid, title, summary, content)
SELECT id, title, summary, content FROM article WHERE deleted_at IS NULL;

-- 之后增/改/删文章由触发器同步（软删是 UPDATE，会走 update 触发器；
-- FTS 里的残留没关系，主查询有 deleted_at IS NULL 过滤）。
-- 全内容 FTS5 表直接用普通 DELETE 删索引行（'delete' 特殊命令是给外部内容表用的）
CREATE TRIGGER IF NOT EXISTS article_fts_insert AFTER INSERT ON article BEGIN
    INSERT INTO article_fts (rowid, title, summary, content)
    VALUES (new.id, new.title, new.summary, new.content);
END;

CREATE TRIGGER IF NOT EXISTS article_fts_update AFTER UPDATE ON article BEGIN
    DELETE FROM article_fts WHERE rowid = old.id;
    INSERT INTO article_fts (rowid, title, summary, content)
    VALUES (new.id, new.title, new.summary, new.content);
END;

CREATE TRIGGER IF NOT EXISTS article_fts_delete AFTER DELETE ON article BEGIN
    DELETE FROM article_fts WHERE rowid = old.id;
END;
