-- 修正 0009 的触发器：全内容 FTS5 表用普通 DELETE 删索引行。
-- 0009 初版误用了 'delete' 特殊命令（那是外部内容/无内容表的语法），
-- 导致 article 的任何 UPDATE/DELETE 都报 SQL logic error。
-- 本迁移幂等：无论旧触发器是否存在都能修好。
DROP TRIGGER IF EXISTS article_fts_insert;
DROP TRIGGER IF EXISTS article_fts_update;
DROP TRIGGER IF EXISTS article_fts_delete;

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
