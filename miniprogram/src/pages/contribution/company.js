const api = require('../../api/index')

const TYPE_MAP = { technical: '技术', behavioral: '行为', algorithm: '算法', 'system-design': '系统设计' }
const DIFF_MAP = {
  easy: { text: '简单', cls: 'success' },
  medium: { text: '中等', cls: 'warning' },
  hard: { text: '困难', cls: 'danger' }
}

Page({
  data: {
    searchCompany: '',
    searching: false,
    credits: 0,
    activeTab: 'verified',
    verifiedList: [],
    pendingList: [],
    // 投票状态：questionId -> helpful/useless
    votedMap: {},
    // 降级提示
    fallback: null,
    // 举报弹窗
    showReport: false,
    reportQuestion: null,
    reportReason: ''
  },

  onLoad(options) {
    this.loadCredits()
    // 支持从其他页面带公司参数进入
    if (options.company) {
      this.setData({ searchCompany: options.company })
      this.loadQuestions(options.company)
    }
  },

  onShow() {
    this.loadCredits()
  },

  // 输入公司名称
  onCompanyInput(e) {
    this.setData({ searchCompany: e.detail.value })
  },

  // 加载积分
  loadCredits() {
    api.contribution.getCredits().then(res => {
      const credits = res.data.credits != null ? res.data.credits : res.data
      this.setData({ credits: typeof credits === 'number' ? credits : 0 })
    }).catch(() => {})
  },

  // 搜索
  search() {
    const company = this.data.searchCompany.trim()
    if (!company) {
      wx.showToast({ title: '请输入公司名称', icon: 'none' })
      return
    }
    this.loadQuestions(company)
  },

  // 加载公司题目（搜索已验证题库扣 1 积分）
  loadQuestions(company) {
    this.setData({ searching: true, fallback: null })
    wx.showLoading({ title: '加载中…', mask: true })

    // 先扣积分（失败不阻断）
    api.contribution.deductCredits(1).then(res => {
      const credits = res.data.credits != null ? res.data.credits : this.data.credits - 1
      this.setData({ credits: typeof credits === 'number' ? credits : this.data.credits - 1 })
    }).catch(() => {})

    api.contribution.getCompanyQuestions(company, '').then(res => {
      wx.hideLoading()
      const data = res.data || {}
      const decorate = q => {
        const diff = DIFF_MAP[q.difficulty] || { text: q.difficulty, cls: 'info' }
        return {
          ...q,
          typeText: TYPE_MAP[q.questionType] || q.questionType || '—',
          diffText: diff.text,
          diffCls: diff.cls
        }
      }
      const verifiedList = (data.verified || []).map(decorate)
      const pendingList = (data.approved || []).map(decorate)

      this.setData({
        searching: false,
        verifiedList,
        pendingList,
        credits: data.credits != null ? data.credits : this.data.credits
      })

      // 搜不到，调用降级接口
      if (verifiedList.length + pendingList.length === 0) {
        this.loadFallback(company)
      }
    }).catch(() => {
      wx.hideLoading()
      this.setData({ searching: false })
    })
  },

  // 降级方案：提示相似公司
  loadFallback(company) {
    api.contribution.search(company, '').then(res => {
      this.setData({ fallback: res.data || { message: '未找到【' + company + '】的专属题库', similar: [] } })
    }).catch(() => {
      this.setData({ fallback: { message: '未找到【' + company + '】的专属题库', similar: [] } })
    })
  },

  // 切换 tab
  switchTab(e) {
    this.setData({ activeTab: e.currentTarget.dataset.tab })
  },

  // 跳转相似公司
  goToCompany(e) {
    const company = e.currentTarget.dataset.company
    this.setData({ searchCompany: company })
    this.loadQuestions(company)
  },

  // 投票（有用/没用）
  vote(e) {
    const { id, vote } = e.currentTarget.dataset
    if (this.data.votedMap[id]) return

    api.contribution.vote(id, vote).then(() => {
      const votedMap = { ...this.data.votedMap, [id]: vote }
      // 更新本地计数
      const update = list => list.map(q => {
        if (q.id !== id) return q
        if (vote === 'helpful') q.helpfulCount = (q.helpfulCount || 0) + 1
        else q.uselessCount = (q.uselessCount || 0) + 1
        return { ...q }
      })
      this.setData({
        votedMap,
        verifiedList: update(this.data.verifiedList),
        pendingList: update(this.data.pendingList)
      })
      wx.showToast({ title: '投票成功 +1 积分', icon: 'success' })
      this.loadCredits()
    }).catch(() => {})
  },

  // 打开举报弹窗
  openReport(e) {
    const id = e.currentTarget.dataset.id
    const question = this.data.pendingList.find(q => q.id === id) ||
      this.data.verifiedList.find(q => q.id === id)
    if (question) {
      this.setData({ showReport: true, reportQuestion: question, reportReason: '' })
    }
  },

  // 举报原因输入
  onReportInput(e) {
    this.setData({ reportReason: e.detail.value })
  },

  // 关闭举报弹窗
  closeReport() {
    this.setData({ showReport: false })
  },

  // 提交举报
  submitReport() {
    const q = this.data.reportQuestion
    if (!q) return
    api.contribution.report(q.id, this.data.reportReason).then(() => {
      this.setData({ showReport: false })
      wx.showToast({ title: '举报已收到，感谢反馈', icon: 'success' })
    }).catch(() => {})
  },

  // 去贡献题目
  goContribute() {
    wx.navigateTo({ url: '/pages/contribution/contribute' })
  },

  // 阻止冒泡
  noop() {}
})
