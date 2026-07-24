import request from './request'

// 查询公司面试情报
export function getCompanyIntel(company, jobTitle) {
  return request.get('/company/intel', { params: { company, jobTitle } })
}

// 获取单道题目的参考答案
export function getCompanyQuestionAnswer(company, jobTitle, content) {
  return request.get('/company/question-answer', { params: { company, jobTitle, content } })
}
