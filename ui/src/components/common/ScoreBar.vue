<template>
  <div class="score-bar-wrap">
    <div class="score-bar-track">
      <div
        class="score-bar-fill"
        :style="{ width: percent + '%', background: color }"
      ></div>
    </div>
    <span class="score-bar-num" :style="{ color }">{{ score }}</span>
  </div>
</template>

<script>
export default {
  name: 'ScoreBar',
  props: {
    score: { type: Number, default: 0 }
  },
  computed: {
    percent() {
      return Math.min(100, Math.max(0, this.score || 0))
    },
    color() {
      if (this.score >= 80) return '#67c23a'
      if (this.score >= 60) return '#ff9900'
      return '#f56c6c'
    }
  }
}
</script>

<style scoped>
.score-bar-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.score-bar-track {
  flex: 1;
  height: 10px;
  background: #eee;
  border-radius: 5px;
  overflow: hidden;
}

.score-bar-fill {
  height: 100%;
  border-radius: 5px;
  transition: width 0.4s ease;
}

.score-bar-num {
  font-size: 14px;
  font-weight: bold;
  min-width: 32px;
  text-align: right;
}
</style>
