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
	AdminUsername          string
	AdminPassword          string
	MySQLHost              string
	MySQLPort              string
	MySQLUser              string
	MySQLPassword          string
	MySQLDB                string
	MySQLSyncHour          int
	MySQLEnabled           bool
	WXAppID                string
	WXSecret               string
	TencentSecretID        string
	TencentSecretKey       string
	// 对象存储备份（云托管容器文件系统临时，靠 COS 跨版本/跨冷启动保存数据）
	COSRegion         string
	COSBucket         string // 形如 mybucket-1250000000（桶名-APPID）
	COSPrefix         string
	COSSecretID       string
	COSSecretKey      string
	BackupIntervalMin int
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
	mysqlSyncHour, _ := strconv.Atoi(getEnv("MYSQL_SYNC_HOUR", "2"))
	mysqlEnabled, _ := strconv.ParseBool(getEnv("MYSQL_ENABLED", "true"))
	backupIntervalMin := backupInterval(getEnv("BACKUP_INTERVAL_MIN", "15"))

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
		AdminUsername:          getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:          getEnv("ADMIN_PASSWORD", "admin123456"),
		MySQLHost:              getEnv("MYSQL_HOST", "127.0.0.1"),
		MySQLPort:              getEnv("MYSQL_PORT", "3306"),
		MySQLUser:              getEnv("MYSQL_USER", "root"),
		MySQLPassword:          getEnv("MYSQL_PASSWORD", "admin"),
		MySQLDB:                getEnv("MYSQL_DB", "interview_sim"),
		MySQLSyncHour:          mysqlSyncHour,
		MySQLEnabled:           mysqlEnabled,
		WXAppID:                getEnv("WX_APP_ID", ""),
		WXSecret:               getEnv("WX_SECRET", ""),
		TencentSecretID:        getEnv("TENCENT_SECRET_ID", ""),
		TencentSecretKey:       getEnv("TENCENT_SECRET_KEY", ""),
		// COS 凭证默认复用语音识别那套腾讯云密钥，需单独授权时才用 COS_SECRET_* 覆盖
		COSRegion:    getEnv("COS_REGION", ""),
		COSBucket:    getEnv("COS_BUCKET", ""),
		COSPrefix:    getEnv("COS_PREFIX", "interview-backup"),
		COSSecretID:  getEnvOr("COS_SECRET_ID", "TENCENT_SECRET_ID", ""),
		COSSecretKey: getEnvOr("COS_SECRET_KEY", "TENCENT_SECRET_KEY", ""),
		// 备份频率：默认 15 分钟一次，最多丢一个周期
		BackupIntervalMin: backupIntervalMin,
	}
}

func backupInterval(raw string) int {
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 15
	}
	return n
}

// getEnvOr 优先取 key，其次取 fallbackKey，最后用默认值
func getEnvOr(key, fallbackKey, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	if val := os.Getenv(fallbackKey); val != "" {
		return val
	}
	return defaultVal
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
