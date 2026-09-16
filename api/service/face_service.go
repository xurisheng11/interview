package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"interview-sim/config"
	"interview-sim/model"
)

// 腾讯云人脸识别·人脸属性分析（DetectFaceAttributes）：视频面试抓帧识别情绪/微笑
// 接口文档：iai.tencentcloudapi.com / Version 2020-03-03
const iaiHost = "iai.tencentcloudapi.com"

// 情绪 Type → 中文名（与腾讯云 AttributeItem.Type 对齐）
var faceEmotionNames = []string{"自然", "高兴", "惊讶", "生气", "悲伤", "厌恶", "害怕"}

// DetectFaceEmotion 输入图片 base64，返回主脸情绪与微笑识别结果
func DetectFaceEmotion(imageBase64 string) (*model.FaceEmotionInfo, error) {
	if config.Cfg.TencentSecretID == "" || config.Cfg.TencentSecretKey == "" {
		return nil, errors.New("表情识别未配置（缺少 TENCENT_SECRET_ID / TENCENT_SECRET_KEY）")
	}
	if imageBase64 == "" {
		return nil, errors.New("图片数据为空")
	}

	payload := map[string]interface{}{
		"Image":              imageBase64,
		"FaceAttributesType": "Emotion,Smile",
		"MaxFaceNum":         1,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://"+iaiHost, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	signTC3IAI(req, body)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("表情识别服务请求失败：" + err.Error())
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("表情识别响应读取失败")
	}

	var out struct {
		Response struct {
			FaceDetailInfos []struct {
				FaceDetailAttributesInfo struct {
					Emotion struct {
						Type        int     `json:"Type"`
						Probability float64 `json:"Probability"`
					} `json:"Emotion"`
					Smile struct {
						Type        int     `json:"Type"`
						Probability float64 `json:"Probability"`
					} `json:"Smile"`
				} `json:"FaceDetailAttributesInfo"`
			} `json:"FaceDetailInfos"`
			Error *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, errors.New("表情识别响应解析失败")
	}
	if out.Response.Error != nil {
		return nil, fmt.Errorf("表情识别失败：%s（请确认腾讯云控制台已开通人脸识别服务）", out.Response.Error.Message)
	}
	if len(out.Response.FaceDetailInfos) == 0 {
		return nil, errors.New("画面中未检测到人脸")
	}
	attr := out.Response.FaceDetailInfos[0].FaceDetailAttributesInfo
	info := &model.FaceEmotionInfo{
		Type:        attr.Emotion.Type,
		Probability: attr.Emotion.Probability,
		Smile:       attr.Smile.Type,
		SmileProb:   attr.Smile.Probability,
	}
	if info.Type >= 0 && info.Type < len(faceEmotionNames) {
		info.Name = faceEmotionNames[info.Type]
	} else {
		info.Name = "未知"
	}
	return info, nil
}

// signTC3IAI 为 iai 服务写入 TC3-HMAC-SHA256 签名头（与 asr 同套算法，service/host/action 不同）
func signTC3IAI(req *http.Request, body []byte) {
	ts := time.Now().Unix()
	date := time.Unix(ts, 0).UTC().Format("2006-01-02")

	bodyHash := sha256.Sum256(body)
	canonical := "POST\n/\n\n" +
		"content-type:application/json; charset=utf-8\nhost:" + iaiHost + "\n\n" +
		"content-type;host\n" + hex.EncodeToString(bodyHash[:])
	canonicalHash := sha256.Sum256([]byte(canonical))

	scope := date + "/iai/tc3_request"
	stringToSign := "TC3-HMAC-SHA256\n" + fmt.Sprint(ts) + "\n" + scope + "\n" + hex.EncodeToString(canonicalHash[:])

	secretDate := hmacSHA256([]byte("TC3"+config.Cfg.TencentSecretKey), date)
	secretService := hmacSHA256(secretDate, "iai")
	secretSigning := hmacSHA256(secretService, "tc3_request")
	signature := hex.EncodeToString(hmacSHA256(secretSigning, stringToSign))

	req.Header.Set("Authorization", fmt.Sprintf(
		"TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=content-type;host, Signature=%s",
		config.Cfg.TencentSecretID, scope, signature))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", iaiHost)
	req.Header.Set("X-TC-Action", "DetectFaceAttributes")
	req.Header.Set("X-TC-Version", "2020-03-03")
	req.Header.Set("X-TC-Timestamp", fmt.Sprint(ts))
	req.Header.Set("X-TC-Region", "ap-guangzhou")
}
