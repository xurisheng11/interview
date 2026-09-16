const api = require('../../api/index')

// 枚举中文名（与后端 model 对齐）
const STATUS_NAMES = { ongoing: '进行中', paused: '已暂停', completed: '已完成' }
const ROUND_NAMES = { round1: '一面', round2: '二面', round3: '三面', comprehensive: '综合面试' }
const DIFFICULTY_NAMES = { easy: '简单', medium: '中等', hard: '困难' }
const PASS_NAMES = { pass: '建议通过', pending: '待定', fail: '未通过' }

// ISO 时间 → 本地可读格式（修复带 T/Z 的原始时间串）
function fmtTime(val) {
  if (!val) return '--'
  const d = new Date(val)
  if (!isNaN(d.getTime())) {
    const pad = n => String(n).padStart(2, '0')
    return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) +
      ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes())
  }
  const m = String(val).match(/^(\d{4}-\d{2}-\d{2})[T ](\d{2}:\d{2})/)
  return m ? m[1] + ' ' + m[2] : String(val)
}

function fmtDuration(seconds) {
  if (!seconds && seconds !== 0) return '--'
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return m > 0 ? m + '分' + s + '秒' : s + '秒'
}

function scoreClass(score) {
  if (score >= 90) return 'excellent'
  if (score >= 75) return 'good'
  if (score >= 60) return 'pass'
  return 'warning'
}

