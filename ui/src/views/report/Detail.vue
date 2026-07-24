<template>
  <div class="report-detail" v-loading="loading">
    <!-- 顶部状态横幅 -->
    <div v-if="report" :class="['status-banner', statusClass]">
      <span class="status-icon">{{ statusIcon }}</span>
      <span class="status-text">{{ statusText }}</span>
      <span v-if="report.passReason" class="status-reason">（{{ report.passReason }}）</span>
    </div>

    <div v-if="report" class="report-body">
      <!-- 综合评分区 -->
      <el-card shadow="hover" class="section-card">
        <div slot="header" class="card-header">
          <span class="card-title">📊 综合评分</span>
          <div style="display:flex;align-items:center;gap:8px">
            <el-tag v-if="report.mode === 'video'" type="warning" size="small">
              <i class="el-icon-video-camera"></i> 视频面试
            </el-tag>
            <el-button type="primary" size="small" @click="handleShare" :loading="shareLoading">
              🔗 生成分享链接
            </el-button>
          </div>
        </div>
        <div class="score-overview">
          <div class="score-circle-wrap">
            <score-circle :score="report.totalScore || 0" :size="160" />
            <div class="grade-text">{{ report.grade }}</div>
          </div>
          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">目标岗位</span>
              <span class="info-value">{{ report.jobTitle }}</span>
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
              <span class="info-label">用时</span>
              <span class="info-value">{{ formatDuration(report.totalSeconds) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">作答题数</span>
              <span class="info-value">{{ report.answeredCount }} / {{ report.totalCount }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">跳过题数</span>
              <span class="info-value">{{ report.skippedCount }}</span>
            </div>
          </div>
        </div>
        <!-- 分享链接显示 -->
        <div v-if="shareUrl" class="share-box">
          <el-input v-model="shareUrl" readonly size="small" style="flex:1" />
          <el-button size="small" @click="copyShareUrl">复制链接</el-button>
        </div>
      </el-card>

      <!-- 雷达图 + 模块得分 -->
      <el-row :gutter="16" class="section-row">
        <el-col :span="12">
          <el-card shadow="hover" class="section-card">
            <div slot="header" class="card-header">
              <span class="card-title">🕸️ 知识点雷达图</span>
            </div>
            <radar-chart :modules="radarModules" />
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="hover" class="section-card">
            <div slot="header" class="card-header">
              <span class="card-title">🧩 知识点模块得分</span>
            </div>
            <div class="module-grid">
              <div v-for="(m, i) in report.moduleScores" :key="i" class="module-card">
                <div class="module-name">{{ m.module }}</div>
                <div class="module-score" :style="{ color: scoreColor(m.avgScore) }">
                  {{ m.avgScore }}
                </div>
                <div class="module-meta">共 {{ m.count }} 题 · {{ m.level }}</div>
                <score-bar :score="m.avgScore" />
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 表达能力分析（仅视频模式） -->
      <el-card v-if="report.mode === 'video'" shadow="hover" class="section-card">
        <div slot="header" class="card-header">
          <span class="card-title">🎤 表达能力分析</span>
        </div>
        <el-row :gutter="24">
          <el-col :span="8" class="expr-col">
            <score-circle :score="report.avgExpressionScore || 0" :size="100" />
            <div class="expr-label">平均表达得分</div>
          </el-col>
          <el-col :span="8" class="expr-col">
            <div class="expr-value" :class="speechRateClass">
              {{ report.avgSpeechRate ? report.avgSpeechRate.toFixed(0) : '—' }}
              <small> 字/分钟</small>
            </div>
            <div class="expr-label">平均语速</div>
            <div class="expr-hint" :class="speechRateClass">{{ speechRateHint }}</div>
          </el-col>
          <el-col :span="8" class="expr-col">
            <div class="expr-value">{{ totalPauses }}</div>
            <div class="expr-label">总停顿次数</div>
          </el-col>
        </el-row>
        <div v-if="report.expressionSummary" class="expr-summary">
          {{ report.expressionSummary }}
        </div>
      </el-card>

      <!-- 视频下载入口（Task 12.3） -->
      <el-card v-if="report.mode === 'video' && recordedVideo" shadow="hover" class="section-card">
        <div slot="header" class="card-header">
          <span class="card-title">📹 面试录制</span>
        </div>
        <div class="video-download-row">
          <span class="video-meta">大小：{{ recordedVideo.size }}，时长：{{ recordedVideo.duration }}秒</span>
          <el-button type="primary" size="small" @click="downloadRecordedVideo">下载面试录制</el-button>
        </div>
        <el-alert
          type="warning"
          :closable="false"
          show-icon
          title="视频仅存储于本地，刷新页面后将丢失"
          style="margin-top:10px"
        ></el-alert>
      </el-card>

      <!-- 题目明细 -->
      <el-card shadow="hover" class="section-card">
        <div slot="header" class="card-header">
          <span class="card-title">📝 逐题得分明细</span>
          <div class="filter-tabs">
            <el-radio-group v-model="filterTab" size="mini">
              <el-radio-button label="all">全部</el-radio-button>
              <el-radio-button label="answered">已作答</el-radio-button>
              <el-radio-button label="skipped">已跳过</el-radio-button>
              <el-radio-button label="high">高分≥80</el-radio-button>
              <el-radio-button label="low">低分&lt;60</el-radio-button>
            </el-radio-group>
          </div>
        </div>
        <el-collapse v-model="expandedItems" accordion>
          <el-collapse-item
            v-for="q in filteredQuestions"
            :key="q.index"
            :name="String(q.index)"
          >
            <template slot="title">
              <div class="q-title-row">
                <span class="q-num">Q{{ q.index + 1 }}</span>
                <el-tag v-if="q.skipped" size="mini" type="info" class="q-tag">已跳过</el-tag>
                <el-tag v-else size="mini" type="success" class="q-tag">已作答</el-tag>
                <span class="q-preview">{{ truncate(q.content, 35) }}</span>
                <div class="q-score-bar">
                  <score-bar :score="q.skipped ? 0 : (q.score || 0)" />
                </div>
              </div>
            </template>
            <div class="q-detail">
              <div class="q-section">
                <div class="q-section-label">题目内容</div>
                <div class="q-content">{{ q.content }}</div>
                <div class="q-tags">
                  <el-tag v-for="tag in (q.tags || [])" :key="tag" size="mini" type="info" class="q-tag-item">{{ tag }}</el-tag>
                  <el-tag size="mini" :type="diffTagType(q.difficulty)" class="q-tag-item">{{ q.difficulty }}</el-tag>
                </div>
              </div>
              <div class="q-section" v-if="!q.skipped">
                <div class="q-section-label">我的作答</div>
                <div class="q-answer">{{ q.userAnswer || '（未填写）' }}</div>
              </div>
              <div class="q-section" v-else>
                <div class="q-section-label">作答状态</div>
                <div class="q-answer skipped-tip">已跳过此题（0分）</div>
              </div>
              <div class="q-section" v-if="!q.skipped">
                <div class="q-section-label">AI 点评</div>
                <div class="q-review">
                  <div v-if="q.pros && q.pros.length" class="review-block">
                    <span class="review-label pros">✅ 优点</span>
                    <ul>
                      <li v-for="(p, pi) in q.pros" :key="pi">{{ p }}</li>
                    </ul>
                  </div>
                  <div v-if="q.cons && q.cons.length" class="review-block">
                    <span class="review-label cons">⚠️ 不足</span>
                    <ul>
                      <li v-for="(c, ci) in q.cons" :key="ci">{{ c }}</li>
                    </ul>
                  </div>
                  <!-- 视频模式：表达得分与反馈 -->
                  <div v-if="report.mode === 'video'" class="review-block">
                    <div class="expr-detail-row">
                      <span class="review-label" style="color:#409eff">🎤 表达得分</span>
                      <score-bar :score="q.expressionScore || 0" />
                      <span style="color:#409eff;font-weight:bold;margin-left:8px">{{ q.expressionScore || 0 }}分</span>
                    </div>
                    <p v-if="q.expressionFeedback" class="expr-feedback-text">{{ q.expressionFeedback }}</p>
                    <div v-if="q.nonVerbalMetrics" class="nonverbal-stats">
                      <span>语速：{{ q.nonVerbalMetrics.speechRate }} 字/分钟</span>
                      <span>停顿：{{ q.nonVerbalMetrics.pauseCount }} 次</span>
                      <span>用时：{{ q.nonVerbalMetrics.duration }} 秒</span>
                    </div>
                  </div>
                </div>
              </div>
              <el-collapse class="ref-collapse">
                <el-collapse-item title="📖 查看参考答案" name="ref">
                  <div class="q-ref-answer">{{ q.referenceAnswer }}</div>
                </el-collapse-item>
              </el-collapse>
            </div>
          </el-collapse-item>
        </el-collapse>
        <div v-if="filteredQuestions.length === 0" class="empty-tip">暂无符合筛选条件的题目</div>
      </el-card>

      <!-- AI 综合评价 -->
      <el-card shadow="hover" class="section-card" v-if="report.aiSummary">
        <div slot="header" class="card-header">
          <span class="card-title">🤖 AI 综合评价与建议</span>
        </div>
        <el-row :gutter="16">
          <el-col :span="12">
            <div class="ai-block">
              <div class="ai-block-title">💪 优势亮点</div>
              <ul class="ai-list strengths-list">
                <li v-for="(s, i) in (report.aiSummary.strengths || [])" :key="i">
                  <span class="list-icon">✅</span> {{ s }}
                </li>
              </ul>
            </div>
          </el-col>
          <el-col :span="12">
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

    </div><!-- end report-body -->

    <!-- 空状态 -->
    <el-empty v-if="!loading && !report" description="报告加载失败，请返回重试">
      <el-button type="primary" @click="$router.back()">返回</el-button>
    </el-empty>

  </div>
</template>

<script>
import { getReport, createShareLink } from '@/api/report'
import ScoreCircle from '@/components/common/ScoreCircle.vue'
import RadarChart from '@/components/common/RadarChart.vue'
import ScoreBar from '@/components/common/ScoreBar.vue'

export default {
  name: 'ReportDetail',
  components: { ScoreCircle, RadarChart, ScoreBar },

  data() {
    return {
      loading: true,
      report: null,
      filterTab: 'all',
      expandedItems: [],
      shareLoading: false,
      shareUrl: ''
    }
  },

  computed: {
    statusClass() {
      if (!this.report) return ''
      const s = this.report.totalScore || 0
      if (s >= 75) return 'banner-pass'
      if (s >= 60) return 'banner-pending'
      return 'banner-fail'
    },
    statusIcon() {
      if (!this.report) return ''
      const s = this.report.totalScore || 0
      if (s >= 75) return '✅'
      if (s >= 60) return '⚠️'
      return '❌'
    },
    statusText() {
      if (!this.report) return ''
      const s = this.report.totalScore || 0
      if (s >= 75) return '恭喜，模拟面试通过！'
      if (s >= 60) return '发挥一般，建议加强练习'
      return '本次未通过，继续努力！'
    },
    radarModules() {
      if (!this.report || !this.report.moduleScores) return []
      const modules = this.report.moduleScores.map(m => ({ name: m.module, score: m.avgScore || 0 }))
      // 视频模式：追加「表达能力」维度
      if (this.report.mode === 'video' && this.report.avgExpressionScore) {
        modules.push({ name: '表达能力', score: this.report.avgExpressionScore })
      }
      return modules
    },
    filteredQuestions() {
      if (!this.report || !this.report.questions) return []
      const qs = this.report.questions
      switch (this.filterTab) {
        case 'answered': return qs.filter(q => !q.skipped)
        case 'skipped':  return qs.filter(q => q.skipped)
        case 'high':     return qs.filter(q => !q.skipped && (q.score || 0) >= 80)
        case 'low':      return qs.filter(q => q.skipped || (q.score || 0) < 60)
        default:         return qs
      }
    },
    // 视频模式专用
    speechRateClass() {
      const rate = this.report && this.report.avgSpeechRate
      if (!rate) return 'rate-normal'
      if (rate >= 120 && rate <= 150) return 'rate-good'
      if (rate < 100 || rate > 180) return 'rate-bad'
      return 'rate-warn'
    },
    speechRateHint() {
      const rate = this.report && this.report.avgSpeechRate
      if (!rate) return ''
      if (rate >= 120 && rate <= 150) return '✓ 语速适中'
      if (rate < 100) return '⚠ 语速偏慢'
      if (rate > 180) return '⚠ 语速偏快'
      return '接近推荐范围'
    },
    totalPauses() {
      if (!this.report || !this.report.questions) return 0
      return this.report.questions.reduce((sum, q) => {
        return sum + (q.nonVerbalMetrics ? q.nonVerbalMetrics.pauseCount : 0)
      }, 0)
    },
    recordedVideo() {
      return this.$store && this.$store.state.interview && this.$store.state.interview.recordedVideo
    }
  },

  mounted() {
    this.loadReport()
  },

  methods: {
    async loadReport() {
      const id = this.$route.params.interviewId
      try {
        const res = await getReport(id)
        this.report = res.data || res
      } catch (e) {
        this.$message.error('报告加载失败，请稍后重试')
      } finally {
        this.loading = false
      }
    },

    async handleShare() {
      const id = this.$route.params.interviewId
      this.shareLoading = true
      try {
        const res = await createShareLink(id)
        const data = res.data || res
        const token = data.shareToken || data.token || data.share_token || ''
        this.shareUrl = `${window.location.origin}/report/share/${token}`
        this.$message.success('分享链接已生成')
      } catch (e) {
        this.$message.error('生成分享链接失败')
      } finally {
        this.shareLoading = false
      }
    },

    copyShareUrl() {
      if (!this.shareUrl) return
      if (navigator.clipboard) {
        navigator.clipboard.writeText(this.shareUrl).then(() => {
          this.$message.success('链接已复制到剪贴板')
        })
      } else {
        const el = document.createElement('textarea')
        el.value = this.shareUrl
        document.body.appendChild(el)
        el.select()
        document.execCommand('copy')
        document.body.removeChild(el)
        this.$message.success('链接已复制到剪贴板')
      }
    },

    roundLabel(round) {
      const map = { round1: '一面', round2: '二面', round3: '三面' }
      return map[round] || round || '—'
    },

    difficultyLabel(diff) {
      const map = { junior: '初级', middle: '中级', senior: '高级' }
      return map[diff] || diff || '—'
    },

    diffTagType(diff) {
      const map = { easy: 'success', medium: 'warning', hard: 'danger' }
      return map[diff] || 'info'
    },

    formatDuration(secs) {
      if (!secs) return '—'
      const m = Math.floor(secs / 60)
      const s = secs % 60
      return `${m}分${s}秒`
    },

    truncate(str, len) {
      if (!str) return ''
      return str.length > len ? str.slice(0, len) + '…' : str
    },

    scoreColor(score) {
      if (score >= 80) return '#67c23a'
      if (score >= 60) return '#ff9900'
      return '#f56c6c'
    },

    downloadRecordedVideo() {
      const video = this.recordedVideo
      if (!video || !video.url) return
      const a = document.createElement('a')
      a.href = video.url
      a.download = `interview-${this.$route.params.interviewId}.webm`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
    }
  }
}
</script>

<style scoped>
.report-detail {
  max-width: 960px;
  margin: 0 auto;
  padding: 24px 16px;
}

/* 状态横幅 */
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
}
.banner-pass    { background: linear-gradient(135deg, #52c41a, #389e0d); }
.banner-pending { background: linear-gradient(135deg, #ff9900, #d46b08); }
.banner-fail    { background: linear-gradient(135deg, #f5222d, #cf1322); }
.status-icon { font-size: 22px; }
.status-reason { font-size: 14px; font-weight: normal; opacity: 0.9; }

/* 通用 card */
.section-card { margin-bottom: 20px; }
.section-row   { margin-bottom: 20px; }
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
  grid-template-columns: repeat(3, 1fr);
  gap: 16px 32px;
  flex: 1;
}
.info-item { display: flex; flex-direction: column; gap: 2px; }
.info-label { font-size: 12px; color: #999; }
.info-value { font-size: 15px; font-weight: bold; color: #333; }

/* 分享框 */
.share-box {
  display: flex;
  gap: 8px;
  margin-top: 16px;
  align-items: center;
}

/* 模块网格 */
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

/* 题目列表筛选 */
.filter-tabs { }

/* 题目折叠 */
.q-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  flex-wrap: nowrap;
  overflow: hidden;
}
.q-num { font-weight: bold; color: #ff9900; min-width: 32px; }
.q-tag { flex-shrink: 0; }
.q-preview {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
  color: #333;
}
.q-score-bar { min-width: 160px; }

/* 题目详情 */
.q-detail { padding: 8px 0; }
.q-section { margin-bottom: 14px; }
.q-section-label {
  font-size: 13px;
  font-weight: bold;
  color: #666;
  margin-bottom: 6px;
}
.q-content {
  font-size: 15px;
  color: #222;
  line-height: 1.7;
  background: #fafafa;
  padding: 10px 12px;
  border-radius: 6px;
}
.q-tags { margin-top: 8px; display: flex; flex-wrap: wrap; gap: 6px; }
.q-tag-item { }
.q-answer {
  font-size: 14px;
  color: #444;
  background: #fafafa;
  padding: 10px 12px;
  border-radius: 6px;
  white-space: pre-wrap;
  line-height: 1.6;
}
.skipped-tip { color: #999; font-style: italic; }
.q-review { }
.review-block { margin-bottom: 8px; }
.review-label { font-size: 13px; font-weight: bold; margin-right: 6px; }
.review-label.pros { color: #52c41a; }
.review-label.cons { color: #ff9900; }
.q-review ul { margin: 4px 0 0 20px; padding: 0; }
.q-review li { font-size: 14px; color: #555; margin-bottom: 3px; }
.ref-collapse { margin-top: 8px; }
.q-ref-answer {
  font-size: 14px;
  color: #444;
  line-height: 1.8;
  white-space: pre-wrap;
  background: #f6ffed;
  padding: 10px 12px;
  border-radius: 6px;
}

/* AI 评价 */
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

.empty-tip {
  text-align: center;
  color: #999;
  padding: 20px 0;
  font-size: 13px;
}

.report-body {}

/* 表达能力分析卡片 */
.expr-col {
  text-align: center;
  padding: 12px;
}
.expr-value {
  font-size: 36px;
  font-weight: bold;
  color: #333;
  line-height: 1.2;
}
.expr-value small { font-size: 14px; font-weight: normal; color: #999; }
.expr-label { font-size: 13px; color: #666; margin-top: 6px; }
.expr-hint { font-size: 12px; margin-top: 4px; }
.rate-good { color: #67c23a; }
.rate-warn { color: #ff9900; }
.rate-bad  { color: #f56c6c; }
.rate-normal { color: #999; }
.expr-summary {
  margin-top: 16px;
  padding: 12px 16px;
  background: #f6f9ff;
  border-radius: 6px;
  font-size: 14px;
  color: #444;
  line-height: 1.8;
  border-left: 4px solid #409eff;
}

/* 视频下载 */
.video-download-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.video-meta { font-size: 13px; color: #666; }

/* 表达详情 */
.expr-detail-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}
.expr-feedback-text {
  font-size: 13px;
  color: #444;
  line-height: 1.7;
  margin: 4px 0;
  padding: 6px 10px;
  background: #f0f5ff;
  border-radius: 4px;
}
.nonverbal-stats {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #777;
  margin-top: 4px;
}
</style>
