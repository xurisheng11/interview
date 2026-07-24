package repository

import (
	"fmt"
	"time"

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

// AddUserToList 将用户加入全局用户列表（有序集合，按注册时间排序）
func AddUserToList(userID string, score float64) error {
	return RDB.ZAdd(Ctx, "users:all", &redis.Z{Score: score, Member: userID}).Err()
}

// GetAllUserIDs 获取所有用户ID列表（按注册时间倒序）
func GetAllUserIDs() ([]string, error) {
	return RDB.ZRevRange(Ctx, "users:all", 0, -1).Result()
}

// GetUserCount 获取用户总数
func GetUserCount() (int64, error) {
	return RDB.ZCard(Ctx, "users:all").Result()
}

// SaveLastLogin 保存用户最后登录时间
func SaveLastLogin(userID string, loginTime string) error {
	return RDB.HSet(Ctx, userKey(userID), "lastLoginAt", loginTime).Err()
}

// MigrateUsersToList 一次性迁移：把 Redis 中所有现有用户补录进 users:all
// 通过 SCAN 找到所有 user:{id} hash key（排除 user:account:* 索引）
func MigrateUsersToList() (int, error) {
	var cursor uint64
	count := 0
	for {
		keys, nextCursor, err := RDB.Scan(Ctx, cursor, userKeyPrefix+"*", 100).Result()
		if err != nil {
			return count, err
		}
		for _, key := range keys {
			// 跳过账号索引 key（user:account:xxx）
			if len(key) > len(userAccountKeyPrefix) && key[:len(userAccountKeyPrefix)] == userAccountKeyPrefix {
				continue
			}
			// 取出用户 hash 里的 createdAt 和 userId
			vals, err := RDB.HMGet(Ctx, key, "userId", "createdAt").Result()
			if err != nil || vals[0] == nil {
				continue
			}
			userID := vals[0].(string)
			if userID == "" {
				continue
			}
			// 解析 createdAt 作为 score，解析失败则用 0
			score := float64(0)
			if vals[1] != nil {
				if t, err := time.Parse(time.RFC3339, vals[1].(string)); err == nil {
					score = float64(t.Unix())
				}
			}
			// ZAdd NX：只在不存在时插入，不覆盖已有记录
			RDB.ZAddNX(Ctx, "users:all", &redis.Z{Score: score, Member: userID})
			count++
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return count, nil
}
// 需要传入完整用户对象以清理所有账号索引
func DeleteUser(user *model.User) error {
	// 1. 删除用户数据 + 所有账号索引（用户名、手机、邮箱）
	keysToDelete := []string{userKey(user.UserID)}
	if user.Username != "" {
		keysToDelete = append(keysToDelete, accountKey(user.Username))
	}
	if user.Phone != "" {
		keysToDelete = append(keysToDelete, accountKey(user.Phone))
	}
	if user.Email != "" {
		keysToDelete = append(keysToDelete, accountKey(user.Email))
	}
	if err := RDB.Del(Ctx, keysToDelete...).Err(); err != nil {
		return fmt.Errorf("DeleteUser Del: %w", err)
	}
	// 2. 从全局用户列表中移除
	return RDB.ZRem(Ctx, "users:all", user.UserID).Err()
}
