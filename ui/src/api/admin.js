import request from './request'

// 获取所有用户列表
export function listUsers() {
  return request.get('/admin/users')
}

// 获取单个用户详情
export function getUser(userId) {
  return request.get(`/admin/users/${userId}`)
}

// 重置用户密码
export function resetUserPassword(userId, newPassword) {
  return request.put(`/admin/users/${userId}/password`, { newPassword })
}

// 修改用户角色
export function setUserRole(userId, role) {
  return request.put(`/admin/users/${userId}/role`, { role })
}

// 删除用户
export function deleteUser(userId) {
  return request.delete(`/admin/users/${userId}`)
}
