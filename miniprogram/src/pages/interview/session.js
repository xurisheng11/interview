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
    // 视频面试自动识别：进题即自动录音，说话自动转写追加，无需手动点麦克风
    autoRecord: false,
    // 因摄像头占用麦克风而自动暂停画面（部分安卓机型冲突）
    cameraPausedForMic: false,
    
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
    this.autoLoop = false
    if (this.data.isRecording) {
      try { recorderManager.stop() } catch (e) {}
    }
    innerAudioContext.destroy()
    wx.setKeepScreenOn({ keepScreenOn: false })
  },

  // 初始化录音管理器
  initRecorder() {
    recorderManager.onStart(() => {
      this.cameraConflictRetried = false // 启动成功后，允许下次冲突再自动恢复一次
      this.setData({ isRecording: true, recordTime: 0 })
      this.markFirstSpeech()
      this.syncSessionDisplay()
      this.startRecordTimer()
    })

    recorderManager.onStop((res) => {
      this.setData({
        isRecording: false,
        audioPath: res.tempFilePath
      })
      this.stopRecordTimer()
      this.syncSessionDisplay()
      // 语音/视频模式：录音停止后上传后端转文字（腾讯云一句话识别），结果写回录音时所在题目
      if (this.data.isVoice && res.tempFilePath) {
        this.transcribeAudio(res.tempFilePath, {
          segMs: res && res.duration,
          index: this.data.currentIndex,
          silent: this.autoLoop, // 自动识别循环中静音段属正常，不打扰
          onDone: () => {
            // 自动识别循环：仍在同一题且循环未停 → 接着录下一段（单次上限 60 秒）
            if (this.autoLoop && this.data.currentIndex === this.autoLoopIndex) {
              this.beginRecord()
            }
          }
        })
      } else if (this.autoLoop && this.data.currentIndex === this.autoLoopIndex) {
        this.beginRecord()
      }
    })

    recorderManager.onError((err) => {
      console.error('录音错误', err)
      this.setData({ isRecording: false })
      this.stopRecordTimer()
      this.syncSessionDisplay()
      const msg = (err && err.errMsg) || ''
      // NotFoundError = 环境无录音设备（Windows 模拟器常见），给出可操作提示
      if (msg.indexOf('NotFound') > -1) {
        this.autoLoop = false
        this.setData({ autoRecord: false })
        wx.showToast({
          title: '当前环境无录音设备：模拟器不支持录音，请真机预览测试',
          icon: 'none',
          duration: 4000
        })
        return
      }
      // 麦克风未授权：引导到设置页开启后重试
      if (msg.indexOf('auth') > -1 || msg.indexOf('deny') > -1 || msg.indexOf('permission') > -1) {
        this.autoLoop = false
        this.setData({ autoRecord: false })
        wx.showModal({
          title: '未获得麦克风权限',
          content: '语音/视频面试需要麦克风权限才能自动识别你的回答，去设置中开启？',
          confirmText: '去开启',
          success: (r) => { if (r.confirm) wx.openSetting() }
        })
        return
      }
      // 部分安卓机型 camera 组件会占用麦克风导致录音失败：自动暂停画面并重试一次
      if (this.data.isCamera && !this.data.cameraMinimized && !this.cameraConflictRetried) {
        this.cameraConflictRetried = true
        this.setData({ cameraMinimized: true, cameraPausedForMic: true })
        wx.showToast({ title: '摄像头占用了麦克风，已暂停画面重试录音', icon: 'none', duration: 3000 })
        setTimeout(() => {
          if (!this.data.isRecording) this.beginRecord()
        }, 800)
        return
      }
      // 兜底：展示真实错误便于定位
      this.autoLoop = false
      this.setData({ autoRecord: false })
      wx.showModal({
        title: '录音失败',
        content: (msg ? ('错误信息：' + msg) : '请检查麦克风权限后重试') + '\n可点麦克风重试，或手动输入文字作答',
        showCancel: false
      })
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

  // 累计指定题目说话时长（秒）；仅识别出文字的录音段才计入，避免静音段拉低语速
  accumulateSpeak(durationMs, index) {
    const qs = this.data.questions
    const cur = qs[index != null ? index : this.data.currentIndex]
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
      // 视频面试：进入即开启自动识别，直接说话即可
      this.startAutoRecord()
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
    this.stopAutoRecord()
    // 等进行中的转写落定再提交，避免最后一段语音丢失
    this.waitTranscripts().then(() => {
      this.saveCurrentAnswer()
      return this.submitCurrentAnswer()
    }).then(() => {
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

  // 开始一段录音（手动/自动共用入口）
  beginRecord() {
    if (this.data.isRecording) return
    if (recognitionManager) {
      this.usingRecognition = true
      recognitionManager.start({ duration: 60000, lang: 'zh_CN' })
    } else {
      this.usingRecognition = false
      recorderManager.start({
        duration: 60000, // 60秒，到点自动停止并触发转写（自动模式会循环续录）
        sampleRate: 16000,
        numberOfChannels: 1,
        encodeBitRate: 48000,
        format: 'mp3'
      })
    }
  },

  // 停止当前段录音（停止后触发转写）
  endRecord() {
    if (!this.data.isRecording) return
    if (this.usingRecognition) {
      recognitionManager.stop()
    } else {
      recorderManager.stop()
    }
  },

  // 麦克风按钮：录音中=暂停（自动循环一并停止）；空闲=开始/恢复
  toggleRecord() {
    if (this.data.isRecording) {
      this.autoLoop = false
      this.setData({ autoRecord: false })
      this.endRecord()
    } else {
      // 视频面试手动恢复时重新进入自动识别循环
      if (this.data.isCamera && !this.autoLoop) {
        this.autoLoop = true
        this.autoLoopIndex = this.data.currentIndex
        this.setData({ autoRecord: true })
      }
      this.beginRecord()
    }
  },

  // 视频面试：进题自动开启识别循环（说话即转文字，无需点麦克风）
  startAutoRecord() {
    if (!this.data.isCamera || this.autoLoop) return
    // 先确保麦克风授权，避免与摄像头授权弹窗竞争导致录音失败
    wx.authorize({
      scope: 'scope.record',
      success: () => {
        if (this.autoLoop) return
        this.autoLoop = true
        this.autoLoopIndex = this.data.currentIndex
        this.setData({ autoRecord: true })
        this.beginRecord()
      },
      fail: () => {
        wx.showModal({
          title: '未获得麦克风权限',
          content: '视频面试需要麦克风权限才能自动识别你的回答，去设置中开启？',
          confirmText: '去开启',
          success: (r) => { if (r.confirm) wx.openSetting() }
        })
      }
    })
  },

  // 停止自动识别循环（切题/交卷前）；在录的那一段仍会转写并写回原题目
  stopAutoRecord() {
    this.autoLoop = false
    this.setData({ autoRecord: false })
    if (this.data.isRecording) this.endRecord()
  },

  // 等待所有进行中的转写完成（转写是异步上传，交卷前必须落定）
  waitTranscripts() {
    return new Promise((resolve) => {
      const check = () => {
        if ((this.pendingTrans || 0) <= 0) { resolve(); return }
        setTimeout(check, 200)
      }
      check()
    })
  },

  // 上传录音到后端转写为文字，写回录音时所在题目的回答
  // opts: { segMs 本段录音时长毫秒, index 录音时题目下标, silent 静默模式, onDone 完成回调 }
  transcribeAudio(filePath, opts) {
    const o = opts || {}
    const app = getApp()
    const token = wx.getStorageSync('token')
    this.pendingTrans = (this.pendingTrans || 0) + 1
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
          if (!o.silent) wx.showToast({ title: errMsg + '，可手动输入', icon: 'none' })
          return
        }
        if (text) {
          this.accumulateSpeak(o.segMs || null, o.index)
          this.mergeAnswer(o.index, text)
          if (!o.silent) wx.showToast({ title: '已转写并追加到回答', icon: 'none' })
        } else if (!o.silent) {
          wx.showToast({ title: '未识别到语音，可重录或手动输入', icon: 'none' })
        }
      },
      fail: () => {
        if (!o.silent) wx.showToast({ title: '转写请求失败，可手动输入', icon: 'none' })
      },
      complete: () => {
        this.pendingTrans--
        this.setData({ isTranscribing: this.pendingTrans > 0 })
        if (o.onDone) o.onDone()
      }
    })
  },

  // 把转写结果追加到指定题目的回答（即使已切到别的题也写回原题）
  mergeAnswer(index, text) {
    const trimmed = (text || '').trim()
    if (!trimmed) return
    // 先把输入框最新内容同步回题目，避免覆盖用户正在编辑的文字
    this.saveCurrentAnswer()
    const qs = this.data.questions
    const target = qs[index != null ? index : this.data.currentIndex]
    if (!target) return
    const merged = target.userAnswer ? target.userAnswer + '\n' + trimmed : trimmed
    target.userAnswer = merged
    if ((index != null ? index : this.data.currentIndex) === this.data.currentIndex) {
      this.setData({ questions: qs, answer: merged })
    } else {
      this.setData({ questions: qs })
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
      // 停止自动识别，保存当前答案
      this.stopAutoRecord()
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
      // 视频面试：新题继续自动识别
      this.startAutoRecord()
    }
  },

  // 下一题：先停自动识别并等转写落定，再提交当前答案给后端 AI 点评，最后前进
  nextQuestion() {
    if (this.data.submitting) return
    this.stopAutoRecord()
    this.saveCurrentAnswer()

    this.setData({ submitting: true })
    wx.showLoading({ title: '正在转写回答…', mask: true })
    this.waitTranscripts().then(() => {
      this.saveCurrentAnswer()
      const cur = this.data.questions[this.data.currentIndex] || {}
      const needSubmit = (cur.userAnswer || '').trim() !== '' && !cur.submitted
      if (!needSubmit) {
        wx.hideLoading()
        this.setData({ submitting: false })
        this.goNext()
        return
      }
      wx.showLoading({ title: 'AI 正在点评…', mask: true })
      return this.submitCurrentAnswer().then(() => {
        wx.hideLoading()
        this.setData({ submitting: false })
        this.goNext()
      })
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
      // 视频面试：新题自动开启识别循环
      this.startAutoRecord()
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
          this.stopAutoRecord()
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

  // 完成面试：停自动识别、等转写落定，再把最后一题答案提交（含 AI 点评），最后触发报告生成
  completeInterview() {
    this.stopAutoRecord()
    this.setData({ 
      submitting: true,
      showCompleteModal: true 
    })

    this.waitTranscripts().then(() => {
      this.saveCurrentAnswer()
      return this.submitCurrentAnswer()
    }).then(() => {
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
