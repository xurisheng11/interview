const api = require('../../api/index')

// 录音管理器
const recorderManager = wx.getRecorderManager()
const innerAudioContext = wx.createInnerAudioContext()

Page({
  data: {
    // 面试信息
    interviewId: '',
    questions: [],
    currentIndex: 0,
    totalCount: 0,
    remainingTime: 0,
    
    // 当前问题
    currentQuestion: null,
    answer: '',
    maxLength: 2000,
    
    // 录音状态
    isRecording: false,
    audioPath: '',
    recordTime: 0,
    isPlaying: false,
    
    // UI状态
    submitting: false,
    showCompleteModal: false,
    
    // 定时器
    timer: null
  },

  // 格式化时间
  formatTime(seconds) {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  },

  // 获取类型名称
  getTypeName(type) {
    const map = {
      'technical': '技术问题',
      'behavioral': '行为问题',
      'situational': '情景问题',
      'competency': '能力问题'
    }
    return map[type] || '通用问题'
  },

  // 获取难度名称
  getDifficultyName(difficulty) {
    const map = {
      'easy': '简单',
      'medium': '中等',
      'hard': '困难'
    }
    return map[difficulty] || '中等'
  },

  onLoad(options) {
    if (!options.id) {
      wx.showToast({ title: '参数错误', icon: 'none' })
      wx.navigateBack()
      return
    }

    this.setData({ interviewId: options.id })
    this.initRecorder()
    this.loadInterview()
  },

  onUnload() {
    this.clearTimer()
    innerAudioContext.destroy()
  },

  // 初始化录音管理器
  initRecorder() {
    recorderManager.onStart(() => {
      this.setData({ isRecording: true, recordTime: 0 })
      this.startRecordTimer()
    })

    recorderManager.onStop((res) => {
      this.setData({
        isRecording: false,
        audioPath: res.tempFilePath
      })
      this.stopRecordTimer()
    })

    recorderManager.onError((err) => {
      console.error('录音错误', err)
      this.setData({ isRecording: false })
      this.stopRecordTimer()
      wx.showToast({ title: '录音失败', icon: 'none' })
    })
  },

  // 加载面试数据
  loadInterview() {
    wx.showLoading({ title: '加载中...' })
    
    api.interview.get(this.data.interviewId).then(res => {
      wx.hideLoading()
      const interview = res.data
      
      // 使用后端返回的题目数据
      const questions = interview.questions || []
      
      if (questions.length === 0) {
        wx.showToast({ title: '暂无题目数据', icon: 'none' })
        setTimeout(() => wx.navigateBack(), 1500)
        return
      }
      
      this.setData({
        questions: questions,
        totalCount: questions.length,
        currentQuestion: questions[0] || null,
        remainingTime: interview.questionTime || 300 // 默认5分钟
      })
      
      this.startTimer()
    }).catch(err => {
      wx.hideLoading()
      wx.showToast({ title: '加载失败', icon: 'none' })
      setTimeout(() => wx.navigateBack(), 1500)
    })
  },

  // 计时器
  startTimer() {
    this.clearTimer()
    this.data.timer = setInterval(() => {
      const remaining = this.data.remainingTime - 1
      if (remaining <= 0) {
        this.clearTimer()
        this.handleTimeUp()
      } else {
        this.setData({ remainingTime: remaining })
      }
    }, 1000)
  },

  clearTimer() {
    if (this.data.timer) {
      clearInterval(this.data.timer)
      this.data.timer = null
    }
  },

  handleTimeUp() {
    wx.showToast({ title: '时间到！', icon: 'none' })
    // 自动提交当前答案并进入下一题
    this.submitCurrentAnswer().then(() => {
      if (this.data.currentIndex < this.data.totalCount - 1) {
        this.nextQuestion()
      } else {
        this.completeInterview()
      }
    })
  },

  // 答案输入
  onAnswerInput(e) {
    this.setData({ answer: e.detail.value })
  },

  // 切换录音
  toggleRecord() {
    if (this.data.isRecording) {
      recorderManager.stop()
    } else {
      recorderManager.start({
        duration: 60000, // 60秒
        sampleRate: 16000,
        numberOfChannels: 1,
        encodeBitRate: 48000,
        format: 'mp3'
      })
    }
  },

  // 播放录音
  playAudio() {
    if (!this.data.audioPath) return

    if (this.data.isPlaying) {
      innerAudioContext.stop()
      this.setData({ isPlaying: false })
    } else {
      innerAudioContext.src = this.data.audioPath
      innerAudioContext.play()
      this.setData({ isPlaying: true })

      innerAudioContext.onEnded(() => {
        this.setData({ isPlaying: false })
      })
    }
  },

  // 录音计时器
  startRecordTimer() {
    this.recordTimer = setInterval(() => {
      this.setData({ recordTime: this.data.recordTime + 1 })
    }, 1000)
  },

  stopRecordTimer() {
    if (this.recordTimer) {
      clearInterval(this.recordTimer)
      this.recordTimer = null
    }
  },

  // 上一题
  prevQuestion() {
    if (this.data.currentIndex > 0) {
      // 保存当前答案
      this.saveCurrentAnswer()
      
      const newIndex = this.data.currentIndex - 1
      this.setData({
        currentIndex: newIndex,
        currentQuestion: this.data.questions[newIndex],
        answer: this.data.questions[newIndex].userAnswer || '',
        audioPath: this.data.questions[newIndex].audioPath || ''
      })
    }
  },

  // 下一题
  nextQuestion() {
    // 保存当前答案
    this.saveCurrentAnswer()

    if (this.data.currentIndex < this.data.totalCount - 1) {
      const newIndex = this.data.currentIndex + 1
      this.setData({
        currentIndex: newIndex,
        currentQuestion: this.data.questions[newIndex],
        answer: this.data.questions[newIndex].userAnswer || '',
        audioPath: this.data.questions[newIndex].audioPath || ''
      })
    } else {
      // 最后一题，提交
      this.completeInterview()
    }
  },

  // 跳过
  skipQuestion() {
    wx.showModal({
      title: '确认跳过',
      content: '确定要跳过这道题吗？',
      success: (res) => {
        if (res.confirm) {
          this.nextQuestion()
        }
      }
    })
  },

  // 保存当前答案
  saveCurrentAnswer() {
    const { currentIndex, questions, answer, audioPath } = this.data
    questions[currentIndex].userAnswer = answer
    questions[currentIndex].audioPath = audioPath
    this.setData({ questions })
  },

  // 提交当前答案
  submitCurrentAnswer() {
    return new Promise((resolve) => {
      const { interviewId, currentIndex, questions } = this.data
      
      api.interview.submitAnswer(interviewId, {
        questionIndex: currentIndex,
        answer: questions[currentIndex].userAnswer || this.data.answer
      }).then(() => {
        resolve()
      }).catch(() => {
        resolve() // 即使失败也继续
      })
    })
  },

  // 完成面试
  completeInterview() {
    this.setData({ 
      submitting: true,
      showCompleteModal: true 
    })

    // 提交所有答案
    this.saveCurrentAnswer()

    api.interview.complete(this.data.interviewId).then(res => {
      this.setData({ submitting: false })
      
      // 跳转到报告页
      setTimeout(() => {
        this.setData({ showCompleteModal: false })
        wx.redirectTo({
          url: `/pages/interview/detail?id=${this.data.interviewId}&new=1`
        })
      }, 1500)
    }).catch(err => {
      this.setData({ submitting: false, showCompleteModal: false })
      wx.showToast({ title: '提交失败', icon: 'none' })
    })
  },

  closeModal() {
    // 阻止关闭
  },

  // 计算进度
  get progressPercent() {
    if (this.data.totalCount === 0) return 0
    return ((this.data.currentIndex + 1) / this.data.totalCount) * 100
  }
})
