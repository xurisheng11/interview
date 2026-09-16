package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"interview-sim/config"
	"interview-sim/repository"

	"github.com/go-redis/redis/v8"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// 云托管容器文件系统是临时的：发布新版本、闲置缩容冷启动、容器崩溃重启都会清空 Redis/MySQL。
// 这里用对象存储（独立于容器的持久服务）做逻辑备份：
//   - 定期把 Redis 全部 key 以 DUMP+TTL 导出成 JSONL 上传 COS
//   - 启动时若发现 Redis 是空库，就从 COS 拉最近一次备份用 RESTORE 回灌
//
// MySQL 不需要单独备份：它是 Redis 的归档副本，启动时的 SyncAllToMySQL 会基于恢复后的 Redis 重建。

// backupEntry 单个 key 的逻辑快照
type backupEntry struct {
	Key     string `json:"k"`
	TTLms   int64  `json:"t"` // 剩余存活毫秒；-1 表示永久
	Payload string `json:"p"` // base64(Redis DUMP 序列化结果)
}

const (
	latestSuffix = "redis-latest.jsonl"
	// 单次备份体积上限，防止异常数据把容器内存打爆
	maxBackupBytes = 64 << 20
)

var backupMu sync.Mutex

// BackupEnabled 判断对象存储备份是否可用（未配置则整套机制静默跳过，不影响本地/其他环境）
func BackupEnabled() bool {
	return config.Cfg.COSBucket != "" && config.Cfg.COSRegion != "" &&
		config.Cfg.COSSecretID != "" && config.Cfg.COSSecretKey != ""
}

// ErrBackupDisabled 未配置对象存储时手动触发备份的提示
var ErrBackupDisabled = errors.New("未配置对象存储备份（请检查 COS_BUCKET / COS_REGION 与腾讯云密钥环境变量）")

// AdminTriggerBackup 管理端手动触发一次备份，便于验收而不用等定时周期
func AdminTriggerBackup() (map[string]interface{}, error) {
	if !BackupEnabled() {
		return nil, ErrBackupDisabled
	}
	n, err := UploadBackup("manual")
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"keys":      n,
		"bucket":    config.Cfg.COSBucket,
		"objectKey": latestKey(),
	}, nil
}

// AdminBackupStatus 备份配置与当前库规模，便于确认备份到底有没有生效
func AdminBackupStatus() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	size, _ := repository.RDB.DBSize(ctx).Result()
	return map[string]interface{}{
		"enabled":     BackupEnabled(),
		"bucket":      config.Cfg.COSBucket,
		"region":      config.Cfg.COSRegion,
		"prefix":      config.Cfg.COSPrefix,
		"latestKey":   latestKey(),
		"intervalMin": config.Cfg.BackupIntervalMin,
		"redisKeys":   size,
	}
}

// latestKey 最近一次备份的对象键
func latestKey() string {
	return strings.Trim(config.Cfg.COSPrefix, "/") + "/" + latestSuffix
}

// dailyKey 当天归档键（每天第一个备份会写一份，用于误删后回溯）
func dailyKey(now time.Time) string {
	return fmt.Sprintf("%s/redis-%s.jsonl", strings.Trim(config.Cfg.COSPrefix, "/"), now.Format("2006-01-02"))
}

func newCOSClient() (*cos.Client, error) {
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Cfg.COSBucket, config.Cfg.COSRegion))
	if err != nil {
		return nil, fmt.Errorf("COS_BUCKET/COS_REGION 配置不合法: %w", err)
	}
	return cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Timeout:   2 * time.Minute,
		Transport: &cos.AuthorizationTransport{SecretID: config.Cfg.COSSecretID, SecretKey: config.Cfg.COSSecretKey},
	}), nil
}

// dumpRedis 扫描当前库全部 key，导出为 JSONL 文本
func dumpRedis(ctx context.Context) (string, int, error) {
	var (
		cursor  uint64
		entries []backupEntry
		total   int
	)
	for {
		keys, next, err := repository.RDB.Scan(ctx, cursor, "*", 500).Result()
		if err != nil {
			return "", 0, fmt.Errorf("SCAN 失败: %w", err)
		}
		for _, key := range keys {
			payload, err := repository.RDB.Dump(ctx, key).Result()
			if errors.Is(err, redis.Nil) {
				continue // 扫描与导出之间 key 过期，跳过
			}
			if err != nil {
				return "", 0, fmt.Errorf("DUMP %s 失败: %w", key, err)
			}
			entries = append(entries, backupEntry{
				Key:     key,
				TTLms:   ttlMillis(ctx, key),
				Payload: base64.StdEncoding.EncodeToString([]byte(payload)),
			})
			total++
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}

	var sb strings.Builder
	for _, e := range entries {
		line, err := marshalEntry(e)
		if err != nil {
			return "", 0, err
		}
		sb.Write(line)
		sb.WriteByte('\n')
	}
	data := sb.String()
	if len(data) > maxBackupBytes {
		return "", 0, fmt.Errorf("备份体积 %d 超过上限 %d，已中止（请检查是否有异常大 key）", len(data), maxBackupBytes)
	}
	return data, total, nil
}

// ttlMillis 返回 key 剩余存活毫秒，永久（或已消失）返回 -1
func ttlMillis(ctx context.Context, key string) int64 {
	d, err := repository.RDB.TTL(ctx, key).Result()
	if err != nil || d < 0 {
		return -1
	}
	return int64(d / time.Millisecond)
}

// marshalEntry 将单个备份条目编成一行 JSON
func marshalEntry(e backupEntry) ([]byte, error) {
	line, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("序列化备份条目失败: %w", err)
	}
	return line, nil
}

