<template>
  <div class="report-share" v-loading="loading">
    <div v-if="report" class="report-body">
      <!-- 顶部状态横幅 + 分享标识 -->
      <div :class="['status-banner', statusClass]">
        <span class="status-icon">{{ statusIcon }}</span>
        <span class="status-text">{{ statusText }}</span>
        <span class="share-badge">🔗 这是一份分享的面试报告</span>
      </div>

      <!-- 综合评分区 -->
      <el-card shadow="hover" class="section-card">
        <div slot="header" class="card-header">
          <span class="card-title">📊 综合评分</span>
        </div>
        <div class="score-overview">
          <div class="score-circle-wrap">
            <score-circle :score="report.totalScore || 0" :size="160" />
            <div class="grade-text">{{ report.grade }}</div>
          </div>
          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">目标岗位</span>
              <span class="info-value">{{ report.jobTitle || '—' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">面试轮次</span>
              <span class="info-value">{{ roundLabel(report.round) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">面试难度</span>
              <span class="info-value">{{ difficultyLabel(report.difficulty) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">面试日期</span>
              <span class="info-value">{{ formatDate(report.createdAt) }}</span>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 雷达图 + 模块得分（moduleScores 为空时整体隐藏） -->
      <el-row :gutter="16" class="section-row" v-if="hasModuleScores">
        <el-col :xs="24" :sm="12">
          <el-card shadow="hover" class="section-card">
            <div slot="header" class="card-header">
              <span class="card-title">🕸️ 知识点雷达图</span>
            </div>
            <radar-chart :modules="radarModules" />
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-card shadow="hover" class="section-card">
            <div slot="header" class="card-header">
              <span class="card-title">🧩 知识点模块得分</span>
            </div>
            <div class="module-grid">
              <div v-for="(m, i) in report.moduleScores" :key="i" class="module-card">
                <div class="module-name">{{ m.module }}</div>
                <div class="module-score" :style="{ color: scoreColor(m.avgScore) }">
                  {{ roundScore(m.avgScore) }}
                </div>
                <div class="module-meta">共 {{ m.count }} 题 · {{ m.level }}</div>
                <score-bar :score="roundScore(m.avgScore)" />
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- AI 综合评价（aiSummary 为 null 时隐藏） -->
      <el-card shadow="hover" class="section-card" v-if="report.aiSummary">
        <div slot="header" class="card-header">
          <span class="card-title">🤖 AI 综合评价与建议</span>
        </div>
        <el-row :gutter="16">
          <el-col :xs="24" :sm="12">
            <div class="ai-block">
              <div class="ai-block-title">💪 优势亮点</div>
              <ul class="ai-list strengths-list">
                <li v-for="(s, i) in (report.aiSummary.strengths || [])" :key="i">
                  <span class="list-icon">✅</span> {{ s }}
                </li>
              </ul>
            </div>
          </el-col>
          <el-col :xs="24" :sm="12">
            <div class="ai-block">
              <div class="ai-block-title">📌 待提升方向</div>
              <ul class="ai-list weakness-list">
                <li v-for="(w, i) in (report.aiSummary.weaknesses || [])" :key="i">
                  <span class="list-icon">⚠️</span>
                  <span v-if="w && typeof w === 'object'">
                    {{ w.point }}
                    <span v-if="w.suggestion" class="suggestion"> — {{ w.suggestion }}</span>
                  </span>
                  <span v-else>{{ w }}</span>
                </li>
              </ul>
            </div>
          </el-col>
        </el-row>
        <div class="roadmap-block" v-if="report.aiSummary.roadmap">
          <div class="ai-block-title">🗺️ 备考路线图</div>
          <div class="roadmap-text">{{ report.aiSummary.roadmap }}</div>
        </div>
      </el-card>

      <!-- 产品引导区 -->
      <div class="cta-block">
        <div class="cta-title">想测试你的面试水平？</div>
        <div class="cta-sub">AI 出题 · 智能点评 · 专属备考路线图，立即开启你的模拟面试</div>
        <el-button type="warning" size="medium" @click="goLogin">🚀 免费体验 AI 模拟面试</el-button>
      </div>
    </div><!-- end report-body -->

    <!-- 链接无效 / 已过期（404） -->
    <el-empty
      v-if="!loading && !report && errorType === 'expired'"
      description="分享链接无效或已过期（链接有效期为 7 天）"
      class="state-wrap"
    >
      <el-button type="primary" @click="goLogin">去登录体验</el-button>
    </el-empty>

    <!-- 网络失败 / 服务异常（可重试） -->
    <div v-if="!loading && !report && errorType === 'network'" class="error-state state-wrap">
      <div class="error-icon">📡</div>
      <div class="error-text">报告加载失败，请重试</div>
      <el-button type="primary" @click="loadReport">重新加载</el-button>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import ScoreCircle from '@/components/common/ScoreCircle.vue'
import RadarChart from '@/components/common/RadarChart.vue'
import ScoreBar from '@/components/common/ScoreBar.vue'

// 说明：不复用 @/api/request 封装 —— 其响应拦截器会对 404/5xx 统一弹出
// Message 错误框（分享页要求 404 静默降级为空态），且会自动附加本地 JWT。
// 分享页为公开页面，改用独立 axios 实例请求公开接口，行为完全自包含，
// 不影响全局拦截器与其他页面。
const publicApi = axios.create({
  baseURL: process.env.VUE_APP_API_BASE_URL || '/api/v1',
  timeout: 30000
})

export default {
  name: 'ReportShare',
  components: { ScoreCircle, RadarChart, ScoreBar },

  data() {
    return {
      loading: true,
      report: null,
      errorType: '' // '' | 'expired'（404）| 'network'（网络失败/5xx）
    }
  },

  computed: {
    // 横幅状态：优先使用后端 passStatus，缺失时按分数兜底（阈值同 Detail.vue）
    passLevel() {
      if (!this.report) return 'fail'
      const st = this.report.passStatus
      if (st === 'pass' || st === 'pending' || st === 'fail') return st
      const s = this.report.totalScore || 0
      if (s >= 75) return 'pass'
      if (s >= 60) return 'pending'
      return 'fail'
    },
    statusClass() {
      return { pass: 'banner-pass', pending: 'banner-pending', fail: 'banner-fail' }[this.passLevel]
    },
    statusIcon() {
      return { pass: '✅', pending: '⚠️', fail: '❌' }[this.passLevel]
    },
    statusText() {
      return {
        pass: '恭喜，模拟面试通过！',
        pending: '发挥一般，建议加强练习',
        fail: '本次未通过，继续努力！'
      }[this.passLevel]
    },
    hasModuleScores() {
      return !!(this.report && this.report.moduleScores && this.report.moduleScores.length)
    },
    radarModules() {
      if (!this.hasModuleScores) return []
      return this.report.moduleScores.map(m => ({ name: m.module, score: this.roundScore(m.avgScore) }))
    }
  },

  mounted() {
    this.loadReport()
  },

  methods: {
    async loadReport() {
      const token = this.$route.params.token
      this.loading = true
      this.errorType = ''
      try {
        const res = await publicApi.get(`/reports/share/${token}`)
        const body = res.data || {}
        this.report = body.data || null
        if (!this.report) {
          this.errorType = 'network'
        }
      } catch (e) {
        // 404 = token 无效或已过期 → 空态；其余（网络失败/5xx）→ 可重试错误态
        if (e.response && e.response.status === 404) {
          this.errorType = 'expired'
        } else {
          this.errorType = 'network'
        }
        this.report = null
      } finally {
        this.loading = false
      }
    },

    goLogin() {
      this.$router.push('/login')
    },

    roundLabel(round) {
      const map = { round1: '一面', round2: '二面', round3: '三面', comprehensive: '综合面试' }
      return map[round] || round || '—'
    },

    difficultyLabel(diff) {
      const map = { junior: '初级', middle: '中级', senior: '高级' }
      return map[diff] || diff || '—'
    },

    // createdAt（RFC3339）→ YYYY-MM-DD
    formatDate(str) {
      if (!str) return '—'
      const d = new Date(str)
      if (isNaN(d.getTime())) return '—'
      const pad = n => (n < 10 ? '0' + n : '' + n)
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
    },

    roundScore(v) {
      return Math.round(v || 0)
    },

    scoreColor(score) {
      if (score >= 80) return '#67c23a'
      if (score >= 60) return '#ff9900'
      return '#f56c6c'
    }
  }
}
</script>

<style scoped>
.report-share {
  max-width: 960px;
  margin: 0 auto;
  padding: 24px 16px;
  min-height: calc(100vh - 52px);
}

/* 状态横幅（与 Detail.vue 同源） */
.status-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 24px;
  border-radius: 8px;
  margin-bottom: 20px;
  font-size: 18px;
  font-weight: bold;
  color: #fff;
  flex-wrap: wrap;
}
.banner-pass    { background: linear-gradient(135deg, #52c41a, #389e0d); }
.banner-pending { background: linear-gradient(135deg, #ff9900, #d46b08); }
.banner-fail    { background: linear-gradient(135deg, #f5222d, #cf1322); }
.status-icon { font-size: 22px; }
.share-badge {
  margin-left: auto;
  font-size: 13px;
  font-weight: normal;
  background: rgba(255, 255, 255, 0.2);
  padding: 4px 12px;
  border-radius: 12px;
  white-space: nowrap;
}

/* 通用 card（与 Detail.vue 同源） */
.section-card { margin-bottom: 20px; }
.section-row  { margin-bottom: 0; }
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
}
.card-title {
  font-size: 16px;
  font-weight: bold;
  color: #111;
  border-left: 4px solid #ff9900;
  padding-left: 10px;
}

/* 综合评分区 */
.score-overview {
  display: flex;
  align-items: flex-start;
  gap: 40px;
  flex-wrap: wrap;
}
.score-circle-wrap { text-align: center; }
.grade-text {
  margin-top: 8px;
  font-size: 16px;
  font-weight: bold;
  color: #ff9900;
}
.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px 32px;
  flex: 1;
  align-content: center;
  min-width: 200px;
}
.info-item { display: flex; flex-direction: column; gap: 2px; }
.info-label { font-size: 12px; color: #999; }
.info-value { font-size: 15px; font-weight: bold; color: #333; }

/* 模块网格（与 Detail.vue 同源） */
.module-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}
.module-card {
  background: #fafafa;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  padding: 12px;
  text-align: center;
}
.module-name  { font-size: 13px; color: #555; margin-bottom: 4px; }
.module-score { font-size: 28px; font-weight: bold; margin-bottom: 2px; }
.module-meta  { font-size: 12px; color: #999; margin-bottom: 8px; }

/* AI 评价（与 Detail.vue 同源） */
.ai-block { margin-bottom: 16px; }
.ai-block-title {
  font-size: 15px;
  font-weight: bold;
  color: #333;
  margin-bottom: 10px;
}
.ai-list { list-style: none; padding: 0; margin: 0; }
.ai-list li {
  font-size: 14px;
  color: #444;
  padding: 6px 0;
  border-bottom: 1px solid #f5f5f5;
  line-height: 1.6;
}
.ai-list li:last-child { border-bottom: none; }
.list-icon { margin-right: 6px; }
.suggestion { color: #888; font-size: 13px; }
.roadmap-block {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}
.roadmap-text {
  font-size: 14px;
  color: #444;
  line-height: 1.8;
  background: #fff7e6;
  padding: 12px 16px;
  border-radius: 6px;
  border-left: 4px solid #ff9900;
}

/* 产品引导区 */
.cta-block {
  text-align: center;
  padding: 32px 20px;
  margin-bottom: 20px;
  background: linear-gradient(135deg, #131921, #232f3e);
  border-radius: 8px;
}
.cta-title {
  font-size: 20px;
  font-weight: bold;
  color: #fff;
  margin-bottom: 8px;
}
.cta-sub {
  font-size: 13px;
  color: #aab7c4;
  margin-bottom: 18px;
}

/* 异常态 */
.state-wrap { padding: 60px 0; }
.error-state { text-align: center; }
.error-icon { font-size: 48px; margin-bottom: 12px; }
.error-text { font-size: 14px; color: #666; margin-bottom: 18px; }

/* 响应式：≤768px 单列可读布局 */
@media (max-width: 768px) {
  .report-share { padding: 16px 12px; }
  .status-banner {
    font-size: 15px;
    padding: 12px 16px;
    gap: 8px;
  }
  .status-icon { font-size: 18px; }
  .share-badge { margin-left: 0; font-size: 12px; }
  .score-overview {
    flex-direction: column;
    align-items: center;
    gap: 20px;
  }
  .info-grid {
    width: 100%;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px 16px;
  }
  .module-grid { grid-template-columns: repeat(2, 1fr); }
  .cta-title { font-size: 17px; }
}
</style>
