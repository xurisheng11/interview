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
	JobStatus    string    `json:"jobStatus"`    // 求职状态：job_hunting, job_offered, employed_looking, employed_stable, intern, fresh_graduate
	Experience   string    `json:"experience"`  // 工作经验年限：fresh, 0-1, 1-3, 3-5, 5-10, 10+
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
		"jobStatus":    u.JobStatus,
		"experience":   u.Experience,
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
		JobStatus:    h["jobStatus"],
		Experience:   h["experience"],
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
	JobStatus   string `json:"jobStatus"`  // 求职状态
	Experience  string `json:"experience"` // 工作经验年限
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
		JobStatus:   u.JobStatus,
		Experience:  u.Experience,
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
		JobStatus:   u.JobStatus,
		Experience:  u.Experience,
	}
}
