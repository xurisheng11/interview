package repository

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	"interview-sim/model"
)

const userKeyPrefix = "user:"
const userAccountKeyPrefix = "user:account:"

func userKey(userID string) string     { return userKeyPrefix + userID }
func accountKey(account string) string { return userAccountKeyPrefix + account }

// SaveUser 永久保存用户信息到 Redis Hash
func SaveUser(user *model.User) error {
	key := userKey(user.UserID)
	hash := user.ToRedisHash()
	// 使用 HMSet，它接受 map[string]interface{} 且语义稳定
	if err := RDB.HMSet(Ctx, key, hash).Err(); err != nil {
		return fmt.Errorf("SaveUser HSet: %w", err)
	}
	// 永久保存，移除过期时间
	return Persist(key)
}

// SaveAccountIndex 保存 account -> userId 索引（永久）
func SaveAccountIndex(account, userID string) error {
	return SetPermanent(accountKey(account), userID)
}

// GetUserByID 通过 userId 获取用户
func GetUserByID(userID string) (*model.User, error) {
	hash, err := HGetAll(userKey(userID))
	if err != nil {
		return nil, err
	}
	if len(hash) == 0 {
		return nil, nil
	}
	return model.UserFromRedisHash(hash), nil
}

// GetUserByAccount 通过 account (手机/邮箱/用户名) 获取用户
func GetUserByAccount(account string) (*model.User, error) {
	userID, err := Get(accountKey(account))
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return GetUserByID(userID)
}

// AccountExists 检查 account 是否已被注册
func AccountExists(account string) (bool, error) {
	return Exists(accountKey(account))
}

// UpdateUserField 更新用户单个字段
func UpdateUserField(userID, field, value string) error {
	return RDB.HSet(Ctx, userKey(userID), field, value).Err()
}
