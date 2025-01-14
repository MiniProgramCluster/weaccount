CREATE DATABASE IF NOT EXISTS `weaccount`;

USE `weaccount`;

DROP TABLE IF EXISTS `users`;
CREATE TABLE IF NOT EXISTS `users` (
    `uid` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,  	    -- 管理员ID
    `appid` VARCHAR(64) NOT NULL,                           -- 玩家appid
    `openid` VARCHAR(64) NOT NULL,                          -- openid
    `unionid` VARCHAR(64) DEFAULT '' NOT NULL,              -- unionid
    `session_key` VARCHAR(255) DEFAULT '' NOT NULL,         -- 玩家session_key
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,       -- 账号创建时间，默认当前时间戳
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `appid` (`appid`, `openid`)
) ENGINE=InnoDB AUTO_INCREMENT=1000000 DEFAULT CHARSET=utf8mb4;
