<template>
  <div ref="chart" class="radar-chart"></div>
</template>

<script>
import * as echarts from 'echarts'

export default {
  name: 'RadarChart',
  props: {
    // modules: [{ name: '算法', score: 80 }, ...]
    modules: {
      type: Array,
      default: () => []
    }
  },

  data() {
    return { chart: null }
  },

  watch: {
    modules() { this.renderChart() }
  },

  mounted() {
    this.chart = echarts.init(this.$refs.chart)
    this.renderChart()
    window.addEventListener('resize', this.handleResize)
  },

  beforeDestroy() {
    window.removeEventListener('resize', this.handleResize)
    if (this.chart) { this.chart.dispose() }
  },

  methods: {
    handleResize() {
      this.chart && this.chart.resize()
    },

    renderChart() {
      if (!this.chart || !this.modules.length) return

      const indicators = this.modules.map(m => ({ name: m.name, max: 100 }))
      const values = this.modules.map(m => m.score || 0)

      this.chart.setOption({
        tooltip: { trigger: 'item' },
        radar: {
          indicator: indicators,
          center: ['50%', '50%'],
          radius: '65%',
          name: { textStyle: { color: '#555', fontSize: 12 } },
          splitArea: { areaStyle: { color: ['rgba(255,153,0,0.02)', 'rgba(255,153,0,0.05)'] } },
          splitLine: { lineStyle: { color: 'rgba(255,153,0,0.2)' } },
          axisLine: { lineStyle: { color: 'rgba(255,153,0,0.3)' } }
        },
        series: [{
          type: 'radar',
          data: [{
            value: values,
            name: '得分',
            areaStyle: { color: 'rgba(255,153,0,0.15)' },
            lineStyle: { color: '#ff9900', width: 2 },
            itemStyle: { color: '#ff9900' }
          }]
        }]
      })
    }
  }
}
</script>

<style scoped>
.radar-chart {
  width: 100%;
  height: 280px;
}
</style>
