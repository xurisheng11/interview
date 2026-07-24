package service

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"interview-sim/model"
	"interview-sim/pkg/jwt"
	"interview-sim/repository"
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
	Account         string `json:"account" binding:"required"`
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
	// account 格式校验（手机号或邮箱）
	if !isValidAccount(req.Account) {
		return nil, ErrInvalidAccount
	}
	// 检查 account 唯一性
	exists, err := repository.AccountExists(req.Account)
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
	if isEmail(req.Account) {
		user.Email = req.Account
	} else {
		user.Phone = req.Account
	}

	// 保存到 Redis
	if err := repository.SaveUser(user); err != nil {
		return nil, err
	}
	if err := repository.SaveAccountIndex(req.Account, user.UserID); err != nil {
		return nil, err
	}
	// 同时保存用户名索引
	if err := repository.SaveAccountIndex(req.Username, user.UserID); err != nil {
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
