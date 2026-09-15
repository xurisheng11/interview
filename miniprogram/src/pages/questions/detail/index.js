const api = require('../../../api/index')

Page({
  data: {
    questionId: null,
    question: {},
    loading: true,
    error: null,
    isCollected: false
  },

  onLoad(options) {
    if (options.id) {
      this.setData({ questionId: options.id })
      this.loadQuestion()
    }
  },

  loadQuestion() {
    this.setData({ loading: true, error: null })
    
    api.question.get(this.data.questionId).then(res => {
      const question = res.data || {}
      question.typeName = this.getTypeName(question.type)
      question.difficultyName = this.getDifficultyName(question.difficulty)
      this.setData({ 
        question,
        isCollected: !!question.isCollected,
        loading: false 
      })
      // 详情接口匿名公开、不含收藏态，登录后用收藏集合反查真实状态
      this.syncCollectState()
    }).catch(err => {
      console.error('加载题目失败', err)
      this.setData({ 
        loading: false,
        error: '加载失败，请重试'
      })
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
    })
  },

  getTypeName(type) {
    const map = {
      'technical': '技术问题',
      'behavioral': '行为问题',
      'hr': 'HR问题',
      'project': '项目问题',
      'basic': '基础问题',
      'advanced': '高级问题'
    }
    return map[type] || '通用问题'
  },

  getDifficultyName(difficulty) {
    const map = {
      'easy': '简单',
      'medium': '中等',
      'hard': '困难'
    }
    return map[difficulty] || '中等'
  },

  // 同步当前用户对本题的收藏状态（详情接口不返回该字段）
  syncCollectState() {
    if (!wx.getStorageSync('token')) return
    api.profile.getCollections().then(res => {
      const data = res.data || {}
      const qList = Array.isArray(data.questions) ? data.questions : []
      const collected = qList.some(q => q.questionId === this.data.questionId)
      this.setData({ isCollected: collected })
    }).catch(() => {})
  },

  // 收藏 / 取消收藏
  toggleCollect() {
    if (!wx.getStorageSync('token')) {
      wx.navigateTo({ url: '/pages/login/login' })
      return
    }
    const next = !this.data.isCollected
    this.setData({ isCollected: next })
    // 收藏走 POST、取消走 DELETE（后端非切换式接口）
    const req = next
      ? api.question.collect(this.data.questionId)
      : api.question.uncollect(this.data.questionId)
    req.then(() => {
      wx.showToast({
        title: next ? '已收藏' : '已取消收藏',
        icon: 'none'
      })
    }).catch(err => {
      // 失败回滚
      this.setData({ isCollected: !next })
      wx.showToast({ title: '操作失败', icon: 'none' })
    })
  },

  // 开始练习
  startPractice() {
    if (!wx.getStorageSync('token')) {
      wx.navigateTo({ url: '/pages/login/login' })
      return
    }
    const q = this.data.question
    const jobTitle = q.jobTitle || ''
    wx.navigateTo({
      url: `/pages/interview/create?jobTitle=${encodeURIComponent(jobTitle)}`
    })
  }
})
