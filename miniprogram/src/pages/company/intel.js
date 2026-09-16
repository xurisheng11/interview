const api = require('../../api/index')
const { mdToHtml } = require('../../utils/markdown')

// 岗位方向选项（与 UI 前端一致）
const JOB_TITLES = [
  '后端开发', '前端开发', '全栈开发', '移动端开发', '算法工程师', '大数据工程师',
  '测试工程师', '运维/DevOps', '网络安全', '产品经理', 'UI/UX设计师', '平面设计师',
  '数据分析师', '市场营销', '运营专员', '新媒体运营', '人力资源', '行政管理',
  '销售/商务', '会计/财务', '项目管理', '客户服务'
]

// 每页展示题目数
const PAGE_SIZE = 11
const PAGE_INCREMENT = 10

Page({
  data: {
    company: '',
    jobTitleIndex: -1,
    jobTitles: JOB_TITLES,
    loading: false,
    loadingPercent: 0,
    result: null,
    // 题目筛选
    qFilter: 'all',
    filteredQuestions: [],
    visibleQuestions: [],
    visibleCount: PAGE_SIZE,
    hasMore: false,
    // 题目详情弹层
    showQuestion: false,
    activeQuestion: null,
    answerLoading: false,
    // 示例
    hotCompanies: ['字节跳动', '腾讯', '阿里巴巴', '美团', '京东', '百度', '华为', '网易', '滴滴', '快手'],
    examples: [
      { company: '字节跳动', desc: '以算法题著称，多轮技术面' },
      { company: '腾讯', desc: '强调项目深度和技术原理' },
      { company: '阿里巴巴', desc: '注重系统设计和业务理解' },
      { company: '华为', desc: '笔试+技术面+HR面流程完整' }
    ]
  },

  onLoad() {
    // 参考答案缓存：必须初始化，否则 openQuestion 访问 undefined 报错导致弹层打不开
    this._answerCache = {}
  },

  // 公司名输入
  onCompanyInput(e) {
    this.setData({ company: e.detail.value })
  },

  // 岗位选择
  onJobTitleChange(e) {
    const index = Number(e.detail.value)
    this.setData({
      jobTitleIndex: index,
      jobTitle: index > -1 ? JOB_TITLES[index] : ''
    })
  },

  // 搜索公司情报
  search() {
    const company = this.data.company.trim()
    if (!company) {
      wx.showToast({ title: '请输入公司名称', icon: 'none' })
      return
    }
    if (this.data.loading) return

    this.setData({
      loading: true,
      loadingPercent: 0,
      result: null,
      qFilter: 'all',
      visibleCount: PAGE_SIZE
    })

    // 进度条模拟
    this._timer = setInterval(() => {
      if (this.data.loadingPercent < 90) {
        this.setData({ loadingPercent: this.data.loadingPercent + 10 })
      }
    }, 800)

    const jobTitle = this.data.jobTitle || ''
    api.company.getIntel(company, jobTitle).then(res => {
      clearInterval(this._timer)
      const result = res.data || res
      this.setData({
        loading: false,
        loadingPercent: 100,
        result
      })
      this.applyFilter()
    }).catch(() => {
      clearInterval(this._timer)
      this.setData({ loading: false })
    })
  },

  // 快捷搜索
  quickSearch(e) {
    const company = e.currentTarget.dataset.company
    this.setData({ company })
    this.search()
  },

  // 题目类型筛选
  setFilter(e) {
    this.setData({
      qFilter: e.currentTarget.dataset.filter,
      visibleCount: PAGE_SIZE
    })
    this.applyFilter()
  },

  // 应用筛选
  applyFilter() {
    const result = this.data.result
    if (!result) return
    const questions = result.questions || []
    const filtered = this.data.qFilter === 'all'
      ? questions
      : questions.filter(q => q.type === this.data.qFilter)
    this.setData({
      filteredQuestions: filtered,
      visibleQuestions: filtered.slice(0, this.data.visibleCount),
      hasMore: this.data.visibleCount < filtered.length
    })
  },

  // 查看更多
  loadMore() {
    this.setData({ visibleCount: this.data.visibleCount + PAGE_INCREMENT })
    this.applyFilter()
  },

  // 打开题目详情弹层（按下标取题，避免长文本经 dataset 回传出现转义问题）
  openQuestion(e) {
    const index = Number(e.currentTarget.dataset.index)
    const question = this.data.visibleQuestions[index]
    if (!question) return
    // 答案缓存命中则直接展示
    const content = question.content || ''
    const cached = (this._answerCache || {})[content] || ''
    this.setData({
      showQuestion: true,
      activeQuestion: {
        ...question,
        answer: cached,
        answerHtml: cached ? mdToHtml(cached) : '',
        answerLoaded: !!cached
      }
    })
  },

  // 关闭题目详情
  closeQuestion() {
    this.setData({ showQuestion: false })
  },

  // AI 生成参考答案
  loadAnswer() {
    const q = this.data.activeQuestion
    if (!q || this.data.answerLoading) return
    this.setData({ answerLoading: true })

    api.company.getQuestionAnswer(
      this.data.result.company,
      this.data.result.jobTitle || '',
      q.content
    ).then(res => {
      const answer = (res.data || res).answer || ''
      if (!this._answerCache) this._answerCache = {}
      this._answerCache[q.content] = answer
      this.setData({
        answerLoading: false,
        // AI 答案带 Markdown（## / ** / 代码块），走 rich-text 渲染，不再把星号直接露出来
        activeQuestion: { ...q, answer, answerHtml: mdToHtml(answer), answerLoaded: true }
      })
    }).catch(() => {
      this.setData({ answerLoading: false })
    })
  },

  // 发起模拟面试（带上岗位）
  startInterview() {
    const jobTitle = this.data.result.jobTitle || ''
    wx.navigateTo({
      url: '/pages/interview/create' + (jobTitle ? '?jobTitle=' + encodeURIComponent(jobTitle) : '')
    })
  },

  // 阻止冒泡
  noop() {},

  onUnload() {
    if (this._timer) clearInterval(this._timer)
  }
})
