-- 站点配置 KV 表：以后加新配置项不用改表结构
CREATE TABLE IF NOT EXISTS settings (
                                      key        TEXT PRIMARY KEY,
                                      value      TEXT NOT NULL DEFAULT '',
                                      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
