const api = require('../../api/index')

// 题目类型与难度选项
const QUESTION_TYPES = [
  { value: 'technical', label: '技术题' },
  { value: 'behavioral', label: '行为题' },
  { value: 'algorithm', label: '手撕算法' },
  { value: 'system-design', label: '系统设计' }
]
const DIFFICULTIES = [
  { value: 'easy', label: '简单' },
  { value: 'medium', label: '中等' },
  { value: 'hard', label: '困难' }
]
const ROUNDS = ['一面', '二面', '三面', 'HR面', '终面']

// 状态显示映射
const STATUS_MAP = {
  pending: { text: '待审核', cls: 'warning' },
  approved: { text: '已通过', cls: 'success' },
  verified: { text: '已验证', cls: 'primary' },
  rejected: { text: '已驳回', cls: 'danger' }
}

Page({
  data: {
    credits: 0,
    submitting: false,
    // 表单
    company: '',
    jobTitle: '',
    year: new Date().getFullYear(),
    roundIndex: -1,
    rounds: ROUNDS,
    questions: [],
    // 选项
    questionTypes: QUESTION_TYPES,
    difficulties: DIFFICULTIES,
    // 我的贡献（最近）
    myContributions: [],
    // 类型选择弹层状态
    editingQuestionIndex: -1,
    showTypePicker: false,
    typeColumns: [QUESTION_TYPES.map(t => t.label)],
    showDiffPicker: false,
    diffColumns: [DIFFICULTIES.map(d => d.label)]
  },

  onLoad() {
    this.setData({ questions: [this.newQuestion()] })
  },

  onShow() {
    this.loadData()
  },

  onPullDownRefresh() {
    this.loadData(() => wx.stopPullDownRefresh())
  },

  // 加载积分与我的贡献
  loadData(done) {
    Promise.all([
      api.contribution.getCredits().catch(() => null),
      api.contribution.getMine().catch(() => null)
    ]).then(([creditRes, contribRes]) => {
      const credits = (creditRes && (creditRes.data.credits != null ? creditRes.data.credits : creditRes.data)) || 0
      const list = (contribRes && (contribRes.data.contributions || contribRes.data)) || []
      this.setData({
        credits: typeof credits === 'number' ? credits : 0,
        myContributions: (list || []).slice(0, 5).map(c => ({
          ...c,
          statusText: (STATUS_MAP[c.status] || {}).text || c.status,
          statusCls: (STATUS_MAP[c.status] || {}).cls || 'info'
        }))
      })
      if (done) done()
    })
  },

  // 新建一道空白题目
  newQuestion() {
    return {
      content: '',
      questionType: 'technical',
      questionTypeLabel: '技术题',
      difficulty: 'medium',
      difficultyLabel: '中等',
      tagsInput: ''
    }
  },

  // ===== 表单输入 =====

  onCompanyInput(e) {
    this.setData({ company: e.detail.value })
  },

  onJobTitleInput(e) {
    this.setData({ jobTitle: e.detail.value })
  },

  onYearChange(e) {
    this.setData({ year: Number(e.detail.value) || new Date().getFullYear() })
  },

  onRoundChange(e) {
    const index = Number(e.detail.value)
    this.setData({
      roundIndex: index,
      round: ROUNDS[index] || ''
    })
  },

  // 题目内容输入
  onQuestionInput(e) {
    const index = e.currentTarget.dataset.index
    const key = `questions[${index}].content`
    this.setData({ [key]: e.detail.value })
  },

  // 标签输入
  onTagsInput(e) {
    const index = e.currentTarget.dataset.index
    const key = `questions[${index}].tagsInput`
    this.setData({ [key]: e.detail.value })
  },

  // 打开题目类型选择
  openTypePicker(e) {
    const index = e.currentTarget.dataset.index
    this.setData({ editingQuestionIndex: index, showTypePicker: true })
  },

  // 选择题目类型
  chooseType(e) {
    const { value, label } = e.currentTarget.dataset
    const index = this.data.editingQuestionIndex
    if (index > -1) {
      this.setData({
        [`questions[${index}].questionType`]: value,
        [`questions[${index}].questionTypeLabel`]: label
      })
    }
    this.setData({ showTypePicker: false })
  },

  closeTypePicker() {
    this.setData({ showTypePicker: false })
  },

  // 打开难度选择
  openDiffPicker(e) {
    const index = e.currentTarget.dataset.index
    this.setData({ editingQuestionIndex: index, showDiffPicker: true })
  },

  // 选择难度
  chooseDiff(e) {
    const { value, label } = e.currentTarget.dataset
    const index = this.data.editingQuestionIndex
    if (index > -1) {
      this.setData({
        [`questions[${index}].difficulty`]: value,
        [`questions[${index}].difficultyLabel`]: label
      })
    }
    this.setData({ showDiffPicker: false })
  },

  closeDiffPicker() {
    this.setData({ showDiffPicker: false })
  },

  // 阻止事件冒泡/滑动穿透
  noop() {},

  // 添加一道题目（每次最多 20 道）
  addQuestion() {
    if (this.data.questions.length >= 20) {
      wx.showToast({ title: '每次最多提交 20 道题目', icon: 'none' })
      return
    }
    this.setData({ questions: [...this.data.questions, this.newQuestion()] })
  },

  // 删除题目
  removeQuestion(e) {
    const index = e.currentTarget.dataset.index
    if (this.data.questions.length <= 1) return
    const questions = [...this.data.questions]
    questions.splice(index, 1)
    this.setData({ questions })
  },

  // ===== 提交 =====

  submit() {
    const data = this.data
    if (data.submitting) return
    if (!data.company.trim()) {
      wx.showToast({ title: '请填写公司名称', icon: 'none' })
      return
    }
    const validQuestions = data.questions.filter(q => q.content && q.content.trim())
    if (!validQuestions.length) {
      wx.showToast({ title: '请至少填写 1 道题目内容', icon: 'none' })
      return
    }

    this.setData({ submitting: true })
    wx.showLoading({ title: '提交中…', mask: true })

    api.contribution.submit({
      company: data.company.trim(),
      jobTitle: data.jobTitle.trim(),
      year: data.year,
      round: data.round || '',
      questions: validQuestions.map(q => ({
        content: q.content.trim(),
        questionType: q.questionType,
        difficulty: q.difficulty,
        tags: q.tagsInput ? q.tagsInput.split(/[,，]/).map(t => t.trim()).filter(Boolean) : []
      }))
    }).then(res => {
      wx.hideLoading()
      const credits = res.data.credits || (this.data.credits + validQuestions.length * 2)
      this.setData({
        submitting: false,
        credits,
        company: '',
        jobTitle: '',
        roundIndex: -1,
        round: '',
        questions: [this.newQuestion()]
      })
      wx.showModal({
        title: '🎉 提交成功！',
        content: `感谢您的贡献！每道题奖励 2 积分，审核通过后再奖励 3 积分。当前积分：${credits}`,
        confirmText: '查看我的贡献',
        cancelText: '继续贡献',
        confirmColor: '#1677ff',
        success: r => {
          if (r.confirm) {
            wx.navigateTo({ url: '/pages/contribution/my' })
          }
          this.loadData()
        }
      })
    }).catch(err => {
      wx.hideLoading()
      this.setData({ submitting: false })
      console.error('提交贡献失败:', err)
    })
  },

  // 跳转我的贡献 / 公司题库
  goMy() {
    wx.navigateTo({ url: '/pages/contribution/my' })
  },

  goCompany() {
    wx.navigateTo({ url: '/pages/contribution/company' })
  }
})
