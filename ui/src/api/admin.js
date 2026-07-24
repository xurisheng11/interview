import adminRequest from './adminRequest'
import request from './request'

// 管理员登录（复用 auth 接口，但用普通 request 发请求，不需要 token）
export function adminLogin(data) {
  return request.post('/auth/login', data)
}

// 获取所有用户列表
export function listUsers() {
  return adminRequest.get('/admin/users')
}

// 获取单个用户详情
export function getUser(userId) {
  return adminRequest.get(`/admin/users/${userId}`)
}

// 重置用户密码
export function resetUserPassword(userId, newPassword) {
  return adminRequest.put(`/admin/users/${userId}/password`, { newPassword })
}

// 修改用户角色
export function setUserRole(userId, role) {
  return adminRequest.put(`/admin/users/${userId}/role`, { role })
}

// 删除用户
export function deleteUser(userId) {
  return adminRequest.delete(`/admin/users/${userId}`)
}

// 一次性迁移老用户数据
export function migrateUsers() {
  return adminRequest.post('/admin/migrate-users')
}
