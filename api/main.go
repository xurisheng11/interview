package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"interview-sim/config"
	"interview-sim/model"
	"interview-sim/repository"
	"interview-sim/router"
)

func main() {
	// 初始化配置
	config.Init()

	// 设置 Gin 模式
	gin.SetMode(config.Cfg.GinMode)

	// 初始化 Redis
	if err := repository.InitRedis(); err != nil {
		log.Fatalf("Redis 连接失败: %v", err)
	}
	log.Println("Redis 连接成功")

	// 初始化默认管理员账号
	if err := initDefaultAdmin(); err != nil {
		log.Printf("初始化管理员账号失败: %v", err)
	}

	// 启动路由
	r := router.Setup()
	addr := fmt.Sprintf(":%s", config.Cfg.ServerPort)
	log.Printf("服务启动，监听 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// initDefaultAdmin 初始化默认管理员账号（若不存在则创建）
func initDefaultAdmin() error {
	adminUsername := config.Cfg.AdminUsername
	// 检查管理员账号是否已存在
	existing, err := repository.GetUserByAccount(adminUsername)
	if err != nil {
		return err
	}
	if existing != nil {
		// 已存在，确保 role 是 admin
		if existing.Role != "admin" {
			if err := repository.UpdateUserField(existing.UserID, "role", "admin"); err != nil {
				return err
			}
			log.Printf("管理员账号 [%s] 已升级为 admin 角色", adminUsername)
		} else {
			log.Printf("管理员账号 [%s] 已存在，跳过初始化", adminUsername)
		}
		return nil
	}

	// 创建管理员账号
	hash, err := bcrypt.GenerateFromPassword([]byte(config.Cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &model.User{
		UserID:       uuid.New().String(),
		Username:     adminUsername,
		PasswordHash: string(hash),
		Nickname:     "系统管理员",
		Avatar:       "",
		Bio:          "",
		CreatedAt:    time.Now(),
		Role:         "admin",
	}

	if err := repository.SaveUser(admin); err != nil {
		return err
	}
	if err := repository.SaveAccountIndex(adminUsername, admin.UserID); err != nil {
		return err
	}
	// 管理员加入用户列表
	if err := repository.AddUserToList(admin.UserID, float64(admin.CreatedAt.Unix())); err != nil {
		return err
	}

	log.Printf("默认管理员账号创建成功: 用户名=%s，请及时修改默认密码", adminUsername)
	return nil
}
