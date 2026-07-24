<template>
  <div class="dashboard">

    <!-- ① 欢迎横幅 -->
    <div class="welcome-banner">
      <div class="welcome-left">
        <div class="welcome-greeting">{{ greeting }}，<span class="welcome-name">{{ displayName }}</span> 👋</div>
        <div class="welcome-sub">{{ motivationalText }}</div>
        <div class="banner-actions">
          <el-button type="primary" size="small" class="banner-btn" @click="$router.push('/interview/config')">
            🚀 发起面试
          </el-button>
          <el-button size="small" class="banner-btn-ghost" @click="$router.push('/questions')">
            📚 题库练习
          </el-button>
          <el-button size="small" class="banner-btn-ghost" @click="$router.push('/company/intel')">
            🏢 公司面试知识库
          </el-button>
        </div>
      </div>
      <div class="welcome-right">
        <div class="banner-decoration">🎯</div>
      </div>
    </div>

    <!-- ② 统计卡片 -->
    <el-row :gutter="16" class="stats-row">
      <el-col :span="6" v-for="(card, i) in statCards" :key="i">
        <div class="stat-card" :class="'stat-card--' + card.theme" v-loading="statsLoading">
          <div class="stat-card-inner">
            <div class="stat-icon-big">{{ card.icon }}</div>
            <div class="stat-right">
              <div class="stat-num">{{ card.value }}</div>
              <div class="stat-label">{{ card.label }}</div>
            </div>
          </div>
          <div class="stat-bar">
            <div class="stat-bar-fill" :style="{ width: card.barWidth }"></div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- ③ 主内容 -->
    <el-row :gutter="16" class="main-row">

      <!-- 左列：近期记录 + 趋势图 -->
      <el-col :span="16">
        <!-- 近期面试记录 -->
        <el-card shadow="never" class="section-card" style="margin-bottom:16px">
          <div slot="header" class="card-header">
            <span class="card-title">📋 近期面试记录</span>
            <router-link to="/interview/history" class="view-all">查看全部 ›</router-link>
          </div>
          <el-table
            :data="interviews"
            v-loading="interviewsLoading"
            stripe
            style="width:100%"
            empty-text="暂无面试记录，去发起一次面试吧 🚀"
          >
            <el-table-column label="时间" width="145">
              <template slot-scope="{ row }">
                <span class="table-time">{{ formatTime(row.startTime || row.createdAt) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="岗位">
              <template slot-scope="{ row }">
                <span class="table-position">{{ row.jobTitle || row.position || '—' }}</span>
              </template>
            </el-table-column>
            <el-table-column label="轮次" width="70" align="center">
              <template slot-scope="{ row }">
                <el-tag size="mini" type="info">{{ roundLabel(row.round) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="得分" width="70" align="center">
              <template slot-scope="{ row }">
                <span :class="scoreClass(row.totalScore ?? row.score)">
                  {{ (row.totalScore ?? row.score) != null ? (row.totalScore ?? row.score) : '-' }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="80" align="center">
              <template slot-scope="{ row }">
                <el-tag size="mini" :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="90" align="center">
              <template slot-scope="{ row }">
                <el-button type="text" size="mini" @click="goReport(row.interviewId || row.id)" v-if="row.status === 'completed'">查看报告</el-button>
                <el-button type="text" size="mini" @click="continueInterview(row.interviewId || row.id)" v-else-if="row.status === 'ongoing' || row.status === 'paused'">继续</el-button>
                <span v-else class="text-muted">—</span>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <!-- 得分趋势 -->
        <el-card shadow="never" class="section-card">
          <div slot="header" class="card-header">
            <span class="card-title">📈 得分趋势</span>
            <span class="trend-tip">近期面试得分变化</span>
          </div>
          <div v-loading="trendLoading" class="chart-wrap">
            <div ref="trendChart" class="trend-chart"></div>
            <div v-if="!trendLoading && trendEmpty" class="empty-chart">
              <div class="empty-chart-icon">📊</div>
              <div>完成更多面试后，这里将展示你的进步曲线</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 右列 -->
      <el-col :span="8">
        <!-- 快捷导航 -->
        <el-card shadow="never" class="section-card" style="margin-bottom:16px">
          <div slot="header" class="card-title">⚡ 快捷导航</div>
          <div class="quick-nav-grid">
            <div
              v-for="nav in quickNavs"
              :key="nav.path"
              class="quick-nav-item"
              @click="$router.push(nav.path)"
            >
              <div class="quick-nav-icon" :style="{ background: nav.bg }">{{ nav.icon }}</div>
              <div class="quick-nav-label">{{ nav.label }}</div>
            </div>
          </div>
        </el-card>

        <!-- 今日推荐练习 -->
        <el-card shadow="never" class="section-card">
          <div slot="header" class="card-header">
            <span class="card-title">💡 今日推荐</span>
            <router-link to="/questions" class="view-all">更多 ›</router-link>
          </div>
          <div v-loading="questionsLoading">
            <div
              v-for="q in questions"
              :key="q.questionId || q.id"
              class="question-item"
              @click="goPractice(q.questionId || q.id)"
            >
              <div class="question-title">{{ q.content || q.title }}</div>
              <div class="question-meta">
                <el-tag size="mini" :type="difficultyTagType(q.difficulty)" class="meta-tag">
                  {{ difficultyLabel(q.difficulty) }}
                </el-tag>
                <el-tag size="mini" type="info" class="meta-tag" v-if="q.jobTitle || q.position">
                  {{ q.jobTitle || q.position }}
                </el-tag>
              </div>
            </div>
            <div v-if="!questionsLoading && questions.length === 0" class="empty-tip">暂无推荐题目</div>
          </div>
          <div class="start-btn-wrap">
            <el-button type="primary" size="small" style="width:100%" @click="$router.push('/interview/config')">
              🚀 发起新面试
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import request from '@/api/request'
import * as echarts from 'echarts'
import { mapGetters } from 'vuex'

export default {
  name: 'Dashboard',
  computed: {
    ...mapGetters('user', ['userInfo']),
    displayName() {
      if (!this.userInfo) return '同学'
      return this.userInfo.nickname || this.userInfo.username || '同学'
    },
    greeting() {
      const h = new Date().getHours()
      if (h < 6)  return '夜深了'
      if (h < 12) return '早上好'
      if (h < 14) return '中午好'
      if (h < 18) return '下午好'
      return '晚上好'
    },
    motivationalText() {
      const texts = [
        '每一次练习都是离梦想更近一步 💪',
        '今天的努力，是明天 offer 的基石 🌟',
        '坚持练习，面试无惧 🔥',
        '持续进步，下一个 offer 就是你的 🎉',
        '保持专注，好机会正在等你 🚀'
      ]
      return texts[new Date().getDay() % texts.length]
    },
    statCards() {
      const total = this.stats.totalInterviews || 0
      const avg   = this.stats.avgScore || 0
      const high  = this.stats.highestScore || 0
      return [
        { label: '累计面试', value: total, icon: '🎯', theme: 'orange', barWidth: Math.min(total * 10, 100) + '%' },
        { label: '平均得分', value: avg ? avg.toFixed(1) : '0.0', icon: '📊', theme: 'blue',   barWidth: avg + '%' },
        { label: '最高得分', value: high || 0, icon: '🏆', theme: 'green',  barWidth: high + '%' },
        { label: '最常岗位', value: this.stats.topPosition || '—', icon: '💼', theme: 'purple', barWidth: '60%' }
      ]
    },
    trendEmpty() {
      return !this.trendData || this.trendData.length === 0
    },
    quickNavs() {
      return [
        { icon: '🚀', label: '发起面试', path: '/interview/config',  bg: '#fff3e0' },
        { icon: '📁', label: '面试记录', path: '/interview/history', bg: '#e8f5e9' },
        { icon: '📚', label: '题库练习', path: '/questions',         bg: '#e3f2fd' },
        { icon: '🏢', label: '公司面试知识库', path: '/company/intel',   bg: '#fce4ec' },
        { icon: '🌐', label: '知识社区', path: '/community',         bg: '#f3e5f5' },
        { icon: '👤', label: '个人中心', path: '/profile',           bg: '#e0f7fa' }
      ]
    }
  },
  data() {
    return {
      statsLoading: true,
      stats: { totalInterviews: 0, avgScore: 0, highestScore: 0, topPosition: '—' },
      interviewsLoading: true,
      interviews: [],
      questionsLoading: true,
      questions: [],
      trendLoading: true,
      trendData: [],
      trendChart: null
    }
  },
  created() {
    this.loadAll()
  },
  mounted() {
    window.addEventListener('resize', this.handleResize)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.handleResize)
    if (this.trendChart) { this.trendChart.dispose(); this.trendChart = null }
  },
  methods: {
    async loadAll() {
      const [statsRes, interviewsRes, questionsRes, trendRes] = await Promise.allSettled([
        request.get('/profile/stats'),
        request.get('/interviews', { params: { page: 1, pageSize: 5 } }),
        request.get('/questions',  { params: { page: 1, pageSize: 4 } }),
        request.get('/profile/trend')
      ])
      if (statsRes.status === 'fulfilled') {
        const d = statsRes.value.data || statsRes.value
        this.stats = {
          totalInterviews: d.totalCount ?? d.totalInterviews ?? 0,
          avgScore:        d.avgScore ?? d.avg_score ?? 0,
          highestScore:    d.maxScore ?? d.highestScore ?? 0,
          topPosition:     d.topJobTitle || d.topPosition || '—'
        }
      }
      this.statsLoading = false
      if (interviewsRes.status === 'fulfilled') {
        const d = interviewsRes.value.data || interviewsRes.value
        this.interviews = Array.isArray(d) ? d : (d.list || d.items || [])
      }
      this.interviewsLoading = false
      if (questionsRes.status === 'fulfilled') {
        const d = questionsRes.value.data || questionsRes.value
        this.questions = Array.isArray(d) ? d : (d.list || d.items || [])
      }
      this.questionsLoading = false
      if (trendRes.status === 'fulfilled') {
        const d = trendRes.value.data || trendRes.value
        this.trendData = Array.isArray(d) ? d : (d.list || d.items || [])
      }
      this.trendLoading = false
      this.$nextTick(() => { this.initChart() })
    },
    initChart() {
      if (!this.$refs.trendChart || this.trendEmpty) return
      if (this.trendChart) this.trendChart.dispose()
      this.trendChart = echarts.init(this.$refs.trendChart)
      const dates  = this.trendData.map(i => i.date || i.day || '')
      const scores = this.trendData.map(i => i.score ?? i.avgScore ?? 0)
      this.trendChart.setOption({
        tooltip: { trigger: 'axis', formatter: p => `${p[0].axisValue}<br/>得分：<b>${p[0].value}</b>` },
        grid: { left: 40, right: 20, top: 16, bottom: 40 },
        xAxis: {
          type: 'category', data: dates,
          axisLabel: { fontSize: 12, color: '#888' },
          axisLine: { lineStyle: { color: '#eee' } },
          axisTick: { show: false }
        },
        yAxis: {
          type: 'value', min: 0, max: 100, interval: 25,
          axisLabel: { fontSize: 12, color: '#888' },
          splitLine: { lineStyle: { color: '#f5f5f5', type: 'dashed' } }
        },
        series: [{
          type: 'line', data: scores, smooth: true,
          symbol: 'circle', symbolSize: 7,
          lineStyle: { color: '#ff9900', width: 2.5 },
          itemStyle: { color: '#ff9900', borderColor: '#fff', borderWidth: 2 },
          areaStyle: {
            color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [{ offset: 0, color: 'rgba(255,153,0,0.25)' }, { offset: 1, color: 'rgba(255,153,0,0)' }] }
          }
        }]
      })
    },
    handleResize() { if (this.trendChart) this.trendChart.resize() },
    formatTime(val) {
      if (!val) return '—'
      const d = new Date(val)
      if (isNaN(d.getTime())) return val
      const p = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${p(d.getMonth()+1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
    },
    scoreClass(s) {
      if (s == null) return 'score-na'
      if (s >= 80) return 'score-high'
      if (s >= 60) return 'score-mid'
      return 'score-low'
    },
    statusTagType(s) { return { completed: 'success', in_progress: 'warning', pending: 'info', failed: 'danger' }[s] || 'info' },
    statusLabel(s)   { return { completed: '已完成', in_progress: '进行中', pending: '待开始', failed: '已中断', ongoing: '进行中', paused: '已暂停' }[s] || s || '未知' },
    difficultyTagType(d) { return { easy: 'success', medium: 'warning', hard: 'danger' }[d] || 'info' },
    difficultyLabel(d)   { return { easy: '简单', medium: '中等', hard: '困难' }[d] || d || '未知' },
    goReport(id)         { this.$router.push(`/report/${id}`) },
    continueInterview(id){ this.$router.push(`/interview/${id}/doing`) },
    goPractice(id)       { this.$router.push(`/questions/${id}/practice`) },
    roundLabel(r)        { return { round1: '一面', round2: '二面', round3: '三面' }[r] || r || '—' }
  }
}
</script>

<style scoped>
.dashboard { padding: 20px; max-width: 1200px; margin: 0 auto; }

/* ── 欢迎横幅 ── */
.welcome-banner {
  background: linear-gradient(135deg, #131921 0%, #232f3e 60%, #3a4a5c 100%);
  border-radius: 12px;
  padding: 28px 32px;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  overflow: hidden;
  position: relative;
}
.welcome-greeting {
  font-size: 24px;
  font-weight: bold;
  color: #fff;
  margin-bottom: 6px;
}
.welcome-name { color: #ff9900; }
.welcome-sub {
  font-size: 14px;
  color: #aab7c4;
  margin-bottom: 18px;
}
.banner-actions { display: flex; gap: 10px; flex-wrap: wrap; }
.banner-btn {
  background: #ff9900 !important;
  border-color: #ff9900 !important;
  color: #111 !important;
  font-weight: bold;
}
.banner-btn-ghost {
  background: transparent !important;
  border-color: rgba(255,255,255,0.3) !important;
  color: #fff !important;
}
.banner-btn-ghost:hover {
  border-color: #ff9900 !important;
  color: #ff9900 !important;
}
.welcome-right { flex-shrink: 0; }
.banner-decoration {
  font-size: 72px;
  opacity: 0.18;
  line-height: 1;
  user-select: none;
}

/* ── 统计卡片 ── */
.stats-row { margin-bottom: 20px; }
.stat-card {
  border-radius: 10px;
  padding: 18px 20px 10px;
  background: #fff;
  border: 1px solid #eee;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
  transition: transform 0.2s, box-shadow 0.2s;
  overflow: hidden;
}
.stat-card:hover { transform: translateY(-3px); box-shadow: 0 6px 20px rgba(0,0,0,0.09); }
.stat-card-inner { display: flex; align-items: center; gap: 14px; margin-bottom: 14px; }
.stat-icon-big { font-size: 32px; line-height: 1; }
.stat-num { font-size: 26px; font-weight: bold; line-height: 1.2; color: #111; }
.stat-label { font-size: 12px; color: #888; margin-top: 3px; }
.stat-bar { height: 4px; background: #f0f0f0; border-radius: 2px; overflow: hidden; }
.stat-bar-fill { height: 100%; border-radius: 2px; transition: width 0.8s ease; }
.stat-card--orange .stat-num { color: #e07b00; }
.stat-card--orange .stat-bar-fill { background: linear-gradient(90deg, #ff9900, #ffb84d); }
.stat-card--blue   .stat-num { color: #0066c0; }
.stat-card--blue   .stat-bar-fill { background: linear-gradient(90deg, #0066c0, #4da6ff); }
.stat-card--green  .stat-num { color: #067d62; }
.stat-card--green  .stat-bar-fill { background: linear-gradient(90deg, #067d62, #3dbf9a); }
.stat-card--purple .stat-num { color: #6e40c9; }
.stat-card--purple .stat-bar-fill { background: linear-gradient(90deg, #6e40c9, #a07de0); }

/* ── 卡片公共 ── */
.section-card { height: 100%; }
.card-header { display: flex; align-items: center; justify-content: space-between; }
.card-title {
  font-size: 15px; font-weight: bold; color: #111;
  border-left: 4px solid #ff9900; padding-left: 10px;
}
.view-all { font-size: 13px; color: #0066c0; text-decoration: none; cursor: pointer; }
.view-all:hover { color: #ff9900; }
.trend-tip { font-size: 12px; color: #bbb; }

/* ── 表格 ── */
.table-time   { font-size: 12px; color: #888; }
.table-position { font-weight: 500; color: #333; }
.score-high { color: #067d62; font-weight: bold; }
.score-mid  { color: #ff9900; font-weight: bold; }
.score-low  { color: #c7511f; font-weight: bold; }
.score-na   { color: #ccc; }
.text-muted { color: #ccc; font-size: 12px; }

/* ── 趋势图 ── */
.chart-wrap { position: relative; min-height: 220px; }
.trend-chart { width: 100%; height: 220px; }
.empty-chart {
  position: absolute; top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  text-align: center; color: #bbb; font-size: 13px;
}
.empty-chart-icon { font-size: 36px; margin-bottom: 8px; }

/* ── 快捷导航 ── */
.quick-nav-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}
.quick-nav-item {
  display: flex; flex-direction: column; align-items: center;
  padding: 12px 6px; border-radius: 10px; cursor: pointer;
  border: 1px solid #f0f0f0; transition: all 0.18s;
}
.quick-nav-item:hover {
  border-color: #ff9900;
  box-shadow: 0 2px 10px rgba(255,153,0,0.15);
  transform: translateY(-2px);
}
.quick-nav-icon {
  width: 40px; height: 40px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  font-size: 20px; margin-bottom: 6px;
}
.quick-nav-label { font-size: 12px; color: #555; text-align: center; }

/* ── 推荐题目 ── */
.question-item {
  padding: 10px 0; border-bottom: 1px solid #f5f5f5; cursor: pointer; transition: 0.15s;
}
.question-item:last-child { border-bottom: none; }
.question-item:hover .question-title { color: #ff9900; }
.question-title {
  font-size: 13px; color: #0066c0; margin-bottom: 5px; line-height: 1.5;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
.question-meta { display: flex; flex-wrap: wrap; gap: 4px; }
.meta-tag { margin: 0; }
.empty-tip { font-size: 13px; color: #bbb; text-align: center; padding: 16px 0; }
.start-btn-wrap { margin-top: 14px; padding-top: 12px; border-top: 1px solid #f5f5f5; }

/* ── 覆盖 Element UI 橙色 ── */
::v-deep .el-button--primary {
  background: #ff9900; border-color: #ff9900; color: #111;
}
::v-deep .el-button--primary:hover { background: #f3a847; border-color: #f3a847; }
::v-deep .el-button--primary.is-disabled { background: #ddd; border-color: #ddd; color: #999; }
</style>
