CREATE TABLE `admin` (
                         `id`            BIGINT       NOT NULL AUTO_INCREMENT,
                         `username`      VARCHAR(64)  NOT NULL DEFAULT '',
                         `password_hash` VARCHAR(128) NOT NULL DEFAULT '',
                         `nickname`      VARCHAR(64)  NOT NULL DEFAULT '',
                         `avatar_url`    VARCHAR(255) NOT NULL DEFAULT '',
                         `created_at`    DATETIME     DEFAULT CURRENT_TIMESTAMP,
                         `updated_at`    DATETIME     DEFAULT CURRENT_TIMESTAMP,
                         PRIMARY KEY (`id`),
                         UNIQUE KEY `uk_username` (`username`)
);