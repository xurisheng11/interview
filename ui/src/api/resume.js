import request from './request'

// 上传并解析简历（FormData）
export function uploadResume(formData) {
  return request.post('/resumes', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 300000  // 5 分钟，大文件OCR处理需要更长时间
  })
}

// 获取简历列表
export function getResumeList() {
  return request.get('/resumes')
}

// 获取单条简历详情
export function getResume(id) {
  return request.get(`/resumes/${id}`)
}

// 删除简历
export function deleteResume(id) {
  return request.delete(`/resumes/${id}`)
}

// 重新分析简历
export function retryAnalyze(id) {
  return request.post(`/resumes/${id}/retry`)
}

// 基于简历创建面试
export function createResumeInterview(id) {
  return request.post(`/resumes/${id}/interview`)
}
