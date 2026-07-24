import request from './request'

// 获取个人信息
export function getProfile() {
  return request.get('/profile')
}

// 更新个人信息
export function updateProfile(data) {
  return request.put('/profile', data)
}

// 修改密码
export function changePassword(data) {
  return request.put('/profile/password', data)
}

// 个人统计数据
export function getStats() {
  return request.get('/profile/stats')
}

// 得分趋势（近30次）
export function getScoreTrend() {
  return request.get('/profile/trend')
}

// 我的收藏
export function getCollections() {
  return request.get('/profile/collections')
}
