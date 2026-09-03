<template>
  <div class="dashboard">

    <!-- ① 简化的欢迎横幅 -->
    <div class="welcome-banner">
      <div class="welcome-content">
        <div class="welcome-text">
          <h1 class="welcome-title">{{ greeting }}，<span class="welcome-name">{{ displayName }}</span> 👋</h1>
          <p class="welcome-sub">{{ motivationalText }}</p>
        </div>
        <div class="welcome-actions">
          <el-button type="primary" size="medium" class="btn-primary-action" @click="$router.push('/interview/config')">
            🚀 开始面试
          </el-button>
          <el-button size="medium" class="btn-secondary-action" @click="$router.push('/help')">
            📖 新手引导
          </el-button>
        </div>
      </div>
    </div>

    <!-- ② 简洁功能入口 -->
    <div class="quick-access">
      <h2 class="section-title">快捷入口</h2>
      <div class="access-grid">
        <div class="access-card access-card--primary" @click="$router.push('/interview/config')">
          <div class="access-icon">🚀</div>
          <div class="access-text">
            <h3>发起面试</h3>
            <p>AI 模拟真实面试</p>
          </div>
        </div>
        <div class="access-card" @click="$router.push('/interview/history')">
          <div class="access-icon">📁</div>
          <div class="access-text">
            <h3>面试记录</h3>
            <p>查看历史表现</p>
          </div>
        </div>
        <div class="access-card" @click="$router.push('/questions')">
          <div class="access-icon">📝</div>
          <div class="access-text">
            <h3>题库练习</h3>
            <p>提升答题技巧</p>
          </div>
        </div>
        <div class="access-card" @click="$router.push('/company/intel')">
          <div class="access-icon">🏢</div>
          <div class="access-text">
            <h3>公司知识库</h3>
            <p>了解目标公司</p>
          </div>
        </div>
      </div>
    </div>

    <!-- ③ 核心数据（仅在有数据时显示） -->
    <div class="stats-section" v-if="hasStats && stats.totalInterviews > 0">
      <h2 class="section-title">📈 我的成长</h2>
      <div class="stats-row">
        <div class="stat-item">
          <span class="stat-num">{{ stats.totalInterviews || 0 }}</span>
          <span class="stat-label">次面试</span>
        </div>
        <div class="stat-item">
          <span class="stat-num">{{ (stats.avgScore || 0).toFixed(1) }}</span>
          <span class="stat-label">平均分</span>
        </div>
        <div class="stat-item">
          <span class="stat-num">{{ stats.highestScore || 0 }}</span>
          <span class="stat-label">最高分</span>
        </div>
        <div class="stat-item" v-if="stats.topPosition">
          <span class="stat-num stat-position">{{ stats.topPosition }}</span>
          <span class="stat-label">常练岗位</span>
        </div>
      </div>
    </div>

    <!-- ④ 近期面试记录（精简版） -->
    <div class="recent-section" v-if="!interviewsLoading">
      <div class="section-header">
        <h2 class="section-title">近期面试</h2>
        <router-link to="/interview/history" class="view-all" v-if="interviews.length > 0">查看全部 ›</router-link>
      </div>
      
      <div v-if="interviews.length === 0" class="empty-state">
        <div class="empty-icon">📋</div>
        <p class="empty-text">还没有面试记录</p>
        <el-button type="primary" size="small" @click="$router.push('/interview/config')">
          🚀 发起第一次面试
        </el-button>
      </div>
      
      <div v-else class="interview-list">
        <div 
          v-for="item in interviews.slice(0, 3)" 
          :key="item.interviewId || item.id" 
          class="interview-item"
        >
          <div class="interview-info">
            <div class="interview-position">{{ item.jobTitle || item.position || '—' }}</div>
            <div class="interview-meta">
              <span class="interview-time">{{ formatTime(item.startTime || item.createdAt) }}</span>
              <el-tag size="mini" :type="statusTagType(item.status)">{{ statusLabel(item.status) }}</el-tag>
            </div>
          </div>
          <div class="interview-score" v-if="item.totalScore ?? item.score">
            <span :class="scoreClass(item.totalScore ?? item.score)">
              {{ item.totalScore ?? item.score }}
            </span>
          </div>
          <el-button 
            v-if="item.status === 'completed'" 
            type="text" 
            size="mini" 
            @click="goReport(item.interviewId || item.id)"
          >
            查看报告
          </el-button>
        </div>
      </div>
    </div>

    <!-- ⑤ 推荐题目（精简） -->
    <div class="recommend-section" v-if="!questionsLoading && questions.length > 0">
      <div class="section-header">
        <h2 class="section-title">推荐练习</h2>
        <router-link to="/questions" class="view-all">更多 ›</router-link>
      </div>
      <div class="recommend-list">
        <div 
          v-for="q in questions.slice(0, 3)" 
          :key="q.questionId || q.id" 
          class="recommend-item"
          @click="goPractice(q.questionId || q.id)"
        >
          <span class="recommend-title">{{ q.content || q.title }}</span>
          <el-tag size="mini" :type="difficultyTagType(q.difficulty)">
            {{ difficultyLabel(q.difficulty) }}
          </el-tag>
        </div>
      </div>
    </div>

  </div>
