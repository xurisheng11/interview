import request from './request'

// 文章列表
export function getArticles(params) {
  return request.get('/community/articles', { params })
}

// AI 生成文章
export function generateAIArticle(data) {
  return request.post('/community/articles/ai', data)
}

// 文章详情
export function getArticle(id) {
  return request.get(`/community/articles/${id}`)
}

// 点赞文章
export function likeArticle(id) {
  return request.post(`/community/articles/${id}/like`)
}

// 收藏文章
export function collectArticle(id) {
  return request.post(`/community/articles/${id}/collect`)
}

// 评论列表
export function getComments(id) {
  return request.get(`/community/articles/${id}/comments`)
}

// 发表评论
export function addComment(id, data) {
  return request.post(`/community/articles/${id}/comments`, data)
}
