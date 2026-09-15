const api = require('../../api/index')

Page({
  data: {
    list: []
  },

  onLoad() {
    this.loadList()
  },

  onShow() {
    this.loadList()
  },

  onPullDownRefresh() {
    this.loadList().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  loadList() {
    return api.interview.list().then(res => {
      // 后端返回数组直接是列表
      const list = (res.data || []).map(item => ({
        ...item,
        statusText: this.getStatusText(item.status),
        difficultyText: this.getDifficultyText(item.difficulty),
        roundText: this.getRoundText(item.round),
        scoreClass: this.getScoreClass(item.totalScore)
      }))
      this.setData({ list })
    }).catch(err => {
      console.log('加载面试列表失败', err)
      this.setData({ list: [] })
    })
  },

  startNewInterview() {
    // 检查登录状态
    const token = wx.getStorageSync('token')
    if (!token) {
      wx.navigateTo({ url: '/pages/login/login' })
      return
    }

    // 跳转到面试配置页
    wx.navigateTo({
      url: '/pages/interview/create'
    })
  },

  viewDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/interview/detail?id=${id}`
    })
  },

  getStatusText(status) {
    const map = {
      'pending': '待开始',
      'ongoing': '进行中',
      'in_progress': '进行中',
      'completed': '已完成',
      'paused': '已暂停'
    }
    return map[status] || status
  },

  getDifficultyText(difficulty) {
    const map = {
      'easy': '简单',
      'medium': '中等',
      'hard': '困难'
    }
    return map[difficulty] || '中等'
  },

  getRoundText(round) {
    const map = {
      'round1': '初试',
      'round2': '复试',
      'round3': '终面',
      'comprehensive': '综合面试'
    }
    return map[round] || round
  },

  getScoreClass(score) {
    if (score >= 80) return 'excellent'
    if (score >= 60) return 'good'
    return 'warning'
  }
})
