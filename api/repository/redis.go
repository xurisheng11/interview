package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"interview-sim/config"
)

var RDB *redis.Client
var Ctx = context.Background()

func InitRedis() error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Cfg.RedisHost, config.Cfg.RedisPort),
		Password: config.Cfg.RedisPassword,
		DB:       config.Cfg.RedisDB,
	})
	_, err := RDB.Ping(Ctx).Result()
	return err
}

// Set 设置字符串值
func Set(key string, value interface{}, ttl time.Duration) error {
	return RDB.Set(Ctx, key, value, ttl).Err()
}

// SetPermanent 永久设置（无过期时间）
func SetPermanent(key string, value interface{}) error {
	return RDB.Set(Ctx, key, value, 0).Err()
}

// Get 获取字符串值
func Get(key string) (string, error) {
	return RDB.Get(Ctx, key).Result()
}

// HSet 设置 Hash 单个或少量字段，用法：HSet(key, "field1", val1, "field2", val2)
func HSet(key string, fieldValues ...interface{}) error {
	return RDB.HSet(Ctx, key, fieldValues...).Err()
}

// HSetMap 通过 map 批量设置 Hash 字段，使用 HMSet 确保兼容性
func HSetMap(key string, values map[string]interface{}) error {
	return RDB.HMSet(Ctx, key, values).Err()
}

// HGet 获取 Hash 单个字段
func HGet(key, field string) (string, error) {
	return RDB.HGet(Ctx, key, field).Result()
}

// HGetAll 获取 Hash 全部字段
func HGetAll(key string) (map[string]string, error) {
	return RDB.HGetAll(Ctx, key).Result()
}

// Del 删除 key
func Del(keys ...string) error {
	return RDB.Del(Ctx, keys...).Err()
}

// Expire 设置过期时间
func Expire(key string, ttl time.Duration) error {
	return RDB.Expire(Ctx, key, ttl).Err()
}

// Persist 移除过期时间（永久保存）
func Persist(key string) error {
	return RDB.Persist(Ctx, key).Err()
}

// Exists 检查 key 是否存在
func Exists(key string) (bool, error) {
	n, err := RDB.Exists(Ctx, key).Result()
	return n > 0, err
}

// Incr 自增并返回新值
func Incr(key string) (int64, error) {
	return RDB.Incr(Ctx, key).Result()
}

// IncrBy 按量自增
func IncrBy(key string, value int64) (int64, error) {
	return RDB.IncrBy(Ctx, key, value).Result()
}

// SAdd 添加到 Set
func SAdd(key string, members ...interface{}) error {
	return RDB.SAdd(Ctx, key, members...).Err()
}

// SRem 从 Set 移除
func SRem(key string, members ...interface{}) error {
	return RDB.SRem(Ctx, key, members...).Err()
}

// SIsMember 检查是否在 Set 中
func SIsMember(key string, member interface{}) (bool, error) {
	return RDB.SIsMember(Ctx, key, member).Result()
}

// SMembers 获取 Set 所有成员
func SMembers(key string) ([]string, error) {
	return RDB.SMembers(Ctx, key).Result()
}

// ZAdd 添加到有序集合
func ZAdd(key string, score float64, member string) error {
	return RDB.ZAdd(Ctx, key, &redis.Z{Score: score, Member: member}).Err()
}

// ZRevRange 有序集合倒序获取（高分在前）
func ZRevRange(key string, start, stop int64) ([]string, error) {
	return RDB.ZRevRange(Ctx, key, start, stop).Result()
}

// ExpireAt 设置在某时刻过期（用于每日限额）
func ExpireAt(key string, t time.Time) error {
	return RDB.ExpireAt(Ctx, key, t).Err()
}

// ZRem 从有序集合移除成员
func ZRem(key string, members ...interface{}) error {
	return RDB.ZRem(Ctx, key, members...).Err()
}

// ZRange 有序集合正序获取（低分在前）
func ZRange(key string, start, stop int64) ([]string, error) {
	return RDB.ZRange(Ctx, key, start, stop).Result()
}

// HIncrBy Hash 字段整数自增
func HIncrBy(key, field string, incr int64) (int64, error) {
	return RDB.HIncrBy(Ctx, key, field, incr).Result()
}

// LPush 从列表头部插入一个或多个值
func LPush(key string, values ...interface{}) error {
	return RDB.LPush(Ctx, key, values...).Err()
}

// LRange 获取列表指定范围的元素
func LRange(key string, start, stop int64) ([]string, error) {
	return RDB.LRange(Ctx, key, start, stop).Result()
}

// ZIncrBy 有序集合成员分数自增
func ZIncrBy(key string, increment float64, member string) error {
	return RDB.ZIncrBy(Ctx, key, increment, member).Err()
}
