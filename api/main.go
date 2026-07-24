package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"interview-sim/config"
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

	// 启动路由
	r := router.Setup()
	addr := fmt.Sprintf(":%s", config.Cfg.ServerPort)
	log.Printf("服务启动，监听 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
