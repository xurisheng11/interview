const api = require('../../api/index')

// 状态/类型/难度显示映射
const STATUS_MAP = {
  pending: { text: '待审核', cls: 'warning' },
  approved: { text: '已通过', cls: 'success' },
  verified: { text: '已验证', cls: 'primary' },
  rejected: { text: '已驳回', cls: 'danger' }
}
const TYPE_MAP = { technical: '技术', behavioral: '行为', algorithm: '算法', 'system-design': '系统设计' }
const DIFF_MAP = { easy: { text: '简单', cls: 'success' }, medium: { text: '中等', cls: 'warning' }, hard: { text: '困难', cls: 'danger' } }

Page({
  data: {
    loading: true,
    contributions: [],
    credits: 0,
    stats: {
      total: 0,
      verified: 0,
      approved: 0,
      pending: 0
    }
  },

  onShow() {
    this.loadData()
  },

  onPullDownRefresh() {
    this.loadData(() => wx.stopPullDownRefresh())
  },

  // 加载贡献记录与积分
  loadData(done) {
    this.setData({ loading: true })
    Promise.all([
      api.contribution.getMine().catch(() => null),
      api.contribution.getCredits().catch(() => null)
    ]).then(([contribRes, creditRes]) => {
      const list = (contribRes && (contribRes.data.contributions || contribRes.data)) || []
      const credits = (creditRes && (creditRes.data.credits != null ? creditRes.data.credits : creditRes.data)) || 0

      const contributions = (list || []).map(c => {
        const status = STATUS_MAP[c.status] || { text: c.status, cls: 'info' }
        const diff = DIFF_MAP[c.difficulty] || { text: c.difficulty, cls: 'info' }
        return {
          ...c,
          statusText: status.text,
          statusCls: status.cls,
          typeText: TYPE_MAP[c.questionType] || c.questionType || '—',
          diffText: diff.text,
          diffCls: diff.cls,
          dateText: this.formatDate(c.createdAt)
        }
      })

      this.setData({
        loading: false,
        contributions,
        credits: typeof credits === 'number' ? credits : 0,
        stats: {
          total: contributions.length,
          verified: contributions.filter(q => q.status === 'verified').length,
          approved: contributions.filter(q => q.status === 'approved').length,
          pending: contributions.filter(q => q.status === 'pending').length
        }
      })
      if (done) done()
    })
  },

  // 时间戳 → 日期文本
  formatDate(ts) {
    if (!ts) return '—'
    const d = new Date(ts * 1000)
    const month = d.getMonth() + 1
    const day = d.getDate()
    return d.getFullYear() + '-' + (month < 10 ? '0' + month : month) + '-' + (day < 10 ? '0' + day : day)
  },

  // 去贡献题目
  goContribute() {
    wx.navigateTo({ url: '/pages/contribution/contribute' })
  },

  // 去公司题库
  goCompany() {
    wx.navigateTo({ url: '/pages/contribution/company' })
  }
})
