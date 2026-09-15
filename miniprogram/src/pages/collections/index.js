const api = require('../../api/index')

const DIFF_MAP = {
  easy: { text: '简单', cls: 'success' },
  medium: { text: '中等', cls: 'warning' },
  hard: { text: '困难', cls: 'danger' }
}

// 岗位方向枚举 → 中文名（与社区页 categories 保持一致）
const JOB_CATEGORY_MAP = {
  backend: '后端开发',
  frontend: '前端开发',
  bigdata: '大数据',
  ai: 'AI/算法',
  accounting: '会计/财务',
  general: '通用'
}

Page({
  data: {
    loading: true,
    tab: 'question',
    questions: [],
    articles: []
  },

  onShow() {
    this.loadCollections()
  },

  onPullDownRefresh() {
    this.loadCollections(() => wx.stopPullDownRefresh())
  },

  switchTab(e) {
    this.setData({ tab: e.currentTarget.dataset.tab })
  },

  // 加载收藏列表（后端返回 { questions, articles } 结构，需分别取数组）
  loadCollections(done) {
    this.setData({ loading: true })
    api.profile.getCollections().then(res => {
      const data = res.data || {}
      const qList = Array.isArray(data.questions) ? data.questions : []
      const aList = Array.isArray(data.articles) ? data.articles : []
      const questions = qList.map(q => {
        const diff = DIFF_MAP[q.difficulty] || { text: q.difficulty || '—', cls: 'info' }
        // 题目主键为 questionId，统一映射为 id 供模板绑定与跳转使用
        return { ...q, id: q.questionId, diffText: diff.text, diffCls: diff.cls }
      })
      // 文章主键为 articleId，统一映射为 id；jobCategory 枚举转中文
      const articles = aList.map(a => ({
        ...a,
        id: a.articleId,
        jobCategoryLabel: JOB_CATEGORY_MAP[a.jobCategory] || a.jobCategory || ''
      }))
      this.setData({ loading: false, questions, articles })
      if (done) done()
    }).catch(() => {
      this.setData({ loading: false })
      if (done) done()
    })
  },

  // 跳转题目详情
  viewQuestion(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: '/pages/questions/detail/index?id=' + id })
  },

  // 跳转文章详情
  viewArticle(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: '/pages/community/detail?id=' + id })
  },

  // 取消收藏题目
  uncollect(e) {
    const id = e.currentTarget.dataset.id
    const question = this.data.questions.find(q => q.id === id)
    if (!question) return
    wx.showModal({
      title: '取消收藏',
      content: '确定取消收藏该题目吗？',
      confirmColor: '#ff4d4f',
      success: res => {
        if (!res.confirm) return
        api.question.uncollect(id).then(() => {
          const questions = this.data.questions.filter(q => q.id !== id)
          this.setData({ questions })
          wx.showToast({ title: '已取消收藏', icon: 'success' })
        }).catch(() => {})
      }
    })
  },

  // 取消收藏文章
  uncollectArticle(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '取消收藏',
      content: '确定取消收藏该文章吗？',
      confirmColor: '#ff4d4f',
      success: res => {
        if (!res.confirm) return
        api.community.uncollect(id).then(() => {
          const articles = this.data.articles.filter(a => a.id !== id)
          this.setData({ articles })
          wx.showToast({ title: '已取消收藏', icon: 'success' })
        }).catch(() => {})
      }
    })
  },

  // 去题库逛逛
  goQuestions() {
    wx.navigateTo({ url: '/pages/questions/index' })
  },

  // 去社区逛逛
  goCommunity() {
    wx.navigateTo({ url: '/pages/community/index' })
  }
})
