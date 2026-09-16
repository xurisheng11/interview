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
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("GET %s 内容读取失败：%v\n", key, err)
		return
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	fmt.Printf("GET %s：HTTP %d，%d 字节，%d 条记录\n", key, resp.StatusCode, len(data), len(lines))
	for i, ln := range lines {
		if i >= 3 {
			fmt.Printf("    ... 其余 %d 条省略\n", len(lines)-3)
			break
		}
		if r := []rune(ln); len(r) > 110 {
			ln = string(r[:110]) + "…"
		}
		fmt.Printf("    %s\n", ln)
	}
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
