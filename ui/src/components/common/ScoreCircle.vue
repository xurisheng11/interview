<template>
  <svg :width="size" :height="size" :viewBox="`0 0 ${size} ${size}`">
    <!-- 背景圆 -->
    <circle
      :cx="size / 2"
      :cy="size / 2"
      :r="radius"
      fill="none"
      stroke="#eee"
      :stroke-width="strokeWidth"
    />
    <!-- 进度圆（旋转 -90 度从顶部开始） -->
    <circle
      :cx="size / 2"
      :cy="size / 2"
      :r="radius"
      fill="none"
      :stroke="color"
      :stroke-width="strokeWidth"
      :stroke-dasharray="`${progress} ${circumference}`"
      stroke-linecap="round"
      :transform="`rotate(-90 ${size / 2} ${size / 2})`"
    />
    <!-- 中间文字 -->
    <text :x="size / 2" :y="size / 2" text-anchor="middle" dominant-baseline="central">
      <tspan :font-size="size * 0.28" font-weight="bold" :fill="color">{{ score }}</tspan>
      <tspan :font-size="size * 0.14" fill="#999" dy="4">分</tspan>
    </text>
  </svg>
</template>

<script>
export default {
  name: 'ScoreCircle',
  props: {
    score: {
      type: Number,
      required: true
    },
    size: {
      type: Number,
      default: 120
    }
  },
  computed: {
    strokeWidth() {
      return this.size * 0.1
    },
    radius() {
      return (this.size - this.strokeWidth) / 2
    },
    circumference() {
      return 2 * Math.PI * this.radius
    },
    progress() {
      return (this.score / 100) * this.circumference
    },
    color() {
      if (this.score >= 80) return '#067d62'
      if (this.score >= 60) return '#ff9900'
      return '#c7511f'
    }
  }
}
</script>
