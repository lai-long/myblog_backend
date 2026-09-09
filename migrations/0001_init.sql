CREATE TABLE IF NOT EXISTS admin (
                                     id            INTEGER PRIMARY KEY AUTOINCREMENT,
                                     username      TEXT UNIQUE NOT NULL,
                                     password_hash TEXT NOT NULL,
                                     nickname      TEXT NOT NULL DEFAULT '',
                                     avatar_url    TEXT NOT NULL DEFAULT '',
                                     created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
                                     updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);