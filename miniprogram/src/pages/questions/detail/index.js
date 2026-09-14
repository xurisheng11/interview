const api = require('../../../api/index')

Page({
  data: {
    questionId: null,
    question: {},
    loading: true,
    error: null
  },

  onLoad(options) {
    if (options.id) {
      this.setData({ questionId: options.id })
      this.loadQuestion()
    }
  },

  loadQuestion() {
    this.setData({ loading: true, error: null })
    
    api.question.get(this.data.questionId).then(res => {
      this.setData({ 
        question: res.data || {},
        loading: false 
      })
    }).catch(err => {
      console.error('加载题目失败', err)
      this.setData({ 
        loading: false,
        error: '加载失败，请重试'
      })
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
    })
  },

  getTypeName(type) {
    const map = {
      'technical': '技术问题',
      'behavioral': '行为问题',
      'hr': 'HR问题',
      'project': '项目问题',
      'basic': '基础问题',
      'advanced': '高级问题'
    }
    return map[type] || '通用问题'
  },

  getDifficultyName(difficulty) {
    const map = {
      'easy': '简单',
      'medium': '中等',
      'hard': '困难'
    }
    return map[difficulty] || '中等'
  },

  // 开始练习
  startPractice() {
    wx.showModal({
      title: '开始练习',
      content: '将以此题目开始一次模拟面试练习',
      success: (res) => {
        if (res.confirm) {
          wx.navigateBack()
        }
      }
    })
  }
})
