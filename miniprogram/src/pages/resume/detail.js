const api = require('../../api/index')

Page({
  data: {
    loading: true,
    resumeId: null,
    resume: null,
    retrying: false,
    startingInterview: false,
    pollTimer: null
  },

  onLoad(options) {
    if (options.id) {
      this.setData({ resumeId: options.id })
      this.loadData()
    }
  },

  onUnload() {
    this.clearPoll()
  },

  /** 加载简历详情 */
  loadData() {
    this.setData({ loading: true })
    api.resume.get(this.data.resumeId).then(res => {
      var resume = res.data || res
      this.setData({
        loading: false,
        resume: this.formatResume(resume)
      })
      // 分析中时轮询状态
      if (resume.analysisStatus === 'pending' || resume.analysisStatus === 'analyzing') {
        this.startPoll()
      }
    }).catch(err => {
      console.log('加载简历详情失败', err)
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    })
  },

  /** 轮询分析状态 */
  startPoll() {
    this.clearPoll()
    var that = this
    var timer = setInterval(function () {
      api.resume.get(that.data.resumeId).then(function (res) {
        var resume = res.data || res
        that.setData({ resume: that.formatResume(resume) })
        if (resume.analysisStatus === 'done' || resume.analysisStatus === 'failed') {
          that.clearPoll()
        }
      }).catch(function () {})
    }, 5000)
    this.setData({ pollTimer: timer })
  },

  clearPoll() {
    if (this.data.pollTimer) {
      clearInterval(this.data.pollTimer)
      this.setData({ pollTimer: null })
    }
  },

  /** 格式化简历数据 */
  formatResume(resume) {
    // 总分存在 analysis.totalScore（顶层 totalScore 不存在，读错会导致评分环永远灰色）
    var analysisScore = null
    if (resume.analysis && resume.analysis.totalScore != null) {
      analysisScore = resume.analysis.totalScore
    } else if (resume.totalScore != null) {
      analysisScore = resume.totalScore
    }
    // 维度最强/最弱一句话总结
    var dims = (resume.analysis && resume.analysis.dimensions) || []
    var dimSummary = ''
    if (dims.length > 0) {
      var sorted = dims.slice().sort(function (a, b) { return (b.score || 0) - (a.score || 0) })
      var best = sorted[0]
      var worst = sorted[sorted.length - 1]
      dimSummary = sorted.length > 1
        ? '最强 ' + best.name + ' · 最弱 ' + worst.name
        : best.name + ' ' + (best.score || 0) + ' 分'
    }
    return Object.assign({}, resume, {
      uploadedAtText: this.formatDate(resume.uploadedAt),
      fileSizeText: this.formatFileSize(resume.fileSize),
      statusText: this.statusText(resume.analysisStatus),
      statusClass: this.statusClass(resume.analysisStatus),
      scoreClass: this.scoreClass(analysisScore),
      gradeText: this.gradeText(analysisScore),
      dimSummary: dimSummary,
      isAnalyzing: resume.analysisStatus === 'pending' || resume.analysisStatus === 'analyzing',
      isFailed: resume.analysisStatus === 'failed',
      isDone: resume.analysisStatus === 'done',
      hasSkills: resume.parsedContent && resume.parsedContent.skills && resume.parsedContent.skills.length > 0,
      hasWork: resume.parsedContent && resume.parsedContent.workExperience && resume.parsedContent.workExperience.length > 0,
      hasProjects: resume.parsedContent && resume.parsedContent.projects && resume.parsedContent.projects.length > 0,
      hasJobTitle: resume.parsedContent && resume.parsedContent.jobTitle,
      hasSuggestions: resume.analysis && resume.analysis.suggestions && resume.analysis.suggestions.length > 0,
      hasDimensions: resume.analysis && resume.analysis.dimensions && resume.analysis.dimensions.length > 0
    })
  },

  /** 重新分析 */
  handleRetry() {
    if (this.data.retrying) return
    this.setData({ retrying: true })
    api.resume.retryAnalyze(this.data.resumeId).then(() => {
      this.setData({ retrying: false })
      wx.showToast({ title: '重新分析已启动', icon: 'success' })
      // 更新状态为分析中
      var resume = this.data.resume
      resume.analysisStatus = 'pending'
      resume.statusText = '待分析'
      resume.statusClass = 'pending'
      resume.isAnalyzing = true
      resume.isFailed = false
      resume.isDone = false
      this.setData({ resume: resume })
      this.startPoll()
    }).catch(err => {
      this.setData({ retrying: false })
      wx.showToast({ title: '重试失败', icon: 'none' })
    })
  },

  /** 删除简历 */
  handleDelete() {
    var that = this
    wx.showModal({
      title: '确认删除',
      content: '确定要删除简历「' + (this.data.resume.filename || '') + '」吗？此操作不可恢复。',
      confirmColor: '#ff4d4f',
      confirmText: '删除',
      success: function (res) {
        if (res.confirm) {
          wx.showLoading({ title: '删除中…' })
          api.resume.delete(that.data.resumeId).then(function () {
            wx.hideLoading()
            wx.showToast({ title: '删除成功', icon: 'success' })
            setTimeout(function () {
              wx.navigateBack()
            }, 1000)
          }).catch(function () {
            wx.hideLoading()
            wx.showToast({ title: '删除失败', icon: 'none' })
          })
        }
      }
    })
  },

  /** 基于简历创建面试：先选答题模式（文字/语音） */
  handleCreateInterview() {
    if (this.data.startingInterview) return
    if (!this.data.resume || this.data.resume.analysisStatus !== 'done') {
      wx.showToast({ title: '请等待分析完成', icon: 'none' })
      return
    }
    var that = this
    wx.showActionSheet({
      alertText: '选择答题方式',
      itemList: [
        '✍️ 文字答题 · 打字输入',
        '🎙️ 语音答题 · 说话自动转文字，AI 分析语速/口头禅/表达'
      ],
      success: function (res) {
        // tapIndex 1 = 语音模式（后端值 video）；其它为文字
        that.doCreateInterview(res.tapIndex === 1 ? 'video' : 'text')
      }
    })
  },

  /** 按选中模式创建面试 */
  doCreateInterview(mode) {
    if (this.data.startingInterview) return
    this.setData({ startingInterview: true })
    wx.showLoading({ title: '创建面试…', mask: true })
    api.resume.createInterview(this.data.resumeId, mode).then(res => {
      wx.hideLoading()
      this.setData({ startingInterview: false })
      var interviewId = (res.data && res.data.interviewId) || res.interviewId
      wx.showToast({ title: '面试已创建', icon: 'success' })
      if (interviewId) {
        setTimeout(function () {
          wx.navigateTo({ url: '/pages/interview/session?id=' + interviewId })
        }, 1000)
      }
    }).catch(err => {
      wx.hideLoading()
      this.setData({ startingInterview: false })
      wx.showToast({ title: err.message || '发起面试失败', icon: 'none' })
    })
  },

  /** 返回列表 */
  goBack() {
    wx.navigateBack()
  },

  /** 日期格式化 */
  formatDate(val) {
    if (!val) return '—'
    var d = new Date(val)
    var p = function (n) { return String(n).padStart(2, '0') }
    return d.getFullYear() + '-' + p(d.getMonth() + 1) + '-' + p(d.getDate()) + ' ' + p(d.getHours()) + ':' + p(d.getMinutes())
  },

  /** 文件大小格式化 */
  formatFileSize(bytes) {
    if (!bytes) return '未知大小'
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    return (bytes / 1024 / 1024).toFixed(1) + ' MB'
  },

  statusText(s) {
    var map = { pending: '待分析', analyzing: '分析中', done: '已分析', failed: '分析失败' }
    return map[s] || s || '未知'
  },

  statusClass(s) {
    var map = { pending: 'pending', analyzing: 'analyzing', done: 'done', failed: 'failed' }
    return map[s] || 'pending'
  },

  scoreClass(score) {
    if (score == null) return 'score-none'
    if (score >= 80) return 'score-high'
    if (score >= 60) return 'score-mid'
    return 'score-low'
  },

  gradeText(score) {
    if (score == null) return '暂无评分'
    if (score >= 90) return '优秀'
    if (score >= 80) return '良好'
    if (score >= 60) return '及格'
    return '待提升'
  }
})
