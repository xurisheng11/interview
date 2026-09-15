const api = require('../../api/index')

// 录音管理器
const recorderManager = wx.getRecorderManager()
const innerAudioContext = wx.createInnerAudioContext()

// 语音转文字管理器（微信同声传译插件）：语音模式下边录边转文字写入回答
// 插件未在后台添加时 requirePlugin 会抛错，降级为纯录音
let recognitionManager = null
try {
  recognitionManager = requirePlugin('WechatSI').getRecordRecognitionManager()
} catch (e) {
  recognitionManager = null
}

Page({
  data: {
    // 面试信息
    interviewId: '',
    questions: [],
    currentIndex: 0,
    totalCount: 0,
    remainingTime: 0,
    // wxml 不支持函数调用/getter，计时文本与进度百分比必须入 data
    remainingTimeText: '00:00',
    recordTimeText: '00:00',
    progressPercent: 0,
    currentTypeName: '',
    currentDifficultyName: '',
    
    // 当前问题
    currentQuestion: null,
    answer: '',
    maxLength: 2000,
    // 面试模式：text=纯文字；video=语音模式（录音答题，摄像头视频为 Web 端能力）
    mode: 'text',
    
    // 录音状态
    isRecording: false,
    audioPath: '',
    recordTime: 0,
    isPlaying: false,
    // 录音中的实时识别中间结果（语音转文字反馈）
    liveRecognizeText: '',
    
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

  // 同步计时文本与进度条到 data（供 wxml 渲染）
  syncSessionDisplay() {
    const { remainingTime, recordTime, currentIndex, totalCount, currentQuestion } = this.data
    this.setData({
      remainingTimeText: this.formatTime(remainingTime),
      recordTimeText: this.formatTime(recordTime),
      progressPercent: totalCount === 0 ? 0 : ((currentIndex + 1) / totalCount) * 100,
      currentTypeName: currentQuestion ? this.getTypeName(currentQuestion.type) : '',
      currentDifficultyName: currentQuestion ? this.getDifficultyName(currentQuestion.difficulty) : ''
    })
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
      this.syncSessionDisplay()
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
      this.syncSessionDisplay()
      // NotFoundError = 环境无录音设备（Windows 模拟器常见），给出可操作提示
      const msg = (err && err.errMsg) || ''
      if (msg.indexOf('NotFound') > -1) {
        wx.showToast({
          title: '当前环境无录音设备：模拟器不支持录音，请真机预览测试',
          icon: 'none',
          duration: 4000
        })
      } else {
        wx.showToast({ title: '录音失败，请检查麦克风权限后重试', icon: 'none' })
      }
    })

    // 语音识别管理器：识别结果写入回答框，录音文件仅供回放
    if (recognitionManager) {
      recognitionManager.onStart(() => {
        this.setData({ isRecording: true, recordTime: 0, liveRecognizeText: '' })
        this.syncSessionDisplay()
        this.startRecordTimer()
      })

      recognitionManager.onRecognize((res) => {
        // 流式中间结果，让"说话→文字"实时可见
        this.setData({ liveRecognizeText: res.result || '' })
      })

      recognitionManager.onStop((res) => {
        this.setData({
          isRecording: false,
          audioPath: res.tempFilePath || '',
          liveRecognizeText: ''
        })
        this.stopRecordTimer()
        this.syncSessionDisplay()
        this.appendRecognizedText(res.result)
      })

      recognitionManager.onError((err) => {
        console.error('语音识别错误', err)
        this.setData({ isRecording: false, liveRecognizeText: '' })
        this.stopRecordTimer()
        this.syncSessionDisplay()
        const code = err && err.retcode
        if (code === -30012) return // 无识别任务时调 stop，忽略
        if (code === -30004) {
          wx.showToast({ title: '声音太小或听不清，请重试或手动输入', icon: 'none' })
        } else {
          wx.showToast({ title: '语音识别失败，请重试或手动输入', icon: 'none' })
        }
      })
    }
  },

  // 识别结果追加进回答（多段录音自动拼接），并同步到题目记录
  appendRecognizedText(text) {
    const trimmed = (text || '').trim()
    if (!trimmed) {
      wx.showToast({ title: '未识别到语音内容，请重试或手动输入', icon: 'none' })
      return
    }
    const answer = this.data.answer ? this.data.answer + '\n' + trimmed : trimmed
    this.setData({ answer })
    this.saveCurrentAnswer()
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
        remainingTime: interview.questionTime || 300, // 默认5分钟
        mode: interview.mode || 'text'
      })
      this.syncSessionDisplay()
      
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
        this.syncSessionDisplay()
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
      if (this.usingRecognition) {
        recognitionManager.stop()
      } else {
        recorderManager.stop()
      }
    } else if (recognitionManager) {
      // 语音模式首选语音转文字：识别结果自动写入回答（单次上限 60 秒）
      this.usingRecognition = true
      recognitionManager.start({ duration: 60000, lang: 'zh_CN' })
    } else {
      this.usingRecognition = false
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
      this.syncSessionDisplay()
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
        audioPath: this.data.questions[newIndex].audioPath || '',
        liveRecognizeText: ''
      })
      this.syncSessionDisplay()
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
        audioPath: this.data.questions[newIndex].audioPath || '',
        liveRecognizeText: ''
      })
      this.syncSessionDisplay()
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
  }
})
