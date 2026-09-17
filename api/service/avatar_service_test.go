package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"interview-sim/config"
)

// 头像上传直传对象存储，Put/GET 的真实链路已在上线前用 cmd/coscheck 实测过
// （私有桶 + 对象级 public-read：avatars/* 匿名 200，interview-backup/* 仍 403）。
// 这里锁的是不依赖网络的安全不变量：类型白名单、对象键不可逃逸、不许删别人的对象。

func TestAvatarExtOfWhitelist(t *testing.T) {
	ok := []string{"image/jpeg", "image/png", "image/webp", "image/gif", "image/PNG", " image/jpeg ; charset=binary"}
	for _, ct := range ok {
		if _, valid := avatarExtOf(ct); !valid {
			t.Errorf("应接受 %q", ct)
		}
	}
	// 这些如果被放行，就等于往一个匿名可读的桶里开了任意文件上传口
	bad := []string{"", "text/html", "image/svg+xml", "application/x-msdownload", "multipart/form-data"}
	for _, ct := range bad {
		if ext, valid := avatarExtOf(ct); valid {
			t.Errorf("应拒绝 %q，却给了扩展名 %q", ct, ext)
		}
	}
}

func TestAvatarKeySegmentStripsTraversal(t *testing.T) {
	cases := map[string]string{
		"ec4db718-647f-4899-b20e-3ea19f383c9a": "ec4db718-647f-4899-b20e-3ea19f383c9a",
		"../../interview-backup":               "interview-backup",
		"a/b\\c..d":                            "abcd",
		"":                                     "",
	}
	for in, want := range cases {
		if got := avatarKeySegment(in); got != want {
			t.Errorf("avatarKeySegment(%q) = %q，期望 %q", in, got, want)
		}
	}
}

func TestDeleteAvatarObjectIgnoresNonCOSValues(t *testing.T) {
	// 历史脏数据（手机本地临时路径、web 端的 base64）不在对象存储里，必须静默跳过而不是报错，
	// 否则用户换个头像就会被旧数据卡住
	for _, old := range []string{"", "wxfile://tmp_83348a89.jpg", "data:image/png;base64,iVBORw0KGgo="} {
		if err := deleteAvatarObject(context.Background(), nil, "u1", old); err != nil {
			t.Errorf("旧值 %q 应被忽略，却报错：%v", old, err)
		}
	}
}

func TestDeleteAvatarObjectRejectsForeignObject(t *testing.T) {
	// 越界对象必须在真正调用 Delete 之前就被拦下，所以传 nil client 也不会 panic
	foreign := "https://bucket.cos.ap-guangzhou.myqcloud.com/interview-backup/redis-latest.jsonl"
	err := deleteAvatarObject(context.Background(), nil, "u1", foreign)
	if err == nil {
		t.Fatal("应拒绝删除别人/备份前缀下的对象")
	}
	if !strings.Contains(err.Error(), "拒绝") {
		t.Errorf("错误信息应说明是主动拒绝，实际：%v", err)
	}

	otherUser := "https://bucket.cos.ap-guangzhou.myqcloud.com/avatars/u2/1.jpg"
	if err := deleteAvatarObject(context.Background(), nil, "u1", otherUser); err == nil {
		t.Error("应拒绝删除其他用户的头像对象")
	}
}

func TestSaveAvatarRejectsBadInputBeforeAnyNetwork(t *testing.T) {
	// 未配置对象存储时直接拒绝，且不能 panic（config.Cfg 在测试进程里是手工塞的）
	prev := config.Cfg
	config.Cfg = &config.Config{}
	t.Cleanup(func() { config.Cfg = prev })

	if _, err := SaveAvatar("u1", []byte("x"), "image/png", ""); !errors.Is(err, ErrAvatarStorageUnavailable) {
		t.Errorf("未配置 COS 时应返回 ErrAvatarStorageUnavailable，实际：%v", err)
	}
}

func TestSaveAvatarValidatesBeforeStorageCheckOrder(t *testing.T) {
	prev := config.Cfg
	config.Cfg = &config.Config{COSBucket: "b", COSRegion: "r", COSSecretID: "id", COSSecretKey: "k"}
	t.Cleanup(func() { config.Cfg = prev })

	// 这几条都应在发起任何请求前失败，否则会拿假桶名去打真实 COS
	if _, err := SaveAvatar("", []byte("x"), "image/png", ""); err == nil {
		t.Error("缺少用户身份应失败")
	}
	if _, err := SaveAvatar("u1", nil, "image/png", ""); err == nil {
		t.Error("空内容应失败")
	}
	if _, err := SaveAvatar("u1", make([]byte, maxAvatarBytes+1), "image/png", ""); err == nil {
		t.Error("超过体积上限应失败")
	}
	if _, err := SaveAvatar("u1", []byte("x"), "text/html", ""); err == nil {
		t.Error("非图片类型应失败")
	}
}
