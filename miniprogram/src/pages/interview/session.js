const api = require('../../api/index')

// 录音管理器
const recorderManager = wx.getRecorderManager()
const innerAudioContext = wx.createInnerAudioContext()

// 常见中文口头禅（与 Web 端 speechMixin 对齐，转写文本出现≥2次即上报）
const VERBAL_TICS = [
  '然后', '这个', '那个', '嗯', '呃', '啊', '就是', '就是说',
  '的话', '其实', '基本上', '大概', '可能', '应该', '好像',
  '对吧', '是吧', '好吗', '好吧', '所以'
]

// 语音转文字管理器（微信同声传译插件）：插件仅对非个人主体小程序开放，
// 个人主体账号无法添加（app.json 已不声明插件），requirePlugin 失败自动降级为纯录音
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
    // 面试模式：text=纯文字；video=语音模式（录音答题）；video_call=视频面试（摄像头预览+录音答题）
    mode: 'text',
    // isVoice=需录音转写与表达分析（video/video_call）；isCamera=展示摄像头预览（video_call）
    isVoice: false,
    isCamera: false,
    // 摄像头不可用（用户拒绝授权或设备错误）时降级提示
    cameraBlocked: false,
    // 摄像头悬浮窗是否最小化（避免遮挡题目/回答）
    cameraMinimized: false,
    
    // 录音状态
    isRecording: false,
    audioPath: '',
    recordTime: 0,
    isPlaying: false,
    // 录音停止后后端转写中（腾讯云一句话识别）
    isTranscribing: false,
    // 录音中的实时识别中间结果（语音转文字反馈）
    liveRecognizeText: '',
    // 语音转文字能力是否可用（个人主体小程序插件不可用，降级为纯录音）
    speechToText: false,
    
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
    // 按实际能力渲染文案：插件可用才承诺"语音转文字"（个人主体小程序插件不可用，降级纯录音）
    this.setData({ speechToText: !!recognitionManager })
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
    wx.setKeepScreenOn({ keepScreenOn: false })
  },

  // 初始化录音管理器
  initRecorder() {
    recorderManager.onStart(() => {
      this.setData({ isRecording: true, recordTime: 0 })
      this.markFirstSpeech()
      this.syncSessionDisplay()
      this.startRecordTimer()
    })

    recorderManager.onStop((res) => {
      this.accumulateSpeak(res && res.duration)
      this.setData({
        isRecording: false,
        audioPath: res.tempFilePath
      })
      this.stopRecordTimer()
      this.syncSessionDisplay()
      // 语音/视频模式：录音停止后上传后端转文字（腾讯云一句话识别），结果追加进回答
      if (this.data.isVoice && res.tempFilePath) {
        this.transcribeAudio(res.tempFilePath)
      }
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
        this.markFirstSpeech()
        this.syncSessionDisplay()
        this.startRecordTimer()
      })

      recognitionManager.onRecognize((res) => {
        // 流式中间结果，让"说话→文字"实时可见
        this.setData({ liveRecognizeText: res.result || '' })
      })

      recognitionManager.onStop((res) => {
        this.accumulateSpeak(null)
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

  // 记录本题首次开口时间（用于思考时长 = 首次开口 - 题目展示）
  markFirstSpeech() {
    const qs = this.data.questions
    const cur = qs[this.data.currentIndex]
    if (cur && !cur.firstSpeechAt) cur.firstSpeechAt = Date.now()
  },

  // 累计本题说话时长（秒）；durationMs 为录音返回毫秒，无则用计时器兜底
  accumulateSpeak(durationMs) {
    const qs = this.data.questions
    const cur = qs[this.data.currentIndex]
    if (!cur) return
    const secs = durationMs ? Math.round(durationMs / 1000) : this.data.recordTime
    if (secs > 0) cur.speakSeconds = (cur.speakSeconds || 0) + secs
    this.setData({ questions: qs })
  },

  // 从转写文本检测口头禅（出现≥2次）
  detectVerbalTics(text) {
    if (!text) return []
    const found = []
    for (let i = 0; i < VERBAL_TICS.length; i++) {
      const tic = VERBAL_TICS[i]
      const count = text.split(tic).length - 1
      if (count >= 2) found.push(tic)
    }
    return found
  },

  // 构建语音表达指标（语速/时长/思考时长/口头禅）
  buildSpeechMetrics(cur, answer) {
    const duration = (cur && cur.speakSeconds) || 0
    const thinkDuration = (cur && cur.firstSpeechAt && this.questionShownAt)
      ? Math.max(0, Math.round((cur.firstSpeechAt - this.questionShownAt) / 1000))
      : 0
    // 有效字数：中文 + 字母数字
    const chars = (answer || '').replace(/[^\u4e00-\u9fa5a-zA-Z0-9]/g, '').length
    const speechRate = duration > 0 ? Math.round(chars / (duration / 60)) : 0
    return {
      speechRate: speechRate,
      pauseCount: 0, // 一句话识别无法测停顿
      duration: duration,
      thinkDuration: thinkDuration,
      verbalTics: this.detectVerbalTics(answer)
    }
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
        mode: interview.mode || 'text',
        isVoice: interview.mode === 'video' || interview.mode === 'video_call',
        isCamera: interview.mode === 'video_call'
      })
      // 视频面试：保持屏幕常亮，避免答题中途息屏中断摄像头
      if (interview.mode === 'video_call') {
        wx.setKeepScreenOn({ keepScreenOn: true })
      }
      this.questionShownAt = Date.now()
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

  // 上传录音到后端转写为文字，追加进回答框
  transcribeAudio(filePath) {
    const app = getApp()
    const token = wx.getStorageSync('token')
    this.setData({ isTranscribing: true })
    wx.uploadFile({
      url: app.globalData.apiBaseUrl + '/asr/transcribe',
      filePath: filePath,
      name: 'file',
      header: { Authorization: 'Bearer ' + token },
      success: (res) => {
        let text = ''
        let errMsg = ''
        try {
          const data = JSON.parse(res.data)
          if (data.code === 200 && data.data) {
            text = data.data.text || ''
          } else {
            errMsg = data.message || '转写失败'
          }
        } catch (e) {
          errMsg = '转写响应解析失败'
        }
        if (errMsg) {
          wx.showToast({ title: errMsg + '，可手动输入', icon: 'none' })
          return
        }
        if (text) {
          const answer = this.data.answer ? this.data.answer + '\n' + text : text
          this.setData({ answer })
          wx.showToast({ title: '已转写并追加到回答', icon: 'none' })
        } else {
          wx.showToast({ title: '未识别到语音，可重录或手动输入', icon: 'none' })
        }
      },
      fail: () => {
        wx.showToast({ title: '转写请求失败，可手动输入', icon: 'none' })
      },
      complete: () => {
        this.setData({ isTranscribing: false })
      }
    })
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

  // 摄像头初始化成功：清除降级提示
  onCameraInitDone() {
    if (this.data.cameraBlocked) this.setData({ cameraBlocked: false })
  },

  // 摄像头错误（拒绝授权/设备不可用）：不阻断面试，降级为仅语音答题
  onCameraError(e) {
    console.error('摄像头错误', e && e.detail)
    this.setData({ cameraBlocked: true })
  },

  // 切换摄像头悬浮窗最小化
  toggleCameraMin() {
    this.setData({ cameraMinimized: !this.data.cameraMinimized })
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
      this.questionShownAt = Date.now()
      this.syncSessionDisplay()
    }
  },

  // 下一题：先提交当前答案给后端 AI 点评（报告分数/逐题点评依赖这一步），再前进
  nextQuestion() {
    if (this.data.submitting) return
    this.saveCurrentAnswer()

    const cur = this.data.questions[this.data.currentIndex] || {}
    const needSubmit = (cur.userAnswer || '').trim() !== '' && !cur.submitted
    if (!needSubmit) {
      this.goNext()
      return
    }

    this.setData({ submitting: true })
    wx.showLoading({ title: 'AI 正在点评…', mask: true })
    this.submitCurrentAnswer().then(() => {
      wx.hideLoading()
      this.setData({ submitting: false })
      this.goNext()
    })
  },

  // 前进到下一题（或最后一题后完成面试）
  goNext() {
    if (this.data.currentIndex < this.data.totalCount - 1) {
      const newIndex = this.data.currentIndex + 1
      this.setData({
        currentIndex: newIndex,
        currentQuestion: this.data.questions[newIndex],
        answer: this.data.questions[newIndex].userAnswer || '',
        audioPath: this.data.questions[newIndex].audioPath || '',
        liveRecognizeText: ''
      })
      this.questionShownAt = Date.now()
      this.syncSessionDisplay()
    } else {
      // 最后一题，提交
      this.completeInterview()
    }
  },

  // 跳过：不提交本题答案，报告中按跳过计 0 分
  skipQuestion() {
    wx.showModal({
      title: '确认跳过',
      content: '确定要跳过这道题吗？跳过的题不计入成绩。',
      success: (res) => {
        if (res.confirm) {
          if (this.data.submitting) return
          this.saveCurrentAnswer()
          this.goNext()
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

  // 提交当前答案（空答案/已提交过则跳过；失败不阻断流程，完成时还会重试）
  submitCurrentAnswer() {
    return new Promise((resolve) => {
      const { interviewId, currentIndex, questions } = this.data
      const cur = questions[currentIndex]
      if (!cur) { resolve(); return }
      const answer = (cur.userAnswer || this.data.answer || '').trim()
      if (!answer || cur.submitted) { resolve(); return }

      const payload = {
        questionIndex: currentIndex,
        answer: answer
      }
      // 语音/视频模式：附带表达指标，后端会据此生成语速/自信度/口头禅/普通话等表达点评
      if (this.data.isVoice) {
        payload.nonVerbalMetrics = this.buildSpeechMetrics(cur, answer)
      }

      api.interview.submitAnswer(interviewId, payload).then(() => {
        const qs = this.data.questions
        if (qs[currentIndex]) {
          qs[currentIndex].submitted = true
          this.setData({ questions: qs })
        }
        resolve()
      }).catch(() => {
        resolve() // 即使失败也继续
      })
    })
  },

  // 完成面试：先把最后一题答案提交（含 AI 点评），再触发报告生成
  completeInterview() {
    this.setData({ 
      submitting: true,
      showCompleteModal: true 
    })

    // 提交所有答案
    this.saveCurrentAnswer()

    this.submitCurrentAnswer().then(() => {
      return api.interview.complete(this.data.interviewId)
    }).then(res => {
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