</template>

<script>
import request from '@/api/request'
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
    hasStats() {
      return this.stats.totalInterviews > 0 || this.stats.avgScore > 0
    }
  },
  data() {
    return {
      stats: { totalInterviews: 0, avgScore: 0, highestScore: 0, topPosition: '' },
      interviewsLoading: true,
      interviews: [],
      questionsLoading: true,
      questions: []
    }
  },
  created() {
    this.loadData()
  },
  methods: {
    async loadData() {
      const [interviewsRes, questionsRes, statsRes] = await Promise.allSettled([
        request.get('/interviews', { params: { page: 1, pageSize: 5 } }),
        request.get('/questions', { params: { page: 1, pageSize: 4 } }),
        request.get('/profile/stats')
      ])

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

      if (statsRes.status === 'fulfilled') {
        const d = statsRes.value.data || statsRes.value
        this.stats = {
          totalInterviews: d.totalCount ?? d.totalInterviews ?? 0,
          avgScore:        d.avgScore ?? d.avg_score ?? 0,
          highestScore:    d.maxScore ?? d.highestScore ?? 0,
          topPosition:     d.topJobTitle || d.topPosition || ''
        }
      }
    },
    formatTime(val) {
      if (!val) return '—'
      const d = new Date(typeof val === 'number' && val < 10000000000 ? val * 1000 : val)
      if (isNaN(d.getTime())) return val
      const p = n => String(n).padStart(2, '0')
      return `${p(d.getMonth()+1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
    },
    scoreClass(s) {
      if (s == null) return ''
      if (s >= 80) return 'score-high'
      if (s >= 60) return 'score-mid'
      return 'score-low'
    },
    statusTagType(s) { return { completed: 'success', in_progress: 'warning', pending: 'info', failed: 'danger' }[s] || 'info' },
    statusLabel(s)   { return { completed: '已完成', in_progress: '进行中', pending: '待开始', failed: '已中断', ongoing: '进行中', paused: '已暂停' }[s] || s || '未知' },
    difficultyTagType(d) { return { easy: 'success', medium: 'warning', hard: 'danger' }[d] || 'info' },
    difficultyLabel(d)   { return { easy: '简单', medium: '中等', hard: '困难' }[d] || d || '未知' },
    goReport(id)    { this.$router.push(`/report/${id}`) },
    goPractice(id) { this.$router.push(`/questions/${id}/practice`) }
  }
}
</script>

<style scoped>
.dashboard { padding: 20px; max-width: 1000px; margin: 0 auto; }

/* 欢迎横幅 */
.welcome-banner {
  background: linear-gradient(135deg, #131921 0%, #232f3e 60%, #3a4a5c 100%);
  border-radius: 12px;
  padding: 32px 36px;
  margin-bottom: 24px;
}
.welcome-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 20px;
}
.welcome-title {
  font-size: 26px;
  color: #fff;
  margin: 0 0 8px 0;
  font-weight: 600;
}
.welcome-name { color: #ff9900; }
.welcome-sub {
  font-size: 15px;
  color: #aab7c4;
  margin: 0;
}
.welcome-actions { display: flex; gap: 12px; }
.btn-primary-action {
  background: #ff9900 !important;
  border-color: #ff9900 !important;
  color: #111 !important;
  font-weight: bold;
}
.btn-secondary-action {
  background: rgba(255,255,255,0.1) !important;
  border-color: rgba(255,255,255,0.3) !important;
  color: #fff !important;
}
.btn-secondary-action:hover {
  background: rgba(255,255,255,0.2) !important;
}

/* 通用标题 */
.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #111;
  margin: 0 0 16px 0;
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.section-header .section-title { margin-bottom: 0; }
.view-all {
  font-size: 13px;
  color: #0066c0;
  text-decoration: none;
}
.view-all:hover { color: #ff9900; }

/* 功能入口 */
.quick-access {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 20px;
  border: 1px solid #eee;
}
.access-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.access-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px 16px;
  border: 1px solid #eee;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}
.access-card:hover {
  border-color: #ff9900;
  box-shadow: 0 4px 12px rgba(255, 153, 0, 0.12);
  transform: translateY(-2px);
}
.access-card--primary {
  background: linear-gradient(135deg, #fff8e6 0%, #fff3d9 100%);
  border-color: #ffe4a0;
}
.access-card--primary:hover {
  border-color: #ff9900;
  box-shadow: 0 4px 16px rgba(255, 153, 0, 0.2);
}
.access-icon {
  font-size: 32px;
  line-height: 1;
}
.access-text h3 {
  font-size: 15px;
  color: #111;
  margin: 0 0 4px 0;
  font-weight: 600;
}
.access-text p {
  font-size: 12px;
  color: #888;
  margin: 0;
}

/* 数据统计 */
.stats-section {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 20px;
  border: 1px solid #eee;
}
.mini-stat {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px;
  background: #fafafa;
  border-radius: 10px;
}
.mini-stat-icon { font-size: 28px; }
.mini-stat-num {
  font-size: 22px;
  font-weight: bold;
  color: #111;
  line-height: 1.2;
}
.mini-stat-label {
  font-size: 12px;
  color: #888;
}
.score-high { color: #067d62; font-weight: bold; }
.score-mid  { color: #ff9900; font-weight: bold; }
.score-low  { color: #c7511f; font-weight: bold; }

/* 近期面试 */
.recent-section {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 20px;
  border: 1px solid #eee;
}
.empty-state {
  text-align: center;
  padding: 40px 20px;
}
.empty-icon { font-size: 48px; margin-bottom: 12px; }
.empty-text { font-size: 15px; color: #888; margin-bottom: 16px; }
.interview-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.interview-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 14px 16px;
  background: #fafafa;
  border-radius: 10px;
  transition: background 0.15s;
}
.interview-item:hover { background: #f5f5f5; }
.interview-info { flex: 1; }
.interview-position {
  font-size: 15px;
  font-weight: 500;
  color: #111;
  margin-bottom: 4px;
}
.interview-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  color: #888;
}
.interview-score {
  font-size: 20px;
  font-weight: bold;
  min-width: 50px;
  text-align: center;
}

/* 推荐题目 */
.recommend-section {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  border: 1px solid #eee;
}
.recommend-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.recommend-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  background: #fafafa;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
}
.recommend-item:hover {
  background: #fff3e0;
  transform: translateX(4px);
}
.recommend-title {
  font-size: 14px;
  color: #0066c0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 覆盖 Element UI 按钮 */
::v-deep .el-button--primary {
  background: #ff9900;
  border-color: #ff9900;
  color: #111;
}
::v-deep .el-button--primary:hover {
  background: #f3a847;
  border-color: #f3a847;
}

/* 统计区域 - 简化版 */
.stats-section {
  background: #fff;
  border-radius: 12px;
  padding: 20px 24px;
  margin-bottom: 20px;
  border: 1px solid #eee;
}
.stats-row {
  display: flex;
  gap: 32px;
  flex-wrap: wrap;
}
.stat-item {
  display: flex;
  align-items: baseline;
  gap: 6px;
}
.stat-num {
  font-size: 24px;
  font-weight: 600;
  color: #ff9900;
}
.stat-position {
  font-size: 18px;
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.stat-label {
  font-size: 13px;
  color: #888;
}

/* 响应式 */
@media (max-width: 768px) {
  .access-grid { grid-template-columns: repeat(2, 1fr); }
  .welcome-content { flex-direction: column; align-items: flex-start; }
  .welcome-title { font-size: 22px; }
  .stats-row { gap: 20px; }
}
</style>
