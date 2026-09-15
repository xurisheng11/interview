const api = require('../../api/index')

const DIFF_MAP = {
  easy: { text: '简单', cls: 'success' },
  medium: { text: '中等', cls: 'warning' },
  hard: { text: '困难', cls: 'danger' }
}

Page({
  data: {
    loading: true,
    collections: []
  },

  onShow() {
    this.loadCollections()
  },

  onPullDownRefresh() {
    this.loadCollections(() => wx.stopPullDownRefresh())
  },

  // 加载收藏列表
  loadCollections(done) {
    this.setData({ loading: true })
    api.profile.getCollections().then(res => {
      const list = res.data || []
      const collections = (Array.isArray(list) ? list : []).map(q => {
        const diff = DIFF_MAP[q.difficulty] || { text: q.difficulty || '—', cls: 'info' }
        return {
          ...q,
          diffText: diff.text,
          diffCls: diff.cls
        }
      })
      this.setData({ loading: false, collections })
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

  // 取消收藏
  uncollect(e) {
    const id = e.currentTarget.dataset.id
    const question = this.data.collections.find(q => q.id === id)
    if (!question) return
    wx.showModal({
      title: '取消收藏',
      content: '确定取消收藏该题目吗？',
      confirmColor: '#ff4d4f',
      success: res => {
        if (!res.confirm) return
        api.question.collect(id).then(() => {
          const collections = this.data.collections.filter(q => q.id !== id)
          this.setData({ collections })
          wx.showToast({ title: '已取消收藏', icon: 'success' })
        }).catch(() => {})
      }
    })
  },

  // 去题库逛逛
  goQuestions() {
    wx.navigateTo({ url: '/pages/questions/index' })
  }
})
