import request from './request'

// 获取面试报告
export function getReport(interviewId) {
  return request.get(`/reports/${interviewId}`)
}

// 生成分享链接
export function createShareLink(interviewId) {
  return request.post(`/reports/${interviewId}/share`)
}

// 获取分享报告（无需鉴权）
export function getSharedReport(token) {
  return request.get(`/reports/share/${token}`)
}
