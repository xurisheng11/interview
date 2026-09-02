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

// 搜索公司
export function searchCompanies(query) {
  return request.get('/companies/search', { params: { q: query } })
}

// 获取公司面试题
export function getCompanyQuestions(companyId, params) {
  return request.get(`/companies/${companyId}/questions`, { params })
}

// 提交面试题（用户贡献）
export function submitInterviewQuestion(data) {
  return request.post('/questions/contribute', data)
}

// 获取用户贡献的面试题
export function getMyContributions() {
  return request.get('/questions/my-contributions')
}
