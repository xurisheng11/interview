// 诊断工具：确认密钥所属账号、桶归属、读写权限，定位 COS 备份 403 的根因。
//
// 用法（凭证只从环境变量读，不写进代码）：
//
//	cd api
//	$env:COS_SECRET_ID="AKID..."; $env:COS_SECRET_KEY="..."
//	$env:COS_BUCKET="interview-bucket-1326903464"; $env:COS_REGION="ap-guangzhou"
//	go run ./cmd/coscheck
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

func main() {
	id := firstNonEmpty(os.Getenv("COS_SECRET_ID"), os.Getenv("TENCENT_SECRET_ID"))
	key := firstNonEmpty(os.Getenv("COS_SECRET_KEY"), os.Getenv("TENCENT_SECRET_KEY"))
	bucket := os.Getenv("COS_BUCKET")
	region := os.Getenv("COS_REGION")

	if id == "" || key == "" || bucket == "" || region == "" {
		fmt.Println("缺少环境变量：COS_SECRET_ID/COS_SECRET_KEY/COS_BUCKET/COS_REGION")
		os.Exit(1)
	}
	fmt.Printf("密钥 %s...%s  桶 %s  地域 %s\n\n", id[:6], id[len(id)-4:], bucket, region)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// 1) 列出这对密钥所属账号名下的全部桶——官方语义：只能列出 AccessID 所属账号的桶
	svcURL, _ := url.Parse("https://service.cos.myqcloud.com")
	svc := cos.NewClient(&cos.BaseURL{ServiceURL: svcURL}, &http.Client{
		Timeout:   time.Minute,
		Transport: &cos.AuthorizationTransport{SecretID: id, SecretKey: key},
	})
	res, _, err := svc.Service.Get(ctx)
	if err != nil {
		fmt.Printf("[1] 列举桶失败：%v\n    （子账号通常没有 ListBuckets 权限，此项失败不代表读写不行）\n\n", err)
	} else {
		fmt.Printf("[1] 密钥所属账号 UIN=%s Owner.ID=%s，名下共 %d 个桶：\n", ownerUIN(res.Owner), ownerID(res.Owner), len(res.Buckets))
		found := false
		for _, b := range res.Buckets {
			mark := "  "
			if b.Name == bucket {
				mark = "→ "
				found = true
			}
			fmt.Printf("    %s%-42s %s\n", mark, b.Name, b.Region)
		}
		fmt.Printf("    结论：目标桶%s在该密钥所属账号名下\n\n", map[bool]string{true: "  ", false: " 不"}[found])
	}

	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region))
	if err != nil {
		fmt.Printf("桶地址不合法：%v\n", err)
		os.Exit(1)
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Timeout:   time.Minute,
		Transport: &cos.AuthorizationTransport{SecretID: id, SecretKey: key},
	})

	// -rm 模式：删除探测/残留对象，保持桶干净
	if len(os.Args) > 1 && os.Args[1] == "-rm" {
		if len(os.Args) < 3 {
			fmt.Println("用法：coscheck -rm <对象键>")
			return
		}
		for _, objKey := range os.Args[2:] {
			if _, err := client.Object.Delete(ctx, objKey); err != nil {
				fmt.Printf("RM %s 失败：%v\n", objKey, err)
				continue
			}
			fmt.Printf("RM %s 已删除\n", objKey)
		}
		return
	}

	// -pubput 模式：以 public-read 写入一个探测对象，实测私有桶能否被匿名读
	if len(os.Args) > 1 && os.Args[1] == "-pubput" {
		if len(os.Args) < 3 {
			fmt.Println("用法：coscheck -pubput <对象键>")
			return
		}
		probeKey := os.Args[2]
		opt := &cos.ObjectPutOptions{ACLHeaderOptions: &cos.ACLHeaderOptions{XCosACL: "public-read"}}
		if _, err := client.Object.Put(ctx, probeKey, strings.NewReader("public-read probe"), opt); err != nil {
			fmt.Printf("PUBPUT %s 失败：%v\n", probeKey, err)
			return
		}
		fmt.Printf("PUBPUT %s 成功（对象 ACL=public-read）\n", probeKey)
		fmt.Printf("匿名读地址 https://%s.cos.%s.myqcloud.com/%s\n", bucket, region, probeKey)
		return
	}

	// -copy 模式：桶内复制，把最新快照刷进当天归档（覆盖早期的残缺快照）
	if len(os.Args) > 1 && os.Args[1] == "-copy" {
		if len(os.Args) < 4 {
			fmt.Println("用法：coscheck -copy <源对象键> <目标对象键>")
			return
		}
		copyObject(ctx, client, bucket, os.Args[2], os.Args[3])
		return
	}

	// -save 模式：把对象下载到本地留底。多个实例互相覆盖备份时，先把当前这份抢下来再动手
	if len(os.Args) > 1 && os.Args[1] == "-save" {
		if len(os.Args) < 4 {
			fmt.Println("用法：coscheck -save <对象键> <本地文件>")
			return
		}
		saveObject(ctx, client, os.Args[2], os.Args[3])
		return
	}

	// 带对象键参数时进入只读校验模式：把备份文件从桶里读回来，确认写入真的落地
	if len(os.Args) > 1 {
		for _, objKey := range os.Args[1:] {
			checkObject(ctx, client, objKey)
		}
		return
	}

	// 2) HEAD 桶：验证「这个桶存在 + 是否有权查看元信息」
	resp, err := client.Bucket.Head(ctx)
	if err != nil {
		fmt.Printf("[2] HEAD 桶失败：%v\n\n", err)
	} else {
		fmt.Printf("[2] HEAD 桶：HTTP %d（桶存在且可见）\n\n", resp.StatusCode)
	}

	// 3) PUT 探测对象：备份真正需要的权限
	probe := "interview-backup/_probe.txt"
	if _, err := client.Object.Put(ctx, probe, strings.NewReader("probe"), nil); err != nil {
		fmt.Printf("[3] PUT %s 失败：%v\n", probe, err)
		fmt.Println("    => 备份会在这里失败，需要修权限或换桶")
		return
	}
	fmt.Printf("[3] PUT %s 成功\n", probe)
	if _, err := client.Object.Get(ctx, probe, nil); err != nil {
		fmt.Printf("[4] GET 失败：%v\n", err)
	} else {
		fmt.Printf("[4] GET %s 成功\n", probe)
	}
	if _, err := client.Object.Delete(ctx, probe); err != nil {
		fmt.Printf("    清理探测文件失败（可手动删）：%v\n", err)
	} else {
		fmt.Printf("[5] 已清理探测文件，读写删全通过 => 备份可用\n")
	}
}