// parseBackup 解析 JSONL 备份文本
func parseBackup(data string) ([]backupEntry, error) {
	var out []backupEntry
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e backupEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, fmt.Errorf("备份行解析失败: %w", err)
		}
		if e.Key == "" {
			return nil, errors.New("备份行缺少 key 字段")
		}
		out = append(out, e)
	}
	return out, nil
}

// restoreRedis 用 RESTORE REPLACE 回灌备份，返回成功条数
func restoreRedis(ctx context.Context, entries []backupEntry) (int, error) {
	ok := 0
	var failed []string
	for _, e := range entries {
		payload, err := base64.StdEncoding.DecodeString(e.Payload)
		if err != nil {
			failed = append(failed, e.Key+"(base64)")
			continue
		}
		ttl := time.Duration(0) // 0 = 永久
		if e.TTLms > 0 {
			ttl = time.Duration(e.TTLms) * time.Millisecond
		}
		if err := repository.RDB.RestoreReplace(ctx, e.Key, ttl, string(payload)).Err(); err != nil {
			// BUSYKEY 之外的错误（如序列化版本不兼容）逐条记录但不中断，尽量抢救能救的
			failed = append(failed, e.Key+"("+err.Error()+")")
			continue
		}
		ok++
	}
	if len(failed) > 0 {
		log.Printf("Redis 回灌有 %d 个 key 失败（示例：%s）", len(failed), strings.Join(failed[:minInt(3, len(failed))], ", "))
	}
	return ok, nil
}

// UploadBackup 导出一份 Redis 快照上传对象存储；返回导出 key 数
func UploadBackup(reason string) (int, error) {
	if !BackupEnabled() {
		return 0, nil
	}
	backupMu.Lock()
	defer backupMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	data, count, err := dumpRedis(ctx)
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, nil // 空库不覆盖已有备份，避免把好消息冲掉
	}

	client, err := newCOSClient()
	if err != nil {
		return 0, err
	}
	now := time.Now()
	if _, err := client.Object.Put(ctx, latestKey(), strings.NewReader(data), nil); err != nil {
		return 0, fmt.Errorf("上传 %s 失败: %w", latestKey(), err)
	}
	// 每天留一份归档（已存在就不重复写）
	dk := dailyKey(now)
	if exist, _ := client.Object.IsExist(ctx, dk); !exist {
		if _, err := client.Object.Put(ctx, dk, strings.NewReader(data), nil); err != nil {
			log.Printf("每日归档写入失败（不影响主备份）: %v", err)
		}
	}
	log.Printf("对象存储备份完成（%s）：%d 个 key，%d 字节", reason, count, len(data))
	return count, nil
}

// RestoreFromCOSIfEmpty 启动时调用：Redis 为空则从对象存储回灌最近备份
func RestoreFromCOSIfEmpty() error {
	if !BackupEnabled() {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	size, err := repository.RDB.DBSize(ctx).Result()
	if err != nil {
		return fmt.Errorf("查询 Redis key 数失败: %w", err)
	}
	if size > 0 {
		log.Printf("Redis 已有 %d 个 key，跳过对象存储回灌", size)
		return nil
	}

	client, err := newCOSClient()
	if err != nil {
		return err
	}
	resp, err := client.Object.Get(ctx, latestKey(), nil)
	if err != nil {
		log.Printf("Redis 为空且对象存储无备份（或读取失败），按空库启动: %v", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取备份对象失败: %w", err)
	}
	entries, err := parseBackup(string(body))
	if err != nil {
		return err
	}
	ok, err := restoreRedis(ctx, entries)
	if err != nil {
		return err
	}
	log.Printf("已从对象存储回灌 Redis：%d / %d 个 key", ok, len(entries))
	return nil
}

// StartBackupLoop 启动定时备份，并在容器收到下线信号时补一次备份
// （云托管发布新版本/缩容前会先发 SIGTERM，这是数据留档的最后一道机会）
func StartBackupLoop() {
	if !BackupEnabled() {
		log.Println("未配置 COS_BUCKET/COS_REGION，对象存储备份关闭（容器重启数据会丢）")
		return
	}
	interval := time.Duration(config.Cfg.BackupIntervalMin) * time.Minute
	log.Printf("对象存储备份已开启，间隔 %v，桶 %s 前缀 %s", interval, config.Cfg.COSBucket, config.Cfg.COSPrefix)

	// 启动即备份一次：把刚回灌的数据原样留档，也便于确认凭证可用
	go func() {
		if _, err := UploadBackup("startup"); err != nil {
			log.Printf("启动备份失败: %v", err)
		}
	}()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			if _, err := UploadBackup("ticker"); err != nil {
				log.Printf("定时备份失败: %v", err)
			}
		}
	}()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
		sig := <-ch
		log.Printf("收到 %v，下线前执行最后一次备份", sig)
		if _, err := UploadBackup("shutdown"); err != nil {
			log.Printf("下线备份失败: %v", err)
		}
		// 信号已被 Notify 接管，默认终止行为不会发生，备份完必须主动退出
		os.Exit(0)
	}()
}
