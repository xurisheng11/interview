import request from './request'

// 获取题库列表（支持筛选分页）
export function getQuestions(params) {
  return request.get('/questions', { params })
}

// 获取题目详情
export function getQuestion(id) {
  return request.get(`/questions/${id}`)
}

// 单题练习（AI点评）
export function practiceQuestion(id, data) {
  return request.post(`/questions/${id}/practice`, data)
}

// 收藏/取消收藏题目
export function collectQuestion(id) {
  return request.post(`/questions/${id}/collect`)
}

// 创建题目（管理员）
export function createQuestion(data) {
  return request.post('/questions', data)
}

// 编辑题目（管理员）
export function updateQuestion(id, data) {
  return request.put(`/questions/${id}`, data)
}

// 删除题目（管理员）
export function deleteQuestion(id) {
  return request.delete(`/questions/${id}`)
}