func checkObject(ctx context.Context, client *cos.Client, key string) {
	resp, err := client.Object.Get(ctx, key, nil)
	if err != nil {
		fmt.Printf("GET %s 失败：%v\n", key, err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("GET %s 内容读取失败：%v\n", key, err)
		return
	}
	fmt.Printf("GET %s：HTTP %d，%d 字节\n", key, resp.StatusCode, len(body))

	// 备份是 JSONL（一行一个 key）：把 key 名、TTL、负载大小列出来，便于比对线上 Redis
	var shown int
	for _, line := range strings.Split(strings.TrimSpace(string(body)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e struct {
			K       string `json:"k"`
			TTLms   int64  `json:"t"`
			Payload string `json:"p"`
		}
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			fmt.Printf("  非备份格式行（%d 字节）：%s\n", len(line), trimRunes(line, 90))
			shown++
			continue
		}
		ttl := "永久"
		if e.TTLms > 0 {
			ttl = fmt.Sprintf("剩 %dm", e.TTLms/60000)
		}
		fmt.Printf("  %-46s %-6s payload=%dB\n", trimRunes(e.K, 46), ttl, len(e.Payload))
		shown++
	}
	fmt.Printf("  合计 %d 条记录\n", shown)
}

// copyObject 桶内复制。SDK 的源格式是“桶名(含APPID)/对象键”，且参数顺序是目标在前。
// 注意：COS 可能把错误嵌在 200 响应体里，SDK 已处理，但 ETag 为空仍要当失败看。
func copyObject(ctx context.Context, client *cos.Client, bucket, src, dst string) {
	res, resp, err := client.Object.Copy(ctx, dst, bucket+"/"+src, nil)
	if err != nil {
		fmt.Printf("COPY %s -> %s 失败：%v\n", src, dst, err)
		return
	}
	if resp != nil {
		fmt.Printf("COPY %s -> %s：HTTP %d，ETag=%s\n", src, dst, resp.StatusCode, res.ETag)
	}
	checkObject(ctx, client, dst)
}

// saveObject 下载到本地留底。备份被多方互相覆盖时，本地留底是唯一可靠证据。
func saveObject(ctx context.Context, client *cos.Client, key, path string) {
	resp, err := client.Object.Get(ctx, key, nil)
	if err != nil {
		fmt.Printf("SAVE %s 失败：%v\n", key, err)
		return
	}
	defer resp.Body.Close()
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("创建 %s 失败：%v\n", path, err)
		return
	}
	defer f.Close()
	n, cerr := io.Copy(f, resp.Body)
	if cerr != nil {
		fmt.Printf("SAVE %s -> %s：%d 字节后读取中断：%v\n", key, path, n, cerr)
		return
	}
	fmt.Printf("SAVE %s -> %s：%d 字节\n", key, path, n)
}

func trimRunes(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func ownerID(o *cos.Owner) string {
	if o == nil {
		return "(未知)"
	}
	return o.ID
}

func ownerUIN(o *cos.Owner) string {
	if o == nil || o.UIN == "" {
		return "(未知)"
	}
	return o.UIN
}
