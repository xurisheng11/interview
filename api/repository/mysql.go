package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"interview-sim/config"

	_ "github.com/go-sql-driver/mysql"
)

// mysqlDB 全局 MySQL 连接池；为 nil 表示 MySQL 不可用（纯 Redis 模式）
var mysqlDB *sql.DB

// Ping 结果短暂缓存，避免每次请求都 Ping
var (
	mysqlPingMu    sync.Mutex
	mysqlLastPing  time.Time
	mysqlLastAlive bool
)

const mysqlPingCacheTTL = 30 * time.Second

// InitMySQL 初始化 MySQL：建库 -> 建连接池 -> 执行迁移脚本建表
// 失败时返回 error，由调用方决定是否退化为纯 Redis 模式（不退出进程）
// 注意：建表由 db/migrations/*.sql 迁移脚本自动完成
func InitMySQL() error {
	if !config.Cfg.MySQLEnabled {
		return errors.New("MySQL 已通过配置禁用（MYSQL_ENABLED=false）")
	}

	cfg := config.Cfg
	// 1. 不带库名的 DSN，先建库
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=true&loc=Local&timeout=5s",
		cfg.MySQLUser, cfg.MySQLPassword, cfg.MySQLHost, cfg.MySQLPort)
	rootDB, err := sql.Open("mysql", rootDSN)
	if err != nil {
		return fmt.Errorf("打开 MySQL 连接失败: %w", err)
	}
	defer rootDB.Close()
	if err := rootDB.Ping(); err != nil {
		return fmt.Errorf("MySQL 连接不通: %w", err)
	}
	createDBSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.MySQLDB)
	if _, err := rootDB.Exec(createDBSQL); err != nil {
		return fmt.Errorf("创建数据库失败: %w", err)
	}

	// 2. 带库名的 DSN，建立全局连接池
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local&timeout=5s",
		cfg.MySQLUser, cfg.MySQLPassword, cfg.MySQLHost, cfg.MySQLPort, cfg.MySQLDB)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("打开 MySQL 业务库连接失败: %w", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("MySQL 业务库连接不通: %w", err)
	}

	mysqlDB = db
	mysqlPingMu.Lock()
	mysqlLastPing = time.Now()
	mysqlLastAlive = true
	mysqlPingMu.Unlock()

	// 3. 执行迁移脚本
	if err := runMigrations(db); err != nil {
		db.Close()
		return fmt.Errorf("迁移脚本执行失败: %w", err)
	}

	return nil
}

// runMigrations 按文件名顺序执行 db/migrations 目录下的 .sql 迁移脚本
func runMigrations(db *sql.DB) error {
	// 查找 api 的上一级目录（即项目根目录）
	exec, _ := os.Executable()
	apiDir := filepath.Dir(exec)                    // api 二进制所在目录
	projectRoot := filepath.Dir(apiDir)              // 项目根目录
	migrationsDir := filepath.Join(projectRoot, "db", "migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("迁移目录不存在，跳过: %s", migrationsDir)
			return nil
		}
		return fmt.Errorf("读取迁移目录失败: %w", err)
	}

	// 收集 .sql 文件并排序（确保按版本顺序执行）
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		path := filepath.Join(migrationsDir, f)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取迁移文件失败 [%s]: %w", f, err)
		}
		sql := strings.TrimSpace(string(sqlBytes))
		if sql == "" {
			continue
		}
		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("执行迁移文件失败 [%s]: %w", f, err)
		}
		log.Printf("迁移文件执行成功: %s", f)
	}
	return nil
}

