package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"interview-sim/config"
)

// 腾讯云一句话识别（ASR）：录音结束后整段上传同步转文字
// 接口文档：SentenceRecognition / Version 2019-06-14
const asrHost = "asr.tencentcloudapi.com"

// TranscribeAudio 将音频数据转写为文字（一句话识别：音频 ≤60 秒、≤1MB 原始数据）
func TranscribeAudio(audio []byte, ext string) (string, error) {
	if config.Cfg.TencentSecretID == "" || config.Cfg.TencentSecretKey == "" {
		return "", errors.New("语音识别未配置（缺少 TENCENT_SECRET_ID / TENCENT_SECRET_KEY 环境变量）")
	}
	if len(audio) == 0 {
		return "", errors.New("音频数据为空")
	}

	payload := map[string]interface{}{
		"ProjectId":      0,
		"SubServiceType": 2, // 一句话识别
		"EngSerViceType": "16k_zh",
		"SourceType":     1, // base64 数据
		"VoiceFormat":    voiceFormatOf(ext),
		"UsrAudioKey":    fmt.Sprintf("interview-%d", time.Now().UnixNano()),
		"Data":           base64.StdEncoding.EncodeToString(audio),
		"FilterModal":    1, // 过滤语气词
		"FilterPunc":     0, // 保留标点
		"ConvertNumMode": 1, // 数字转写
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, "https://"+asrHost, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	signTC3(req, body)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("语音识别服务请求失败：" + err.Error())
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("语音识别响应读取失败")
	}

	var out struct {
		Response struct {
			Result string `json:"Result"`
			Error  *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", errors.New("语音识别响应解析失败")
	}
	if out.Response.Error != nil {
		return "", fmt.Errorf("语音识别失败：%s", out.Response.Error.Message)
	}
	return out.Response.Result, nil
}

// voiceFormatOf 音频后缀 → 腾讯云 VoiceFormat 枚举（1 pcm / 2 wav / 3 mp3 / 4 speex / 5 amr / 6 m4a）
func voiceFormatOf(ext string) int {
	switch strings.ToLower(strings.TrimPrefix(ext, ".")) {
	case "pcm":
		return 1
	case "wav":
		return 2
	case "speex":
		return 4
	case "amr":
		return 5
	case "m4a", "aac":
		return 6
	default:
		return 3 // 小程序录音默认 mp3
	}
}

// signTC3 写入 TC3-HMAC-SHA256 签名头
func signTC3(req *http.Request, body []byte) {
	ts := time.Now().Unix()
	date := time.Unix(ts, 0).UTC().Format("2006-01-02")

	bodyHash := sha256.Sum256(body)
	canonical := "POST\n/\n\n" +
		"content-type:application/json; charset=utf-8\nhost:" + asrHost + "\n\n" +
		"content-type;host\n" + hex.EncodeToString(bodyHash[:])
	canonicalHash := sha256.Sum256([]byte(canonical))

	scope := date + "/asr/tc3_request"
	stringToSign := "TC3-HMAC-SHA256\n" + fmt.Sprint(ts) + "\n" + scope + "\n" + hex.EncodeToString(canonicalHash[:])

	secretDate := hmacSHA256([]byte("TC3"+config.Cfg.TencentSecretKey), date)
	secretService := hmacSHA256(secretDate, "asr")
	secretSigning := hmacSHA256(secretService, "tc3_request")
	signature := hex.EncodeToString(hmacSHA256(secretSigning, stringToSign))

	req.Header.Set("Authorization", fmt.Sprintf(
		"TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=content-type;host, Signature=%s",
		config.Cfg.TencentSecretID, scope, signature))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", asrHost)
	req.Header.Set("X-TC-Action", "SentenceRecognition")
	req.Header.Set("X-TC-Version", "2019-06-14")
	req.Header.Set("X-TC-Timestamp", fmt.Sprint(ts))
	req.Header.Set("X-TC-Region", "ap-guangzhou")
}

func hmacSHA256(key []byte, msg string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(msg))
	return h.Sum(nil)
}
