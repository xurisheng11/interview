/**
 * 语音识别 Mixin
 * 封装 Web Speech API，支持实时转文字、语速/停顿计算、口头禅实时检测
 */

// 常见中文口头禅列表（用户说3次以上时触发提示）
const VERBAL_TICS = [
  '然后', '这个', '那个', '嗯', '呃', '啊', '就是', '就是说',
  '的话', '其实', '基本上', '大概', '可能', '应该', '好像',
  '对吧', '是吧', '好吗', '好吧', '那个', '所以', '然后呢',
  '那个什么', '就', '就这样', '反正'
]

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
        pauseThreshold: 2000,
        pauseCount: 0,
        lastResultTime: null,
        firstSpeechTime: null  // 首次检测到语音的时间戳
      },
      verbalTicCount: {},
      ticAlertShown: {}
    }
  },

  methods: {
    checkSpeechSupport() {
      const supported = !!(window.SpeechRecognition || window.webkitSpeechRecognition)
      this.isSpeechSupported = supported
      return supported
    },

    initSpeechRecognition(lang = 'zh-CN') {
      const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition
      if (!SpeechRecognition) {
        this.isSpeechSupported = false
        return false
      }
      this.isSpeechSupported = true

      if (this.recognition) {
        try { this.recognition.abort() } catch (e) { }
        this.recognition = null
      }

      this.recognition = new SpeechRecognition()
      this.recognition.lang = lang
      this.recognition.continuous = true
      this.recognition.interimResults = true
      this.recognition.maxAlternatives = 1

      this.recognition.onresult = (event) => {
        const now = Date.now()
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
            this._scanVerbalTics(text)
          } else {
            interim += text
          }
        }

        // 首次检测到非空语音结果时记录时间戳（仅首次，不覆盖）
        if (!this.speechMetrics.firstSpeechTime && (final || interim)) {
          this.speechMetrics.firstSpeechTime = now
        }

        if (final) this.finalTranscript += final
        this.interimTranscript = interim
        if (typeof this.userAnswer !== 'undefined') {
          this.userAnswer = this.finalTranscript
        }
      }

      this.recognition.onerror = (event) => {
        if (event.error !== 'no-speech' && event.error !== 'aborted') {
          this.$message && this.$message.warning('语音识别错误：' + event.error)
        }
        if (event.error === 'not-allowed') {
          this.isSpeechActive = false
        }
      }

      this.recognition.onend = () => {
        if (this.isSpeechActive) {
          try { this.recognition.start() } catch (e) { }
        }
      }

      return true
    },

    _scanVerbalTics(text) {
      VERBAL_TICS.forEach(tic => {
        const regex = new RegExp(tic, 'g')
        const matches = text.match(regex)
        if (matches) {
          const count = matches.length
          this.verbalTicCount[tic] = (this.verbalTicCount[tic] || 0) + count
          if (this.verbalTicCount[tic] >= 3 && !this.ticAlertShown[tic]) {
            this.ticAlertShown[tic] = true
            if (this.$EventBus && this.$EventBus.$emit) {
              this.$EventBus.$emit('verbal-tic-detected', {
                tic,
                count: this.verbalTicCount[tic],
                totalTics: { ...this.verbalTicCount }
              })
            }
            if (typeof this.onVerbalTicDetected === 'function') {
              this.onVerbalTicDetected(tic, this.verbalTicCount[tic], { ...this.verbalTicCount })
            }
          }
        }
      })
    },

    startSpeech() {
      if (!this.recognition) return
      this.finalTranscript = ''
      this.interimTranscript = ''
      this.speechMetrics.startTime = Date.now()
      this.speechMetrics.totalWords = 0
      this.speechMetrics.pauseCount = 0
      this.speechMetrics.lastResultTime = null
      this.speechMetrics.firstSpeechTime = null  // 重置首次语音时间
      this.verbalTicCount = {}
      this.ticAlertShown = {}
      this.isSpeechActive = true
      try {
        this.recognition.start()
      } catch (e) { }
    },

    stopSpeech() {
      this.isSpeechActive = false
      this.interimTranscript = ''
      if (this.recognition) {
        try { this.recognition.stop() } catch (e) { }
      }
    },

    toggleSpeech() {
      if (this.isSpeechActive) {
        this.stopSpeech()
      } else {
        if (!this.recognition) {
          this.initSpeechRecognition(this.speechLang || 'zh-CN')
        }
        this.isSpeechActive = true
        try { this.recognition.start() } catch (e) { }
      }
    },

    calcSpeechRate() {
      if (!this.speechMetrics.startTime) return 0
      const durationMin = (Date.now() - this.speechMetrics.startTime) / 60000
      return durationMin > 0
        ? Math.round(this.speechMetrics.totalWords / durationMin)
        : 0
    },

    getNonVerbalMetrics() {
      const duration = this.speechMetrics.startTime
        ? Math.round((Date.now() - this.speechMetrics.startTime) / 1000)
        : 0
      const verbalTics = Object.keys(this.verbalTicCount).filter(
        k => this.verbalTicCount[k] > 0
      )
      // 计算思考时长：从题目展示到首次开口（通过 mixin 访问组件的 questionDisplayedAt）
      // H3(2) 修复：加 Math.max(0, ...) 防御，防止出现负数
      const thinkDuration = (this.speechMetrics.firstSpeechTime && this.questionDisplayedAt)
        ? Math.max(0, Math.round((this.speechMetrics.firstSpeechTime - this.questionDisplayedAt) / 1000))
        : 0
      return {
        speechRate: this.calcSpeechRate(),
        pauseCount: this.speechMetrics.pauseCount,
        duration,
        thinkDuration,  // 新增：思考时长（秒）
        verbalTics
      }
    }
  },

  beforeDestroy() {
    this.stopSpeech()
    if (this.recognition) {
      try { this.recognition.abort() } catch (e) { }
      this.recognition = null
    }
  }
}
