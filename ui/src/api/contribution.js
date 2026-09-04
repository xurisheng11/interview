import request from './request'

// 提交面试题贡献
export function submitContributedQuestion(data) {
  return request.post('/contributions/questions', data)
}

// 获取用户贡献记录和积分
export function getUserContributions() {
  return request.get('/contributions/me')
}

// 获取用户积分
export function getUserCredits() {
  return request.get('/contributions/credits')
}

// 扣除用户积分（查看已验证题目时）
export function deductCredits(amount = 1) {
  return request.post('/contributions/deduct', { amount })
}

// 获取公司贡献的题目
export function getCompanyContributedQuestions(company, jobTitle) {
  return request.get(`/contributions/company/${encodeURIComponent(company)}/questions`, {
    params: { jobTitle }
  })
}

// 获取单道题目详情
export function getContributedQuestion(id) {
  return request.get(`/contributions/questions/${id}`)
}

// 对题目投票
export function voteContributedQuestion(id, vote) {
  return request.post(`/contributions/questions/${id}/vote`, { vote })
}

// 举报题目
export function reportContributedQuestion(id, reason) {
  return request.post(`/contributions/questions/${id}/report`, { reason })
}

// 搜索公司题目的降级方案
export function searchQuestionsFallback(company, jobTitle) {
  return request.get('/contributions/search', {
    params: { company, jobTitle }
  })
}
