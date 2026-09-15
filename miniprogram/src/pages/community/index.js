// pages/community/index.js
const api = require('../../api/index')

Page({
  data: {
    loading: false,
    articles: [],
    total: 0,
    page: 1,
    pageSize: 15,
    noMore: false,
    activeCategory: 'all',
    sortBy: 'hot',
    showAIDialog: false,
    aiGenerating: false,
    aiForm: { topic: '', jobCategory: '', jobCategoryLabel: '' },
    categories: [
      { label: '全部', value: 'all' },
      { label: '后端开发', value: 'backend' },
      { label: '前端开发', value: 'frontend' },
      { label: '大数据', value: 'bigdata' },
      { label: 'AI/算法', value: 'ai' },
      { label: '会计/财务', value: 'accounting' },
      { label: '通用', value: 'general' }
    ]
  },

  onLoad() {
    this.loadArticles()
  },

  // 切换分类
  selectCategory(e) {
    const value = e.currentTarget.dataset.value
    if (value === this.data.activeCategory) return
    this.setData({ activeCategory: value, page: 1, articles: [], noMore: false })
    this.loadArticles()
  },

  // 切换排序
  switchSort(e) {
    const sort = e.currentTarget.dataset.sort
    if (sort === this.data.sortBy) return
    this.setData({ sortBy: sort, page: 1, articles: [], noMore: false })
    this.loadArticles()
  },

  // 加载文章列表
  loadArticles() {
    if (this.data.loading) return
    this.setData({ loading: true })

    const params = {
      page: this.data.page,
      pageSize: this.data.pageSize,
      sortBy: this.data.sortBy
    }
    if (this.data.activeCategory !== 'all') {
      params.jobCategory = this.data.activeCategory
    }

    api.community.getArticles(params).then(res => {
      const d = res.data || res
      const list = Array.isArray(d) ? d : (d.list || d.items || [])
      const total = d.total || (Array.isArray(d) ? d.length : 0)

      // 格式化时间
      list.forEach(item => {
        item._time = this.formatTime(item.createdAt)
        item._id = item.articleId || item.id
      })

      const articles = this.data.page === 1 ? list : this.data.articles.concat(list)
      this.setData({
        articles,
        total,
        loading: false,
        noMore: articles.length >= total
      })
    }).catch(err => {
      console.error('加载文章失败:', err)
      this.setData({ loading: false })
      wx.showToast({ title: '加载文章失败', icon: 'none' })
    })
  },

  // 下拉刷新
  onPullDownRefresh() {
    this.setData({ page: 1, articles: [], noMore: false }, () => {
      this.loadArticles()
    })
    wx.stopPullDownRefresh()
  },

  // 上拉加载更多
  onReachBottom() {
    if (this.data.noMore || this.data.loading) return
    this.setData({ page: this.data.page + 1 })
    this.loadArticles()
  },

  // 打开 AI 生成对话框
  openAIDialog() {
    this.setData({ showAIDialog: true })
  },

  // 关闭 AI 对话框
  closeAIDialog() {
    this.setData({ showAIDialog: false, aiForm: { topic: '', jobCategory: '', jobCategoryLabel: '' } })
  },

  // AI 主题输入
  onTopicInput(e) {
    this.setData({ 'aiForm.topic': e.detail.value })
  },

  // AI 岗位方向选择
  onJobCategoryChange(e) {
    const idx = e.detail.value
    const cat = this.data.categories[idx]
    if (cat && cat.value !== 'all') {
      this.setData({
        'aiForm.jobCategory': cat.value,
        'aiForm.jobCategoryLabel': cat.label
      })
    }
  },

  // 提交 AI 生成
  generateArticle() {
    const { topic, jobCategory } = this.data.aiForm
    if (!topic || !jobCategory) {
      wx.showToast({ title: '请填写完整信息', icon: 'none' })
      return
    }

    this.setData({ aiGenerating: true })
    api.community.generateArticle({ topic, jobCategory }).then(() => {
      wx.showToast({ title: 'AI 文章生成成功', icon: 'success' })
      this.setData({
        showAIDialog: false,
        aiForm: { topic: '', jobCategory: '', jobCategoryLabel: '' },
        aiGenerating: false,
        page: 1,
        articles: [],
        noMore: false
      })
      this.loadArticles()
    }).catch(err => {
      console.error('AI 生成失败:', err)
      this.setData({ aiGenerating: false })
      wx.showToast({ title: 'AI 生成失败，请重试', icon: 'none' })
    })
  },

  // 跳转文章详情
  goDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: '/pages/community/detail?id=' + id })
  },

  // 阻止弹窗背景滚动
  preventScroll() {},

  // 格式化时间
  formatTime(val) {
    if (!val) return ''
    const timestamp = typeof val === 'number' && val < 10000000000 ? val * 1000 : val
    const d = new Date(timestamp)
    if (isNaN(d.getTime())) return val
    const pad = n => String(n).padStart(2, '0')
    return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate())
  },

  onShareAppMessage() {
    return {
      title: '知识社区 - 一起学习成长',
      path: '/pages/community/index'
    }
  }
})
