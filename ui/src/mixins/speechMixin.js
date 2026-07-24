/**
 * 语音识别 Mixin
 * 封装 Web Speech API，支持实时转文字、语速/停顿计算
 */
export const speechMixin = {
  data() {
    return {
      recognition: null,
      isSpeechActive: false,
      isSpeechSupported: false,
      finalTranscript: '',
      interimTranscript: '',
      speechMetrics: {
        startTime: null,
        totalWords: 0,
        pauseThreshold: 2000, // 2秒静默视为停顿
        pauseCount: 0,
        lastResultTime: null
      }
    }
  },

  methods: {
    /**
     * 检测语音识别是否可用
     * @returns {boolean}
     */
    checkSpeechSupport() {
      const supported = !!(window.SpeechRecognition || window.webkitSpeechRecognition)
      this.isSpeechSupported = supported
      return supported
    },

    /**
     * 初始化语音识别
     * @param {string} lang - 语言代码，默认 'zh-CN'
     * @returns {boolean} 是否初始化成功
     */
    initSpeechRecognition(lang = 'zh-CN') {
      const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition
      if (!SpeechRecognition) {
        this.isSpeechSupported = false
        return false
      }
      this.isSpeechSupported = true

      // 如果已存在实例先清理
      if (this.recognition) {
        try { this.recognition.abort() } catch (e) { /* ignore */ }
        this.recognition = null
      }

      this.recognition = new SpeechRecognition()
      this.recognition.lang = lang
      this.recognition.continuous = true
      this.recognition.interimResults = true
      this.recognition.maxAlternatives = 1

      this.recognition.onresult = (event) => {
        const now = Date.now()
        // 检测停顿
        if (this.speechMetrics.lastResultTime) {
          const gap = now - this.speechMetrics.lastResultTime
          if (gap > this.speechMetrics.pauseThreshold) {
            this.speechMetrics.pauseCount++
          }
        }
        this.speechMetrics.lastResultTime = now

        let interim = ''
        let final = ''
        for (let i = event.resultIndex; i < event.results.length; i++) {
          const text = event.results[i][0].transcript
          if (event.results[i].isFinal) {
            final += text
            this.speechMetrics.totalWords += text.length
          } else {
            interim += text
          }
        }
        if (final) this.finalTranscript += final
        this.interimTranscript = interim
        // 同步到父组件答案框
        if (typeof this.userAnswer !== 'undefined') {
          this.userAnswer = this.finalTranscript
        }
      }

      this.recognition.onerror = (event) => {
        if (event.error !== 'no-speech' && event.error !== 'aborted') {
          this.$message && this.$message.warning(`语音识别错误：${event.error}`)
        }
        if (event.error === 'not-allowed') {
          this.isSpeechActive = false
        }
      }

      this.recognition.onend = () => {
        // continuous 模式下意外结束时自动重启（除非主动停止）
        if (this.isSpeechActive) {
          try { this.recognition.start() } catch (e) { /* ignore */ }
        }
      }

      return true
    },

    /**
     * 开始语音识别（新题开始时调用，重置计数器）
     */
    startSpeech() {
      if (!this.recognition) return
      this.finalTranscript = ''
      this.interimTranscript = ''
      this.speechMetrics.startTime = Date.now()
      this.speechMetrics.totalWords = 0
      this.speechMetrics.pauseCount = 0
      this.speechMetrics.lastResultTime = null
      this.isSpeechActive = true
      try {
        this.recognition.start()
      } catch (e) {
        // 已经在运行中，忽略
      }
    },

    /**
     * 停止语音识别
     */
    stopSpeech() {
      this.isSpeechActive = false
      this.interimTranscript = ''
      if (this.recognition) {
        try { this.recognition.stop() } catch (e) { /* ignore */ }
      }
    },

    /**
     * 切换语音识别开关
     */
    toggleSpeech() {
      if (this.isSpeechActive) {
        this.stopSpeech()
      } else {
        if (!this.recognition) {
          this.initSpeechRecognition(this.speechLang || 'zh-CN')
        }
        this.isSpeechActive = true
        try { this.recognition.start() } catch (e) { /* ignore */ }
      }
    },

    /**
     * 计算语速（每分钟字数）
     * @returns {number}
     */
    calcSpeechRate() {
      if (!this.speechMetrics.startTime) return 0
      const durationMin = (Date.now() - this.speechMetrics.startTime) / 60000
      return durationMin > 0
        ? Math.round(this.speechMetrics.totalWords / durationMin)
        : 0
    },

    /**
     * 获取非语言指标
     * @returns {{ speechRate: number, pauseCount: number, duration: number }}
     */
    getNonVerbalMetrics() {
      const duration = this.speechMetrics.startTime
        ? Math.round((Date.now() - this.speechMetrics.startTime) / 1000)
        : 0
      return {
        speechRate: this.calcSpeechRate(),
        pauseCount: this.speechMetrics.pauseCount,
        duration
      }
    }
  },

  beforeDestroy() {
    this.stopSpeech()
    if (this.recognition) {
      try { this.recognition.abort() } catch (e) { /* ignore */ }
      this.recognition = null
    }
  }
}
