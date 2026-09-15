package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"

	"interview-sim/config"
	"interview-sim/model"
	"interview-sim/pkg/jwt"
	"interview-sim/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAccountExists    = errors.New("该账号已被注册")
	ErrAccountNotFound  = errors.New("账号不存在")
	ErrWrongPassword    = errors.New("密码错误")
	ErrPasswordTooShort = errors.New("密码不能少于8位")
	ErrPasswordMismatch = errors.New("两次密码不一致")
	ErrInvalidAccount   = errors.New("手机号或邮箱格式不正确")
)

type RegisterReq struct {
	Username        string `json:"username" binding:"required,min=2,max=20"`
	Account         string `json:"account" binding:"omitempty"`
	Phone           string `json:"phone" binding:"omitempty"`
	Email           string `json:"email" binding:"omitempty"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

type LoginReq struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResult struct {
	Token     string         `json:"token"`
	ExpiresAt string         `json:"expiresAt"`
	User      *model.UserDTO `json:"user"`
}

func Register(req *RegisterReq) (*AuthResult, error) {
	// 密码校验
	if len(req.Password) < 8 {
		return nil, ErrPasswordTooShort
	}
	if req.Password != req.ConfirmPassword {
		return nil, ErrPasswordMismatch
	}
	// account 兼容两种入参：单一 account 字段，或分开的 phone / email 字段
	account := req.Account
	if account == "" {
		account = req.Email
	}
	if account == "" {
		account = req.Phone
	}
	// account 格式校验（手机号或邮箱）
	if !isValidAccount(account) {
		return nil, ErrInvalidAccount
	}
	// 检查 account 唯一性
	exists, err := repository.AccountExists(account)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAccountExists
	}
	// 检查用户名唯一性
	usernameExists, err := repository.AccountExists(req.Username)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, errors.New("用户名已被占用")
	}

	// 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		UserID:       uuid.New().String(),
		Username:     req.Username,
		PasswordHash: string(hash),
		Avatar:       "",
		Nickname:     req.Username,
		Bio:          "",
		CreatedAt:    time.Now(),
		Role:         "user",
	}

	// 判断 account 是手机号还是邮箱
	if isEmail(account) {
		user.Email = account
	} else {
		user.Phone = account
	}
	// 分开传了额外邮箱时一并登记索引（支持邮箱登录）
	extraEmail := ""
	if req.Email != "" && req.Email != account {
		if !isEmail(req.Email) {
			return nil, ErrInvalidAccount
		}
		emailExists, err := repository.AccountExists(req.Email)
		if err != nil {
			return nil, err
		}
		if emailExists {
			return nil, ErrAccountExists
		}
		extraEmail = req.Email
		user.Email = extraEmail
	}

	// 保存到 Redis
	if err := repository.SaveUser(user); err != nil {
		return nil, err
	}
	if err := repository.SaveAccountIndex(account, user.UserID); err != nil {
		return nil, err
	}
	// 同时保存用户名索引
	if err := repository.SaveAccountIndex(req.Username, user.UserID); err != nil {
		return nil, err
	}
	if extraEmail != "" {
		if err := repository.SaveAccountIndex(extraEmail, user.UserID); err != nil {
			return nil, err
		}
	}
	// 加入全局用户列表（按注册时间排序）
	if err := repository.AddUserToList(user.UserID, float64(user.CreatedAt.Unix())); err != nil {
		return nil, err
	}

	// 生成 JWT
	token, err := jwt.GenerateToken(user.UserID, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		Token:     token,
		ExpiresAt: time.Now().AddDate(0, 0, 7).Format(time.RFC3339),
		User:      user.ToDTO(),
	}, nil
}

func Login(req *LoginReq) (*AuthResult, error) {
	user, err := repository.GetUserByAccount(req.Account)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrAccountNotFound
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrWrongPassword
	}
	// 保存最后登录时间
	loginTime := time.Now().Format(time.RFC3339)
	_ = repository.SaveLastLogin(user.UserID, loginTime)
	user.LastLoginAt = loginTime

	token, err := jwt.GenerateToken(user.UserID, user.Role)
	if err != nil {
		return nil, err
	}
	return &AuthResult{
		Token:     token,
		ExpiresAt: time.Now().AddDate(0, 0, 7).Format(time.RFC3339),
		User:      user.ToDTO(),
	}, nil
}

func isValidAccount(account string) bool {
	return isEmail(account) || isPhone(account)
}

func isEmail(s string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(s)
}

func isPhone(s string) bool {
	re := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return re.MatchString(s)
}

// WxLoginReq 微信登录请求
type WxLoginReq struct {
	Code string `json:"code" binding:"required"` // 微信授权码
}

// WxLogin 微信登录/注册
func WxLogin(req *WxLoginReq) (*AuthResult, error) {
	// 1. 用 code 向微信服务器换取 openid
	openID, err := getWechatOpenID(req.Code)
	if err != nil {
		return nil, errors.New("微信授权失败: " + err.Error())
	}

	// 2. 检查是否已存在该 openid 的用户
	existingUser, err := repository.GetUserByOpenID(openID)
	if err != nil {
		return nil, err
	}

	var user *model.User
	if existingUser != nil {
		// 已存在，直接登录
		user = existingUser
	} else {
		// 不存在，自动注册新用户
		nickname := "用户" + uuid.New().String()[:8]
		user = &model.User{
			UserID:    uuid.New().String(),
			Username:  nickname,
			Avatar:    "",
			Nickname:  nickname,
			Bio:       "",
			OpenID:    openID,
			CreatedAt: time.Now(),
			Role:      "user",
		}
		// 保存到 Redis
		if err := repository.SaveUser(user); err != nil {
			return nil, err
		}
		// 保存 openid 索引
		if err := repository.SaveOpenIDIndex(openID, user.UserID); err != nil {
			return nil, err
		}
		// 加入全局用户列表
		if err := repository.AddUserToList(user.UserID, float64(user.CreatedAt.Unix())); err != nil {
			return nil, err
		}
	}

	// 3. 保存最后登录时间
	loginTime := time.Now().Format(time.RFC3339)
	_ = repository.SaveLastLogin(user.UserID, loginTime)
	user.LastLoginAt = loginTime

	// 4. 生成 JWT
	token, err := jwt.GenerateToken(user.UserID, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		Token:     token,
		ExpiresAt: time.Now().AddDate(0, 0, config.Cfg.JWTExpireDays).Format(time.RFC3339),
		User:      user.ToDTO(),
	}, nil
}

// getWechatOpenID 通过授权码向微信服务器换取 openid
func getWechatOpenID(code string) (string, error) {
	appID := config.Cfg.WXAppID
	secret := config.Cfg.WXSecret
	url := "https://api.weixin.qq.com/sns/jscode2session?appid=" + appID + "&secret=" + secret + "&js_code=" + code + "&grant_type=authorization_code"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if errMsg, ok := result["errmsg"]; ok {
		return "", errors.New(errMsg.(string))
	}

	openid, ok := result["openid"].(string)
	if !ok || openid == "" {
		return "", errors.New("未获取到 openid")
	}

	return openid, nil
}
