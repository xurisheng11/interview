package model

import "time"

type User struct {
	UserID       string    `json:"userId"`
	Username     string    `json:"username"`
	Phone        string    `json:"phone"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"passwordHash"`
	Avatar       string    `json:"avatar"`
	Nickname     string    `json:"nickname"`
	Bio          string    `json:"bio"`
	CreatedAt    time.Time `json:"createdAt"`
	Role         string    `json:"role"`
	LastLoginAt  string    `json:"lastLoginAt"`
}

func (u *User) ToRedisHash() map[string]interface{} {
	return map[string]interface{}{
		"userId":       u.UserID,
		"username":     u.Username,
		"phone":        u.Phone,
		"email":        u.Email,
		"passwordHash": u.PasswordHash,
		"avatar":       u.Avatar,
		"nickname":     u.Nickname,
		"bio":          u.Bio,
		"createdAt":    u.CreatedAt.Format(time.RFC3339),
		"role":         u.Role,
		"lastLoginAt":  u.LastLoginAt,
	}
}

func UserFromRedisHash(h map[string]string) *User {
	createdAt, _ := time.Parse(time.RFC3339, h["createdAt"])
	return &User{
		UserID:       h["userId"],
		Username:     h["username"],
		Phone:        h["phone"],
		Email:        h["email"],
		PasswordHash: h["passwordHash"],
		Avatar:       h["avatar"],
		Nickname:     h["nickname"],
		Bio:          h["bio"],
		CreatedAt:    createdAt,
		Role:         h["role"],
		LastLoginAt:  h["lastLoginAt"],
	}
}

// UserDTO 对外暴露（不含密码）
type UserDTO struct {
	UserID      string `json:"userId"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Bio         string `json:"bio"`
	Role        string `json:"role"`
	CreatedAt   string `json:"createdAt"`
	LastLoginAt string `json:"lastLoginAt"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}

func (u *User) ToDTO() *UserDTO {
	return &UserDTO{
		UserID:      u.UserID,
		Username:    u.Username,
		Nickname:    u.Nickname,
		Avatar:      u.Avatar,
		Bio:         u.Bio,
		Role:        u.Role,
		CreatedAt:   u.CreatedAt.Format(time.RFC3339),
		LastLoginAt: u.LastLoginAt,
	}
}

// ToAdminDTO 管理员视图（含手机/邮箱）
func (u *User) ToAdminDTO() *UserDTO {
	return &UserDTO{
		UserID:      u.UserID,
		Username:    u.Username,
		Nickname:    u.Nickname,
		Avatar:      u.Avatar,
		Bio:         u.Bio,
		Role:        u.Role,
		CreatedAt:   u.CreatedAt.Format(time.RFC3339),
		LastLoginAt: u.LastLoginAt,
		Phone:       u.Phone,
		Email:       u.Email,
	}
}
