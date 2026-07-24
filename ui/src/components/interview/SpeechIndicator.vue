<template>
  <div class="speech-indicator">
    <canvas ref="waveCanvas" width="120" height="40" class="wave-canvas"></canvas>
    <span class="status-text" :class="statusClass">{{ statusText }}</span>
  </div>
</template>

<script>
export default {
  name: 'SpeechIndicator',
  props: {
    stream: {
      default: null
    },
    active: {
      type: Boolean,
      default: false
    },
    supported: {
      type: Boolean,
      default: true
    }
  },
  data() {
    return {
      audioContext: null,
      analyser: null,
      animationId: null,
      dataArray: null
    }
  },
  computed: {
    statusText() {
      if (!this.supported) return '不支持'
      return this.active ? '正在识别...' : '已暂停'
    },
    statusClass() {
      if (!this.supported) return 'unsupported'
      return this.active ? 'active' : 'paused'
    }
  },
  watch: {
    stream(newStream) {
      this._teardown()
      if (newStream) this._setup(newStream)
    },
    active(val) {
      if (val) {
        this._startDraw()
      } else {
        this._stopDraw()
        this._clearCanvas()
      }
    }
  },
  mounted() {
    if (this.stream) this._setup(this.stream)
  },
  beforeDestroy() {
    this._teardown()
  },
  methods: {
    _setup(stream) {
      try {
        const AudioContext = window.AudioContext || window.webkitAudioContext
        if (!AudioContext) return
        this.audioContext = new AudioContext()
        const source = this.audioContext.createMediaStreamSource(stream)
        this.analyser = this.audioContext.createAnalyser()
        this.analyser.fftSize = 64
        source.connect(this.analyser)
        this.dataArray = new Uint8Array(this.analyser.frequencyBinCount)
        if (this.active) this._startDraw()
      } catch (e) {
        // AudioContext 创建失败，静默处理
      }
    },
    _teardown() {
      this._stopDraw()
      if (this.audioContext) {
        try { this.audioContext.close() } catch (e) { /* ignore */ }
        this.audioContext = null
        this.analyser = null
        this.dataArray = null
      }
    },
    _startDraw() {
      if (!this.analyser) return
      this._stopDraw()
      const draw = () => {
        this.animationId = requestAnimationFrame(draw)
        this._drawWave()
      }
      draw()
    },
    _stopDraw() {
      if (this.animationId) {
        cancelAnimationFrame(this.animationId)
        this.animationId = null
      }
    },
    _clearCanvas() {
      const canvas = this.$refs.waveCanvas
      if (!canvas) return
      const ctx = canvas.getContext('2d')
      ctx.clearRect(0, 0, canvas.width, canvas.height)
    },
    _drawWave() {
      const canvas = this.$refs.waveCanvas
      if (!canvas || !this.analyser) return
      const ctx = canvas.getContext('2d')
      const W = canvas.width
      const H = canvas.height
      this.analyser.getByteFrequencyData(this.dataArray)

      ctx.clearRect(0, 0, W, H)
      ctx.fillStyle = '#f0f0f0'
      ctx.fillRect(0, 0, W, H)

      const barWidth = W / this.dataArray.length
      ctx.fillStyle = '#ff9900'
      for (let i = 0; i < this.dataArray.length; i++) {
        const barHeight = (this.dataArray[i] / 255) * H
        ctx.fillRect(i * barWidth, H - barHeight, barWidth - 1, barHeight)
      }
    }
  }
}
</script>

<style scoped>
.speech-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wave-canvas {
  border-radius: 4px;
  border: 1px solid #e0e0e0;
}

.status-text {
  font-size: 12px;
  white-space: nowrap;
}

.status-text.active   { color: #67c23a; }
.status-text.paused   { color: #999; }
.status-text.unsupported { color: #f56c6c; }
</style>
