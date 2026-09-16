// 分享报告页（好友通过分享卡片打开，只读、免登录、脱敏汇总数据）
const ROUND_NAMES = { round1: '一面', round2: '二面', round3: '三面', comprehensive: '综合面试' }
const DIFFICULTY_NAMES = { easy: '简单', medium: '中等', hard: '困难' }
const PASS_NAMES = { pass: '建议通过', pending: '待定', fail: '未通过' }

function scoreClass(score) {
  if (score >= 90) return 'excellent'
  if (score >= 75) return 'good'
  if (score >= 60) return 'pass'
  return 'warning'
}

Page({
  data: {
    loading: true,
    error: '',
    token: '',
    view: {}
  },

  onLoad(options) {
    const token = options.token || ''
    if (!token) {
      this.setData({ loading: false, error: '分享链接不完整' })
      return
    }
    this.setData({ token })
    this.loadShare(token)
  },

  loadShare(token) {
    const app = getApp()
    wx.request({
      // 直接用 wx.request：公开接口免登录，避免通用封装在 token 过期时误跳登录页
      url: app.globalData.apiBaseUrl + '/reports/share/' + token,
      method: 'GET',
      success: (res) => {
        if (res.statusCode === 200 && res.data && res.data.code === 200 && res.data.data) {
          this.buildView(res.data.data)
          this.setData({ loading: false, error: '' })
        } else {
          this.setData({
            loading: false,
            error: (res.data && res.data.message) || '分享链接无效或已过期'
          })
        }
      },
      fail: () => {
        this.setData({ loading: false, error: '网络异常，请稍后重试' })
      }
    })
  },

  // 后端返回脱敏汇总：totalScore/grade/passStatus/moduleScores/aiSummary/表达指标等
  buildView(r) {
    const v = {
      company: r.companyName || 'AI 模拟面试',
      jobTitle: r.jobTitle || '通用面试',
      roundText: ROUND_NAMES[r.round] || r.round || '通用',
      difficultyText: DIFFICULTY_NAMES[r.difficulty] || r.difficulty || '综合',
      modeText: r.mode === 'video_call' ? '视频面试' : (r.mode === 'video' ? '语音面试' : '文字面试'),
      totalScore: r.totalScore || 0,
      grade: r.grade || '',
      scoreClass: scoreClass(r.totalScore || 0),
      passText: PASS_NAMES[r.passStatus] || r.passStatus || '',
      passClass: r.passStatus || '',
      passReason: r.passReason || '',
      answeredCount: r.answeredCount || 0,
      totalCount: r.totalCount || 0
    }

    // 指标条（语音/视频面试附加表达数据）
    const metrics = []
    if (r.avgExpressionScore > 0) metrics.push({ label: '平均表达分', value: Math.round(r.avgExpressionScore) + ' 分' })
    if (r.avgSpeechRate > 0) metrics.push({ label: '平均语速', value: Math.round(r.avgSpeechRate) + ' 字/分' })
    if (r.faceSummary && r.faceSummary.samples > 0) {
      metrics.push({ label: '紧张程度', value: r.faceSummary.nervousLevel || '放松' })
      v.faceNote = 'AI 通过摄像头抓帧分析了面试者的情绪状态'
    }
    v.metrics = metrics

    v.moduleScores = (r.moduleScores || []).map(m => ({
      module: m.module,
      count: m.count,
      level: m.level,
      avgText: Math.round(m.avgScore),
      width: Math.max(0, Math.min(100, Math.round(m.avgScore))),
      cls: scoreClass(m.avgScore)
    }))

    if (r.aiSummary) {
      v.strengths = (r.aiSummary.strengths || []).slice(0, 3)
      v.weaknesses = (r.aiSummary.weaknesses || []).slice(0, 3).map(w => ({
        point: w.point, suggestion: w.suggestion
      }))
    }
    v.expressionSummary = r.expressionSummary || ''

    this.setData({ view: v })
  },

  goInterview() {
    wx.switchTab({
      url: '/pages/interview/list',
      fail: () => wx.reLaunch({ url: '/pages/index/index' })
    })
  },

  // 允许收到报告的好友二次转发
  onShareAppMessage() {
    const { view, token } = this.data
    return {
      title: view.totalScore
        ? `TA 的面试评测报告：${view.totalScore} 分 · ${view.grade}｜你也来模拟一场？`
        : 'AI 模拟面试 · 面试评测报告',
      path: '/pages/interview/share?token=' + token
    }
  }
})
