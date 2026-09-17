package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"interview-sim/config"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// 头像放对象存储，不放本地目录也不放 Redis：
//   - 容器文件系统是临时的，写到本地磁盘发布/缩容就没了
//   - 用 base64 塞进 user hash 会把单个 key 撑到几百 KB，还会被每次全量备份整体带走（有 64MB 上限）
//
// 存储桶本身保持私有读写（里面还有全量数据备份，不能开公共读），
// 只给 avatars/* 这些对象单独加对象级 public-read。已实测：
//
//	avatars/_probe.txt 匿名 GET -> 200，interview-backup/redis-latest.jsonl 匿名 GET -> 403
const (
	avatarPrefix   = "avatars"
	maxAvatarBytes = 2 << 20
)

// ErrAvatarStorageUnavailable 未配置对象存储时无法保存头像
var ErrAvatarStorageUnavailable = errors.New("未配置对象存储（COS_BUCKET/COS_REGION），头像无法保存")

// avatarExts 只认这几种图片类型，并映射成对应扩展名，
// 避免这个口子被用来往桶里写 html / 可执行文件
var avatarExts = map[string]string{
	"image/jpeg":  "jpg",
	"image/pjpeg": "jpg",
	"image/png":   "png",
	"image/webp":  "webp",
	"image/gif":   "gif",
}

// SaveAvatar 上传头像并返回可匿名访问的地址。
// oldAvatar 若确实是该用户此前的头像对象，会一并删除，避免换一次头像就在桶里留一份垃圾。
func SaveAvatar(userID string, data []byte, contentType string, oldAvatar string) (string, error) {
	if !BackupEnabled() {
		return "", ErrAvatarStorageUnavailable
	}
	segment := avatarKeySegment(userID)
	if segment == "" {
		return "", errors.New("缺少用户身份")
	}
	if len(data) == 0 {
		return "", errors.New("头像内容为空")
	}
	if len(data) > maxAvatarBytes {
		return "", fmt.Errorf("头像超过 %dMB 限制", maxAvatarBytes>>20)
	}
	ext, ok := avatarExtOf(contentType)
	if !ok {
		return "", errors.New("只支持 JPG / PNG / WEBP / GIF 图片")
	}

	// 时间戳做文件名：同一用户换头像产生新 URL，浏览器与微信图片缓存都不会串
	key := fmt.Sprintf("%s/%s/%d.%s", avatarPrefix, segment, time.Now().UnixNano(), ext)

	client, err := newCOSClient()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	opt := &cos.ObjectPutOptions{
		ACLHeaderOptions: &cos.ACLHeaderOptions{XCosACL: "public-read"},
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType:  contentType,
			CacheControl: "public, max-age=31536000", // 文件名唯一，可以往死里缓存
		},
	}
	if _, err := client.Object.Put(ctx, key, bytes.NewReader(data), opt); err != nil {
		return "", fmt.Errorf("头像写入对象存储失败: %w", err)
	}

	publicURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s",
		config.Cfg.COSBucket, config.Cfg.COSRegion, key)

	if err := deleteAvatarObject(ctx, client, segment, oldAvatar); err != nil {
		// 旧头像没删掉不影响新头像生效，留日志便于人工清理
		log.Printf("旧头像清理失败（新头像已生效）: %v", err)
	}
	return publicURL, nil
}

// deleteAvatarObject 只允许删本人 avatars/<userID>/ 下的对象。
// 历史脏数据（手机本地路径 wxfile://、base64）不在对象存储里，直接忽略。
func deleteAvatarObject(ctx context.Context, client *cos.Client, segment, oldAvatar string) error {
	if !strings.HasPrefix(oldAvatar, "https://") {
		return nil
	}
	u, err := url.Parse(oldAvatar)
	if err != nil {
		return err
	}
	prefix := fmt.Sprintf("/%s/%s/", avatarPrefix, segment)
	if !strings.HasPrefix(u.Path, prefix) {
		return fmt.Errorf("拒绝删除非本人头像的对象键 %s", u.Path)
	}
	if _, err := client.Object.Delete(ctx, strings.TrimPrefix(u.Path, "/")); err != nil {
		return err
	}
	return nil
}

// avatarExtOf 取 MIME 对应的扩展名；带 charset 等参数时只取分号前部分
func avatarExtOf(contentType string) (string, bool) {
	mime := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	ext, ok := avatarExts[mime]
	return ext, ok
}

// avatarKeySegment 用户 ID 正常是 UUID，这里兜一层：
// 异常字符一律剔除，防止拼进对象键后改变目录层级（../ 逃逸到备份前缀下）
func avatarKeySegment(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