// MySQLAvailable 判断 MySQL 是否可用（DB 非 nil 且 Ping 通，Ping 结果缓存 30 秒）
func MySQLAvailable() bool {
	if mysqlDB == nil {
		return false
	}
	mysqlPingMu.Lock()
	defer mysqlPingMu.Unlock()
	if time.Since(mysqlLastPing) < mysqlPingCacheTTL {
		return mysqlLastAlive
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := mysqlDB.PingContext(ctx)
	mysqlLastPing = time.Now()
	mysqlLastAlive = err == nil
	if err != nil {
		log.Printf("MySQL Ping 失败，暂时退化为纯 Redis 模式: %v", err)
	}
	return mysqlLastAlive
}

// ---------- 以下为读路径回源（Redis miss -> MySQL）使用的查询辅助 ----------
// 约定：未查到返回 ("", nil)，调用方保持原有 not found 语义

// queryMySQLString 查询单个字符串列，sql.ErrNoRows 转为空字符串
func queryMySQLString(query string, args ...interface{}) (string, error) {
	var s string
	err := mysqlDB.QueryRow(query, args...).Scan(&s)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return s, nil
}

// QueryUserDataByID 从 MySQL 查用户全量 JSON
func QueryUserDataByID(userID string) (string, error) {
	return queryMySQLString("SELECT data FROM users WHERE user_id = ?", userID)
}

// QueryUserDataByAccount 按 phone→email→username 优先级精确单列查询用户全量 JSON，
// 避免跨列碰撞时 OR+LIMIT 1 返回不确定行
func QueryUserDataByAccount(account string) (string, error) {
	for _, col := range []string{"phone", "email", "username"} {
		data, err := queryMySQLString("SELECT data FROM users WHERE "+col+" = ? LIMIT 1", account)
		if err != nil {
			return "", err
		}
		if data != "" {
			return data, nil
		}
	}
	return "", nil
}

// QuerySessionData 从 MySQL 查面试会话全量 JSON
func QuerySessionData(interviewID string) (string, error) {
	return queryMySQLString("SELECT data FROM interview_sessions WHERE interview_id = ?", interviewID)
}

// QueryUserSessionIDs 按用户查最近 limit 条面试（含开始时间，用于重建 zset）
func QueryUserSessionIDs(userID string, limit int) ([]string, []time.Time, error) {
	rows, err := mysqlDB.Query(
		"SELECT interview_id, start_time FROM interview_sessions WHERE user_id = ? ORDER BY start_time DESC LIMIT ?",
		userID, limit)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var ids []string
	var times []time.Time
	for rows.Next() {
		var id string
		var t sql.NullTime
		if err := rows.Scan(&id, &t); err != nil {
			return nil, nil, err
		}
		ids = append(ids, id)
		times = append(times, t.Time)
	}
	return ids, times, rows.Err()
}

// QueryResumeData 从 MySQL 查简历全量 JSON
func QueryResumeData(resumeID string) (string, error) {
	return queryMySQLString("SELECT data FROM resumes WHERE resume_id = ?", resumeID)
}

// QueryUserResumeIDs 按用户查最近 limit 条简历（含上传时间，用于重建 zset）
func QueryUserResumeIDs(userID string, limit int) ([]string, []time.Time, error) {
	rows, err := mysqlDB.Query(
		"SELECT resume_id, uploaded_at FROM resumes WHERE user_id = ? ORDER BY uploaded_at DESC LIMIT ?",
		userID, limit)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var ids []string
	var times []time.Time
	for rows.Next() {
		var id string
		var t sql.NullTime
		if err := rows.Scan(&id, &t); err != nil {
			return nil, nil, err
		}
		ids = append(ids, id)
		times = append(times, t.Time)
	}
	return ids, times, rows.Err()
}

// QueryReportData 从 MySQL 查报告全量 JSON（校验 user_id 归属）
func QueryReportData(userID, interviewID string) (string, error) {
	return queryMySQLString(
		"SELECT data FROM reports WHERE interview_id = ? AND user_id = ?",
		interviewID, userID)
}

// QuerySprintPlanData 从 MySQL 查冲刺计划全量 JSON
func QuerySprintPlanData(userID string) (string, error) {
	return queryMySQLString("SELECT data FROM sprint_plans WHERE user_id = ?", userID)
}
