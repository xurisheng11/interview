// pages/community/detail.js
const api = require('../../api/index')
const { mdToHtml } = require('../../utils/markdown')

Page({
  data: {
    articleId: '',
    loading: true,
    article: null,
    commentsLoading: false,
    comments: [],
    commentText: '',
    submittingComment: false
  },

  onLoad(options) {
    if (options.id) {
      this.setData({ articleId: options.id })
      this.loadArticle()
    } else {
      wx.showToast({ title: '文章ID缺失', icon: 'none' })
      setTimeout(() => wx.navigateBack(), 1500)
    }
  },

  // 加载文章详情
  loadArticle() {
    this.setData({ loading: true })
    api.community.getArticle(this.data.articleId).then(res => {
      const article = res.data || res
      article._time = this.formatTime(article.createdAt)
      // Markdown 正文转 HTML，交给 rich-text 渲染（标题/表格/列表/加粗不再堆成一团乱码）
      article._html = mdToHtml(article.content || '')
      this.setData({ article, loading: false })
      this.loadComments()
    }).catch(err => {
      console.error('加载文章失败:', err)
      this.setData({ loading: false })
      wx.showToast({ title: '文章加载失败', icon: 'none' })
    })
  },

  // 加载评论
  loadComments() {
    this.setData({ commentsLoading: true })
    api.community.getComments(this.data.articleId).then(res => {
      const d = res.data || res
      const list = Array.isArray(d) ? d : (d.list || d.items || [])
      list.forEach(c => {
        c._time = this.formatTime(c.createdAt)
        c._id = c.commentId || c.id
        c._author = c.username || c.nickname || '匿名用户'
      })
      this.setData({ comments: list, commentsLoading: false })
    }).catch(err => {
      console.error('加载评论失败:', err)
      this.setData({ commentsLoading: false })
    })
  },

  // 点赞
  handleLike() {
    api.community.like(this.data.articleId).then(() => {
      const article = this.data.article
      if (article.liked) {
        article.likeCount = Math.max(0, (article.likeCount || 0) - 1)
      } else {
        article.likeCount = (article.likeCount || 0) + 1
      }
      article.liked = !article.liked
      this.setData({ article })
    }).catch(err => {
      console.error('点赞失败:', err)
      wx.showToast({ title: '操作失败', icon: 'none' })
    })
  },

  // 收藏 / 取消收藏（收藏走 POST、取消走 DELETE，后端非切换式接口）
  handleCollect() {
    const wasCollected = this.data.article.collected
    const req = wasCollected
      ? api.community.uncollect(this.data.articleId)
      : api.community.collect(this.data.articleId)
    req.then(() => {
      const article = this.data.article
      if (article.collected) {
        article.collectCount = Math.max(0, (article.collectCount || 0) - 1)
      } else {
        article.collectCount = (article.collectCount || 0) + 1
      }
      article.collected = !article.collected
      this.setData({ article })
      wx.showToast({
        title: article.collected ? '收藏成功' : '已取消收藏',
        icon: 'success'
      })
    }).catch(err => {
      console.error('收藏失败:', err)
      wx.showToast({ title: '操作失败', icon: 'none' })
    })
  },

  // 评论输入
  onCommentInput(e) {
    this.setData({ commentText: e.detail.value })
  },

  // 提交评论
  submitComment() {
    const text = this.data.commentText.trim()
    if (!text) return

    this.setData({ submittingComment: true })
    api.community.addComment(this.data.articleId, { content: text }).then(() => {
      wx.showToast({ title: '评论发表成功', icon: 'success' })
      this.setData({ commentText: '', submittingComment: false })
      this.loadComments()
      // 更新文章评论数
      const article = this.data.article
      if (article) {
        article.commentCount = (article.commentCount || 0) + 1
        this.setData({ article })
      }
    }).catch(err => {
      console.error('评论发表失败:', err)
      this.setData({ submittingComment: false })
      wx.showToast({ title: '评论发表失败', icon: 'none' })
    })
  },

  // 返回社区首页
  goBack() {
    wx.navigateBack({
      fail: () => {
        wx.redirectTo({ url: '/pages/community/index' })
      }
    })
  },

  // 格式化时间
  formatTime(val) {
    if (!val) return ''
    const timestamp = typeof val === 'number' && val < 10000000000 ? val * 1000 : val
    const d = new Date(timestamp)
    if (isNaN(d.getTime())) return val
    const pad = n => String(n).padStart(2, '0')
    return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) +
      ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes())
  },

  onShareAppMessage() {
    const article = this.data.article
    return {
      title: article ? article.title : '文章详情',
      path: '/pages/community/detail?id=' + this.data.articleId
    }
  }
})
