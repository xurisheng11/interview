package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort             string
	GinMode                string
	JWTSecret              string
	JWTExpireDays          int
	RedisHost              string
	RedisPort              string
	RedisPassword          string
	RedisDB                int
	DeepSeekAPIKey         string
	DeepSeekBaseURL        string
	AIDailyLimit           int
	InterviewQuestionCount int
}

var Cfg *Config

func Init() {
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，使用系统环境变量")
	}

	jwtExpireDays, _ := strconv.Atoi(getEnv("JWT_EXPIRE_DAYS", "7"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "1"))
	aiDailyLimit, _ := strconv.Atoi(getEnv("AI_DAILY_LIMIT", "10"))
	questionCount, _ := strconv.Atoi(getEnv("INTERVIEW_QUESTION_COUNT", "10"))

	Cfg = &Config{
		ServerPort:             getEnv("SERVER_PORT", "8080"),
		GinMode:                getEnv("GIN_MODE", "debug"),
		JWTSecret:              getEnv("JWT_SECRET", "interview-sim-secret"),
		JWTExpireDays:          jwtExpireDays,
		RedisHost:              getEnv("REDIS_HOST", "127.0.0.1"),
		RedisPort:              getEnv("REDIS_PORT", "6379"),
		RedisPassword:          getEnv("REDIS_PASSWORD", ""),
		RedisDB:                redisDB,
		DeepSeekAPIKey:         getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL:        getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		AIDailyLimit:           aiDailyLimit,
		InterviewQuestionCount: questionCount,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
