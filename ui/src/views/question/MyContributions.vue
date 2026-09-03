<template>
  <div class="my-contributions-page">
    <div class="page-header">
      <h2 class="page-title">🏆 我的贡献记录</h2>
      <p class="page-sub">你贡献的所有真实面试题目，帮助更多人准备面试</p>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stats-row">
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-value">{{ stats.total }}</div>
          <div class="stat-label">贡献题目</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-value">{{ stats.verified }}</div>
          <div class="stat-label">已验证</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-value">{{ stats.approved }}</div>
          <div class="stat-label">已通过</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-value">{{ stats.pending }}</div>
          <div class="stat-label">待审核</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 积分信息 -->
    <div class="credit-bar">
      <span>💎 当前积分：<strong>{{ credits }}</strong></span>
      <el-button size="mini" type="primary" @click="$router.push('/question/contribute')">
        <i class="el-icon-plus"></i> 继续贡献
      </el-button>
    </div>

    <!-- 贡献列表 -->
    <el-card shadow="never" v-loading="loading">
      <div v-if="!loading && contributions.length === 0" class="empty-state">
        <div class="empty-icon">📝</div>
        <div class="empty-title">还没有贡献记录</div>
        <div class="empty-sub">分享你在面试中遇到的真实题目，帮助更多人</div>
        <el-button type="primary" @click="$router.push('/question/contribute')">去贡献题目</el-button>
      </div>

      <el-table v-else :data="contributions" stripe>
        <el-table-column label="公司" prop="company" width="140" />
        <el-table-column label="岗位" prop="jobTitle" width="130" />
        <el-table-column label="题目内容">
          <template slot-scope="{ row }">
            <div class="q-content">{{ row.content }}</div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template slot-scope="{ row }">
            <el-tag :type="statusType(row.status)" size="mini">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="90">
          <template slot-scope="{ row }">
            <el-tag size="mini" type="info">{{ typeText(row.questionType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="难度" width="80">
          <template slot-scope="{ row }">
            <el-tag size="mini" :type="diffType(row.difficulty)">{{ diffText(row.difficulty) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="有用票" width="80">
          <template slot-scope="{ row }">
            <span class="stat-num">👍 {{ row.helpfulCount || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="被使用" width="80">
          <template slot-scope="{ row }">
            <span class="stat-num">🚀 {{ row.useCount || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="贡献时间" width="120">
          <template slot-scope="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script>
import { getUserContributions, getUserCredits } from '@/api/contribution'

export default {
  name: 'MyContributions',
  data() {
    return {
      loading: true,
      contributions: [],
      credits: 0
    }
  },
  computed: {
    stats() {
      const list = this.contributions
      return {
        total: list.length,
        verified: list.filter(q => q.status === 'verified').length,
        approved: list.filter(q => q.status === 'approved').length,
        pending: list.filter(q => q.status === 'pending').length
      }
    }
  },
  created() {
    this.loadData()
  },
  methods: {
    async loadData() {
      this.loading = true
      try {
        const [contribRes, creditRes] = await Promise.all([
          getUserContributions(),
          getUserCredits()
        ])
        this.contributions = contribRes?.contributions || []
        this.credits = creditRes?.credits || 0
      } catch (e) {
        this.$message.error('加载失败')
      } finally {
        this.loading = false
      }
    },
    statusType(status) {
      return { pending: 'warning', approved: 'success', verified: 'primary', rejected: 'danger' }[status] || 'info'
    },
    statusText(status) {
      return { pending: '待审核', approved: '已通过', verified: '已验证', rejected: '已驳回' }[status] || status
    },
    typeText(t) {
      return { technical: '技术', behavioral: '行为', algorithm: '算法', 'system-design': '系统设计' }[t] || t
    },
    diffType(d) {
      return { easy: 'success', medium: 'warning', hard: 'danger' }[d] || 'info'
    },
    diffText(d) {
      return { easy: '简单', medium: '中等', hard: '困难' }[d] || d
    },
    formatDate(ts) {
      if (!ts) return '—'
      const d = new Date(ts * 1000)
      return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`
    }
  }
}
</script>

<style scoped>
.my-contributions-page { max-width: 1000px; margin: 0 auto; padding: 20px 16px; }
.page-header { margin-bottom: 20px; }
.page-title { font-size: 20px; font-weight: bold; color: #333; margin-bottom: 8px; }
.page-sub { font-size: 13px; color: #666; }

/* Stats */
.stats-row { margin-bottom: 16px; }
.stat-card {
  text-align: center;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
}
.stat-value { font-size: 28px; font-weight: bold; color: #ff9900; }
.stat-label { font-size: 13px; color: #888; margin-top: 4px; }

/* Credit bar */
.credit-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  background: #fff8e6;
  border: 1px solid #ffe58f;
  border-radius: 6px;
  padding: 10px 16px;
  margin-bottom: 16px;
  font-size: 14px;
  color: #555;
}

/* Table */
.q-content {
  font-size: 13px;
  color: #333;
  line-height: 1.5;
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.stat-num { font-size: 12px; color: #888; }

/* Empty */
.empty-state { text-align: center; padding: 60px 0; }
.empty-icon { font-size: 56px; }
.empty-title { font-size: 18px; font-weight: bold; color: #555; margin: 16px 0 8px; }
.empty-sub { font-size: 14px; color: #999; margin-bottom: 20px; }
</style>
