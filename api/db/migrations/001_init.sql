-- ============================================================
-- 数据库迁移脚本 v1.0
-- 适用版本：interview-sim
-- 说明：按顺序执行，幂等设计（IF NOT EXISTS）
-- ============================================================

-- 第1步：创建数据库
CREATE DATABASE IF NOT EXISTS `interview_sim`
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE `interview_sim`;

-- ============================================================
-- 用户相关
-- ============================================================

-- users：用户表
CREATE TABLE IF NOT EXISTS `users` (
    `user_id`      VARCHAR(64)  NOT NULL,
    `username`     VARCHAR(191) NOT NULL DEFAULT '',
    `phone`        VARCHAR(32)  NOT NULL DEFAULT '',
    `email`        VARCHAR(191) NOT NULL DEFAULT '',
    `created_at`   DATETIME NULL,
    `data`         JSON,
    `synced_at`    DATETIME NULL,
    PRIMARY KEY (`user_id`),
    KEY `idx_username` (`username`),
    KEY `idx_phone`    (`phone`),
    KEY `idx_email`    (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- user_credits：用户积分表
CREATE TABLE IF NOT EXISTS `user_credits` (
    `user_id`   VARCHAR(64) NOT NULL,
    `balance`   BIGINT      NOT NULL DEFAULT 0,
    `synced_at` DATETIME    NULL,
    PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- user_question_collects：题目收藏表
CREATE TABLE IF NOT EXISTS `user_question_collects` (
    `user_id`     VARCHAR(64) NOT NULL,
    `question_id` VARCHAR(64) NOT NULL,
    PRIMARY KEY (`user_id`, `question_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 面试相关
-- ============================================================

-- interview_sessions：面试会话表
CREATE TABLE IF NOT EXISTS `interview_sessions` (
    `interview_id` VARCHAR(64)  NOT NULL,
    `user_id`      VARCHAR(64)  NOT NULL DEFAULT '',
    `status`       VARCHAR(16)  NOT NULL DEFAULT '',
    `mode`         VARCHAR(16)  NOT NULL DEFAULT '',
    `start_time`   DATETIME NULL,
    `data`         JSON,
    `synced_at`    DATETIME NULL,
    PRIMARY KEY (`interview_id`),
    KEY `idx_user_time` (`user_id`, `start_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- reports：面试报告表
CREATE TABLE IF NOT EXISTS `reports` (
    `interview_id` VARCHAR(64)  NOT NULL,
    `user_id`      VARCHAR(64)  NOT NULL DEFAULT '',
    `total_score`  INT          NOT NULL DEFAULT 0,
    `created_at`   DATETIME NULL,
    `data`         JSON,
    `synced_at`    DATETIME NULL,
    PRIMARY KEY (`interview_id`),
    KEY `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- resumes：简历表
CREATE TABLE IF NOT EXISTS `resumes` (
    `resume_id`        VARCHAR(64)  NOT NULL,
    `user_id`          VARCHAR(64)  NOT NULL DEFAULT '',
    `uploaded_at`      DATETIME NULL,
    `analysis_status`  VARCHAR(16)  NOT NULL DEFAULT '',
    `data`             JSON,
    `synced_at`        DATETIME NULL,
    PRIMARY KEY (`resume_id`),
    KEY `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 社区相关
-- ============================================================

-- articles：文章表
CREATE TABLE IF NOT EXISTS `articles` (
    `article_id`   VARCHAR(64) NOT NULL,
    `job_category` VARCHAR(64) NOT NULL DEFAULT '',
    `author_id`    VARCHAR(64) NOT NULL DEFAULT '',
    `created_at`   BIGINT      NOT NULL DEFAULT 0,
    `data`         JSON,
    `synced_at`    DATETIME NULL,
    PRIMARY KEY (`article_id`),
    KEY `idx_cat` (`job_category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- comments：评论表
CREATE TABLE IF NOT EXISTS `comments` (
    `comment_id` VARCHAR(64) NOT NULL,
    `article_id` VARCHAR(64) NOT NULL DEFAULT '',
    `created_at` BIGINT      NOT NULL DEFAULT 0,
    `data`       JSON,
    `synced_at`  DATETIME NULL,
    PRIMARY KEY (`comment_id`),
    KEY `idx_article` (`article_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- article_relations：文章用户关系表（点赞/收藏等）
CREATE TABLE IF NOT EXISTS `article_relations` (
    `article_id` VARCHAR(64) NOT NULL,
    `user_id`    VARCHAR(64) NOT NULL,
    `rel_type`   VARCHAR(16) NOT NULL,
    PRIMARY KEY (`article_id`, `user_id`, `rel_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 题目贡献相关
-- ============================================================

-- contributed_questions：贡献题目表
CREATE TABLE IF NOT EXISTS `contributed_questions` (
    `question_id`     VARCHAR(64) NOT NULL,
    `company`         VARCHAR(128) NOT NULL DEFAULT '',
    `contributor_id`  VARCHAR(64)  NOT NULL DEFAULT '',
    `status`          VARCHAR(16) NOT NULL DEFAULT '',
    `created_at`      BIGINT      NOT NULL DEFAULT 0,
    `data`            JSON,
    `synced_at`       DATETIME NULL,
    PRIMARY KEY (`question_id`),
    KEY `idx_company` (`company`),
    KEY `idx_user`    (`contributor_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- contribution_relations：贡献题目用户关系表
CREATE TABLE IF NOT EXISTS `contribution_relations` (
    `question_id` VARCHAR(64) NOT NULL,
    `user_id`      VARCHAR(64) NOT NULL,
    `rel_type`     VARCHAR(16) NOT NULL,
    PRIMARY KEY (`question_id`, `user_id`, `rel_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 冲刺计划相关
-- ============================================================

-- sprint_plans：冲刺计划表
CREATE TABLE IF NOT EXISTS `sprint_plans` (
    `user_id`   VARCHAR(64) NOT NULL,
    `data`      JSON,
    `synced_at` DATETIME NULL,
    PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
