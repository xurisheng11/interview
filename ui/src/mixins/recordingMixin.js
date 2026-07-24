import { getRecommendedMimeType } from '@/utils/mediaCompatibility'

/**
 * 视频录制 Mixin
 * 封装 MediaRecorder API，支持本地录制、下载
 */
export const recordingMixin = {
  data() {
    return {
      recorder: null,
      recordedChunks: [],
      isRecording: false,
      recordingStartTime: null,
      recordingDuration: 0,
      recordingTimer: null
    }
  },

  methods: {
    /**
     * 开始录制
     * @param {MediaStream} stream
     * @returns {boolean} 是否成功启动
     */
    startRecording(stream) {
      if (!window.MediaRecorder) {
        this.$message && this.$message.warning('当前浏览器不支持视频录制')
        return false
      }
      const mimeType = getRecommendedMimeType()
      try {
        const options = mimeType ? { mimeType } : {}
        this.recorder = new MediaRecorder(stream, options)
        this.recordedChunks = []

        this.recorder.ondataavailable = (event) => {
          if (event.data && event.data.size > 0) {
            this.recordedChunks.push(event.data)
          }
        }

        this.recorder.onstop = () => {
          this._generateVideoFile()
        }

        this.recorder.onerror = (event) => {
          const msg = event.error ? event.error.message : '录制发生错误'
          this.$emit && this.$emit('recording-error', msg)
          this.isRecording = false
          this._clearRecordingTimer()
        }

        this.recorder.start(30000) // 每 30 秒一个 chunk
        this.isRecording = true
        this.recordingStartTime = Date.now()
        this.recordingDuration = 0
        this._startRecordingTimer()
        return true
      } catch (err) {
        this.$emit && this.$emit('recording-error', err.message || '录制启动失败')
        return false
      }
    },

    /**
     * 停止录制
     */
    stopRecording() {
      if (this.recorder && this.isRecording) {
        try {
          this.recorder.stop()
        } catch (e) { /* ignore */ }
        this.isRecording = false
        this._clearRecordingTimer()
      }
    },

    /**
     * 合并 chunks 生成视频文件并提交到 Vuex
     */
    _generateVideoFile() {
      if (!this.recordedChunks.length) return
      const mimeType = (this.recorder && this.recorder.mimeType) || 'video/webm'
      const blob = new Blob(this.recordedChunks, { type: mimeType })
      const url = URL.createObjectURL(blob)
      const duration = this.recordingStartTime
        ? Math.round((Date.now() - this.recordingStartTime) / 1000)
        : 0
      const size = (blob.size / 1024 / 1024).toFixed(2) + ' MB'

      // 提交到 Vuex
      if (this.$store) {
        this.$store.commit('interview/SET_RECORDED_VIDEO', { url, blob, size, duration })
      }

      this.$emit && this.$emit('recording-ready', { url, size, duration })
      this.$message && this.$message.success(`视频录制完成，时长 ${duration}s，大小 ${size}`)
    },

    /**
     * 下载录制视频
     * @param {string} url - blob URL
     * @param {string} filename - 文件名
     */
    downloadVideo(url, filename) {
      const target = url || (this.$store && this.$store.state.interview.recordedVideo && this.$store.state.interview.recordedVideo.url)
      if (!target) {
        this.$message && this.$message.warning('暂无录制视频')
        return
      }
      const a = document.createElement('a')
      a.href = target
      a.download = filename || `interview-${Date.now()}.webm`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
    },

    _startRecordingTimer() {
      this._clearRecordingTimer()
      this.recordingTimer = setInterval(() => {
        this.recordingDuration++
      }, 1000)
    },

    _clearRecordingTimer() {
      if (this.recordingTimer) {
        clearInterval(this.recordingTimer)
        this.recordingTimer = null
      }
    },

    /**
     * 格式化录制时长
     * @param {number} seconds
     * @returns {string}
     */
    formatRecordingDuration(seconds) {
      const m = Math.floor(seconds / 60)
      const s = seconds % 60
      return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
    }
  },

  beforeDestroy() {
    this.stopRecording()
    this._clearRecordingTimer()
  }
}
