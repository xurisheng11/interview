const api = require('../../api/index')

Page({
  data: {
    keyword: '',
    selectedCategory: 'all',
    questions: [],
    loading: false,
    categories: [
      { value: 'all', name: '全部' },
      { value: 'technical', name: '技术' },
      { value: 'behavioral', name: '行为' },
      { value: 'hr', name: 'HR' },
      { value: 'project', name: '项目' }
    ]
  },

  onLoad() {
    this.loadQuestions()
  },

  onSearchInput(e) {
    this.setData({ keyword: e.detail.value })
    this.loadQuestions()
  },

  selectCategory(e) {
    this.setData({ 
      selectedCategory: e.currentTarget.dataset.value 
    })
    this.loadQuestions()
  },

  loadQuestions() {
    const { keyword, selectedCategory } = this.data
    
    this.setData({ loading: true })
    
    api.question.list().then(res => {
      // 后端返回 { list: [...], total: 100 }
      let questions = res.data?.list || []
      
      // 筛选（后端返回 type 字段）
      if (selectedCategory !== 'all') {
        questions = questions.filter(q => q.type === selectedCategory)
      }
      
      if (keyword) {
        questions = questions.filter(q => 
          (q.content || '').toLowerCase().includes(keyword.toLowerCase())
        )
      }
      
      const mapped = questions.map(q => ({
        ...q,
        typeName: this.getTypeName(q.type),
        difficultyName: this.getDifficultyName(q.difficulty)
      }))
      this.setData({ loading: false })
      // 列表接口不返回收藏态，登录后用收藏集合反查标记星标
      this.syncCollectState(mapped)
    }).catch(err => {
      console.error('获取题库失败:', err)
      this.setData({ loading: false })
      wx.showToast({
        title: '获取题库失败',
        icon: 'none'
      })
    })
  },

  // 用收藏集合为题目列表标记收藏态
  syncCollectState(questions) {
    if (!wx.getStorageSync('token')) {
      this.setData({ questions })
      return
    }
    api.profile.getCollections().then(res => {
      const data = res.data || {}
      const qList = Array.isArray(data.questions) ? data.questions : []
      const collectedIds = new Set(qList.map(q => q.questionId))
      this.setData({
        questions: questions.map(q => ({ ...q, isCollected: collectedIds.has(q.questionId) }))
      })
    }).catch(() => {
      this.setData({ questions })
    })
  },

  getTypeName(type) {
    const map = {
      'technical': '技术',
      'behavioral': '行为',
      'hr': 'HR',
      'project': '项目'
    }
    return map[type] || type
  },

  getDifficultyName(difficulty) {
    const map = {
      'easy': '简单',
      'medium': '中等',
      'hard': '困难'
    }
    return map[difficulty] || difficulty
  },

  viewQuestion(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/questions/detail/index?id=${id}`
    })
  },

  // 列表快捷收藏
  toggleCollect(e) {
    if (!wx.getStorageSync('token')) {
      wx.navigateTo({ url: '/pages/login/login' })
      return
    }
    const { id, index } = e.currentTarget.dataset
    const key = `questions[${index}].isCollected`
    const next = !this.data.questions[index].isCollected
    this.setData({ [key]: next })
    // 收藏走 POST、取消走 DELETE（后端非切换式接口）
    const req = next ? api.question.collect(id) : api.question.uncollect(id)
    req.then(() => {
      wx.showToast({ title: next ? '已收藏' : '已取消收藏', icon: 'none' })
    }).catch(() => {
      this.setData({ [key]: !next })
      wx.showToast({ title: '操作失败', icon: 'none' })
    })
  }
})