Page({
  data: {
    loading: true,
    interviewId: null,
    interview: null,
    report: null,
    // wxml 不支持函数调用，所有展示字段在 buildView 里预计算
    view: {}
  },

  onLoad(options) {
    if (options.id) {
      this.setData({ interviewId: options.id })
      this.loadData()
    }
    if (options.new === '1') {
      wx.showToast({ title: '面试完成！', icon: 'success' })
    }
  },

  onShow() {
    if (this.data.interviewId && !this.data.loading) this.loadData()
  },

  loadData() {
    this.setData({ loading: true })
    Promise.all([
      api.interview.get(this.data.interviewId),
      api.report.get(this.data.interviewId).catch(() => null)
    ]).then(([interviewRes, reportRes]) => {
      const interview = interviewRes.data
      const report = reportRes ? reportRes.data : null
      this.setData({ loading: false, interview, report })
      this.buildView(interview, report)
    }).catch(() => {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    })
  },

  // 把后端真实结构映射成页面展示字段
  buildView(interview, report) {
    const config = (interview && interview.config) || {}
    const v = {
      title: (interview.companyName || '未指定公司'),
      subtitle: (config.jobTitle || '通用面试'),
      statusText: STATUS_NAMES[interview.status] || interview.status || '',
      roundText: ROUND_NAMES[config.round] || config.round || '通用',
      difficultyText: DIFFICULTY_NAMES[config.difficulty] || config.difficulty || '综合',
      modeText: interview.mode === 'video_call' ? '视频面试' : (interview.mode === 'video' ? '语音面试' : '文字面试'),
      startTimeText: fmtTime(interview.startTime),
      questionCount: (interview.questions || []).length,
      isCompleted: interview.status === 'completed'
    }

    if (report) {
      v.totalScore = report.totalScore || 0
      v.grade = report.grade || ''
      v.scoreClass = scoreClass(report.totalScore || 0)
      v.passText = PASS_NAMES[report.passStatus] || report.passStatus || ''
      v.passClass = report.passStatus || ''
      v.passReason = report.passReason || ''
      v.durationText = fmtDuration(report.totalSeconds)
      v.answeredCount = report.answeredCount || 0
      v.skippedCount = report.skippedCount || 0
      v.totalCount = report.totalCount || 0
      v.startTimeText = fmtTime(report.startTime || interview.startTime)

      // 知识点模块得分
      v.moduleScores = (report.moduleScores || []).map(m => ({
        module: m.module,
        count: m.count,
        level: m.level,
        avgText: Math.round(m.avgScore),
        width: Math.max(0, Math.min(100, Math.round(m.avgScore))),
        cls: scoreClass(m.avgScore)
      }))

      // 逐题点评
      v.questions = (report.questions || []).map((q, i) => {
        const answered = !q.skipped && (q.userAnswer || '').trim() !== ''
        return {
          idx: i + 1,
          content: q.content,
          userAnswer: q.userAnswer || '',
          answered: answered,
          skipped: !answered,
          scoreText: answered ? (q.score || 0) : '跳过',
          scoreCls: answered ? scoreClass(q.score || 0) : 'none',
          difficulty: DIFFICULTY_NAMES[q.difficulty] || q.difficulty || '',
          tagsText: (q.tags || []).join(' · '),
          pros: q.pros || [],
          cons: q.cons || [],
          referenceAnswer: q.referenceAnswer || '',
          expressionFeedback: q.expressionFeedback || '',
          _refOpen: false
        }
      })

      // AI 综合评价
      if (report.aiSummary) {
        v.strengths = report.aiSummary.strengths || []
        v.weaknesses = (report.aiSummary.weaknesses || []).map(w => ({
          point: w.point, suggestion: w.suggestion, resource: w.resource
        }))
        v.roadmap = report.aiSummary.roadmap || ''
      }

      // 语音/视频面试附加表达指标
      if (report.mode === 'video' || report.mode === 'video_call') {
        const metrics = []
        if (report.avgExpressionScore > 0) metrics.push({ label: '平均表达分', value: report.avgExpressionScore + ' 分' })
        if (report.avgSpeechRate > 0) metrics.push({ label: '平均语速', value: Math.round(report.avgSpeechRate) + ' 字/分' })
        if (report.avgThinkDuration > 0) metrics.push({ label: '平均思考', value: fmtDuration(report.avgThinkDuration) })
        v.videoMetrics = metrics
        v.expressionSummary = report.expressionSummary || ''
        // 视频面试：面部情绪抓帧分析（wxml 不支持函数调用，占比文本预计算）
        if (report.faceSummary && report.faceSummary.samples > 0) {
          const fs = report.faceSummary
          v.faceSummary = {
            dominant: fs.dominantName || '未知',
            nervousLevel: fs.nervousLevel || '',
            tenseText: Math.round((fs.tenseRatio || 0) * 100) + '%',
            smileText: Math.round((fs.smileRatio || 0) * 100) + '%',
            samples: fs.samples
          }
        }
      }
    }

    this.setData({ view: v })
  },

  // 展开/收起参考答案
  toggleRef(e) {
    const i = e.currentTarget.dataset.index
    const key = 'view.questions[' + i + ']._refOpen'
    const data = {}
    data[key] = !this.data.view.questions[i]._refOpen
    this.setData(data)
  },

  continueInterview() {
    wx.redirectTo({ url: `/pages/interview/session?id=${this.data.interviewId}` })
  },

  restartInterview() {
    wx.showModal({
      title: '重新面试',
      content: '确定要重新开始面试吗？这将清除之前的记录。',
      success: (res) => {
        if (res.confirm) {
          wx.redirectTo({ url: `/pages/interview/session?id=${this.data.interviewId}&restart=1` })
        }
      }
    })
  },

  // 下载 Word 版报告（打开后可通过右上角菜单转发/保存）
  downloadWord() {
    const app = getApp()
    const token = wx.getStorageSync('token')
    wx.showLoading({ title: '生成Word报告…', mask: true })
    wx.downloadFile({
      url: app.globalData.apiBaseUrl + '/reports/' + this.data.interviewId + '/word',
      header: { Authorization: 'Bearer ' + token },
      success: (res) => {
        if (res.statusCode !== 200) {
          wx.showToast({ title: '报告生成失败', icon: 'none' })
          return
        }
        wx.openDocument({
          filePath: res.tempFilePath,
          fileType: 'docx',
          showMenu: true,
          fail: () => wx.showToast({ title: '打开失败，请重试', icon: 'none' })
        })
      },
      fail: () => wx.showToast({ title: '下载失败，请检查网络', icon: 'none' }),
      complete: () => wx.hideLoading()
    })
  },

  onShareAppMessage() {
    const { view } = this.data
    return {
      title: `我的面试报告 - ${view.title || '面试模拟'} ${view.totalScore || 0}分`,
      path: `/pages/interview/detail?id=${this.data.interviewId}&share=1`
    }
  }
})
