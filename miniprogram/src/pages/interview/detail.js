const api = require('../../api/index')

Page({
  data: {
    loading: true,
    interviewId: null,
    interview: null,
    report: null,
    isNew: false,
    // wxml 不支持函数调用，展示文本/评分等级类必须入 data
    statusText: '',
    typeText: '',
    scoreClassTotal: '',
    scoreClassAccuracy: '',
    scoreClassFluency: '',
    scoreClassLogic: ''
  },

  onLoad(options) {
    if (options.id) {
      this.setData({ 
        interviewId: options.id,
        isNew: options.new === '1'
      })
      this.loadData()
    }

    // 新完成的面试显示提示
    if (options.new === '1') {
      wx.showToast({
        title: '面试完成！',
        icon: 'success'
      })
    }
  },

  loadData() {
    this.setData({ loading: true })
    
    Promise.all([
      api.interview.get(this.data.interviewId),
      api.report.get(this.data.interviewId).catch(() => null)
    ]).then(([interviewRes, reportRes]) => {
      this.setData({
        loading: false,
        interview: interviewRes.data,
        report: reportRes ? reportRes.data : null
      })
      this.syncDetailDisplay()
    }).catch(err => {
      this.setData({ loading: false })
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
    })
  },

  continueInterview() {
    // 继续未完成的面试
    wx.redirectTo({
      url: `/pages/interview/session?id=${this.data.interviewId}`
    })
  },

  restartInterview() {
    // 重新开始面试
    wx.showModal({
      title: '重新面试',
      content: '确定要重新开始面试吗？这将清除之前的记录。',
      success: (res) => {
        if (res.confirm) {
          wx.redirectTo({
            url: `/pages/interview/session?id=${this.data.interviewId}&restart=1`
          })
        }
      }
    })
  },

  shareReport() {
    // 分享报告
    wx.showShareMenu({
      withShareTicket: true
    })
  },

  onShareAppMessage() {
    const { interview, report } = this.data
    return {
      title: `我的面试报告 - ${interview?.company || '面试'} ${report?.totalScore || 0}分`,
      path: `/pages/interview/detail?id=${this.data.interviewId}&share=1`
    }
  },

  copyAnswer() {
    // 复制回答内容
    if (this.data.interview?.answers) {
      const text = this.data.interview.answers
        .map((item, index) => `Q${index + 1}: ${item.question}\nA: ${item.answer}`)
        .join('\n\n')
      
      wx.setClipboardData({
        data: text,
        success: () => {
          wx.showToast({ title: '已复制', icon: 'success' })
        }
      })
    }
  },

  // 同步展示文本与评分等级类到 data（供 wxml 渲染）
  syncDetailDisplay() {
    const { interview, report } = this.data
    this.setData({
      statusText: interview ? this.getStatusText(interview.status) : '',
      typeText: interview ? this.getTypeText(interview.interviewType) : '',
      scoreClassTotal: report ? this.getScoreClass(report.totalScore || 0) : '',
      scoreClassAccuracy: report ? this.getScoreClass(report.accuracy || 0) : '',
      scoreClassFluency: report ? this.getScoreClass(report.fluency || 0) : '',
      scoreClassLogic: report ? this.getScoreClass(report.logic || 0) : ''
    })
  },

  getStatusText(status) {
    const map = {
      'pending': '待开始',
      'in_progress': '进行中',
      'completed': '已完成',
      'paused': '已暂停'
    }
    return map[status] || status
  },

  getTypeText(type) {
    const map = {
      'technical': '技术面试',
      'hr': 'HR面试',
      'behavioral': '行为面试',
      'mock': '模拟面试'
    }
    return map[type] || type
  },

  getScoreClass(score) {
    if (score >= 80) return 'excellent'
    if (score >= 60) return 'good'
    return 'warning'
  }
})
