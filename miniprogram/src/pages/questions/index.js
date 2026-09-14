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
      
      this.setData({ 
        questions,
        loading: false 
      })
    }).catch(err => {
      console.error('获取题库失败:', err)
      this.setData({ loading: false })
      wx.showToast({
        title: '获取题库失败',
        icon: 'none'
      })
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
  }
})
