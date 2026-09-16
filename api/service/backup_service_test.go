package service

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"interview-sim/config"
)

func TestParseBackupRoundTrip(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString([]byte("\x00binary\x01\xff data"))
	entries := []backupEntry{
		{Key: "user:u1", TTLms: -1, Payload: payload},
		{Key: "report:share:abc", TTLms: 604200000, Payload: payload},
		{Key: "interview:session:s1", TTLms: 0, Payload: payload},
	}
	var sb strings.Builder
	for _, e := range entries {
		line, err := marshalEntry(e)
		if err != nil {
			t.Fatalf("序列化失败: %v", err)
		}
		sb.Write(line)
		sb.WriteByte('\n')
	}
	// 末尾空行与多余换行都应被忽略
	got, err := parseBackup(sb.String() + "\n\n")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(got) != len(entries) {
		t.Fatalf("条数不符: got %d want %d", len(got), len(entries))
	}
	for i := range got {
		if got[i].Key != entries[i].Key || got[i].TTLms != entries[i].TTLms || got[i].Payload != entries[i].Payload {
			t.Fatalf("第 %d 条不符: got %+v want %+v", i, got[i], entries[i])
		}
	}
	// 二进制 payload 必须能原样解回
	raw, err := base64.StdEncoding.DecodeString(got[0].Payload)
	if err != nil || string(raw) != "\x00binary\x01\xff data" {
		t.Fatalf("payload 二进制还原失败: %v %q", err, raw)
	}
}

func TestParseBackupRejectsBadLines(t *testing.T) {
	if _, err := parseBackup("not json\n"); err == nil {
		t.Fatal("非法 JSON 应报错")
	}
	if _, err := parseBackup(`{"t":-1,"p":"AAA="}` + "\n"); err == nil {
		t.Fatal("缺 key 应报错")
	}
	// 纯空白应视为空备份而非错误
	entries, err := parseBackup("   \n\n")
	if err != nil || len(entries) != 0 {
		t.Fatalf("空白应返回空备份: %v %v", err, entries)
	}
}

func TestBackupObjectKeys(t *testing.T) {
	restore := withTestConfig(t, &config.Config{COSPrefix: "/interview-backup/"})
	defer restore()
	if got, want := latestKey(), "interview-backup/redis-latest.jsonl"; got != want {
		t.Fatalf("latestKey = %q, want %q", got, want)
	}
	got := dailyKey(time.Date(2026, 9, 16, 10, 0, 0, 0, time.Local))
	if want := "interview-backup/redis-2026-09-16.jsonl"; got != want {
		t.Fatalf("dailyKey = %q, want %q", got, want)
	}
}

func TestBackupEnabledRequiresFullConfig(t *testing.T) {
	full := config.Config{COSBucket: "b-1250000", COSRegion: "ap-shanghai", COSSecretID: "id", COSSecretKey: "sk"}
	restore := withTestConfig(t, &full)
	defer restore()
	if !BackupEnabled() {
		t.Fatal("配置齐全时应启用备份")
	}
	for name, c := range map[string]config.Config{
		"缺桶":  {COSRegion: "ap-shanghai", COSSecretID: "id", COSSecretKey: "sk"},
		"缺地域": {COSBucket: "b-1250000", COSSecretID: "id", COSSecretKey: "sk"},
		"缺密钥": {COSBucket: "b-1250000", COSRegion: "ap-shanghai"},
		"全空":  {},
	} {
		cfg := c
		restoreInner := withTestConfig(t, &cfg)
		enabled := BackupEnabled()
		restoreInner()
		if enabled {
			t.Fatalf("%s 时不应启用备份", name)
		}
	}
}

// withTestConfig 临时替换全局配置，返回恢复函数（避免影响同包其他测试）
func withTestConfig(t *testing.T, c *config.Config) func() {
	t.Helper()
	old := config.Cfg
	config.Cfg = c
	return func() { config.Cfg = old }
}
