package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"interview-sim/model"
	"interview-sim/repository"
)

var (
	ErrUserNotFound    = errors.New("用户不存在")
	ErrCannotEditAdmin = errors.New("不能修改管理员账号")
)

type UserListResult struct {
	Total int              `json:"total"`
	List  []*model.UserDTO `json:"list"`
}

// AdminListUsers 获取所有用户列表
func AdminListUsers() (*UserListResult, error) {
	userIDs, err := repository.GetAllUserIDs()
	if err != nil {
		return nil, err
	}

	list := make([]*model.UserDTO, 0, len(userIDs))
	for _, uid := range userIDs {
		user, err := repository.GetUserByID(uid)
		if err != nil || user == nil {
			continue
		}
		list = append(list, user.ToAdminDTO())
	}

	return &UserListResult{
		Total: len(list),
		List:  list,
	}, nil
}

type AdminResetPasswordReq struct {
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

// AdminResetPassword 管理员重置指定用户密码
func AdminResetPassword(targetUserID string, req *AdminResetPasswordReq) error {
	user, err := repository.GetUserByID(targetUserID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return repository.UpdateUserField(targetUserID, "passwordHash", string(hash))
}

// AdminGetUser 获取单个用户详情（管理员视图）
func AdminGetUser(userID string) (*model.UserDTO, error) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user.ToAdminDTO(), nil
}

// AdminSetRole 修改用户角色
func AdminSetRole(targetUserID string, role string) error {
	if role != "user" && role != "admin" {
		return errors.New("角色值无效，只允许 user 或 admin")
	}
	user, err := repository.GetUserByID(targetUserID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	return repository.UpdateUserField(targetUserID, "role", role)
}

// AdminDeleteUser 删除用户（不允许删除管理员自身）
func AdminDeleteUser(operatorID, targetUserID string) error {
	if operatorID == targetUserID {
		return errors.New("不能删除自己的账号")
	}
	user, err := repository.GetUserByID(targetUserID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	return repository.DeleteUser(user)
}
