import request from './request'

// 创建面试（发起面试）
export function createInterview(data) {
  return request.post('/interviews', data)
}

// 获取面试列表
export function getInterviewList(params) {
  return request.get('/interviews', { params })
}

// 获取单个面试详情
export function getInterview(id) {
  return request.get(`/interviews/${id}`)
}

// 提交答案
export function submitAnswer(id, data) {
  return request.post(`/interviews/${id}/answers`, data)
}

// 暂停面试
export function pauseInterview(id) {
  return request.put(`/interviews/${id}/pause`)
}

// 完成面试（结束）
export function completeInterview(id) {
  return request.put(`/interviews/${id}/complete`)
}
