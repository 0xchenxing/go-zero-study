CREATE TABLE `user`
(
    `id`         bigint       NOT NULL AUTO_INCREMENT,
    `username`   varchar(255) NOT NULL DEFAULT '',
    `password`   varchar(255) NOT NULL DEFAULT '',
    `mobile`     varchar(20)  NOT NULL DEFAULT '',
    `created_at` timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_username` (`username`),
    UNIQUE KEY `idx_mobile` (`mobile`)
) ENGINE=InnoDB;