<template>
  <div class="layout">
    <Sidebar :items="sidebarItems" />

    <div class="main-content">
      <div class="card">
        <div class="card-title">📁 我的面试记录</div>

        <!-- 加载中 -->
        <div v-if="loading" class="state-box">
          <i class="el-icon-loading" style="font-size:32px;color:#ff9900"></i>
          <p>加载中...</p>
        </div>

        <!-- 空状态 -->
        <div v-else-if="!list.length" class="state-box">
          <div style="font-size:48px">📋</div>
          <p style="color:#999;margin-top:8px">暂无面试记录</p>
          <el-button type="primary" class="btn-orange" @click="$router.push('/interview/config')">
            🚀 发起第一次面试
          </el-button>
        </div>

        <!-- 列表 -->
        <el-table v-else :data="list" style="width:100%" stripe>
          <el-table-column label="时间" width="160">
            <template slot-scope="{ row }">
              {{ formatTime(row.startTime || row.createdAt) }}
            </template>
          </el-table-column>
          <el-table-column prop="jobTitle" label="目标岗位" width="140"></el-table-column>
          <el-table-column label="轮次" width="120">
            <template slot-scope="{ row }">
              {{ roundLabel(row.round) }}
            </template>
          </el-table-column>
          <el-table-column prop="difficulty" label="难度" width="80"></el-table-column>
          <el-table-column label="得分" width="90">
            <template slot-scope="{ row }">
              <span :style="{ color: scoreColor(row.score), fontWeight: 'bold' }">
                {{ row.score != null ? row.score + '分' : '-' }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template slot-scope="{ row }">
              <el-tag :type="statusType(row.status)" size="mini">
                {{ statusLabel(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作">
            <template slot-scope="{ row }">
              <el-button
                v-if="row.status === 'completed'"
                type="text"
                class="op-btn"
                @click="viewReport(row)"
              >查看报告</el-button>
              <el-button
                v-if="row.status === 'paused' || row.status === 'in_progress'"
                type="text"
                class="op-btn"
                @click="continueInterview(row)"
              >继续面试</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script>
import Sidebar from '@/components/layout/Sidebar.vue'
import { getInterviewList, getInterview } from '@/api/interview'

const ROUND_MAP = { round1: '一面', round2: '二面', round3: '三面' }

export default {
  name: 'InterviewHistory',
  components: { Sidebar },

  data() {
    return {
      loading: false,
      list: [],
      sidebarItems: [
        {
          title: '面试中心',
          children: [
            { icon: '🚀', label: '发起面试', path: '/interview/config' },
            { icon: '📁', label: '我的记录', path: '/interview/history' }
          ]
        }
      ]
    }
  },

  created() {
    this.fetchList()
  },

  methods: {
    async fetchList() {
      this.loading = true
      try {
        const res = await getInterviewList({ page: 1, pageSize: 50 })
        const d = res?.data?.data || res?.data || res
        this.list = Array.isArray(d) ? d : (d.list || d.items || [])
      } catch (err) {
        this.$message.error('获取面试记录失败')
      } finally {
        this.loading = false
      }
    },

    viewReport(row) {
      this.$router.push(`/report/${row.interviewId || row.id}`)
    },

    async continueInterview(row) {
      const id = row.interviewId || row.id
      try {
        const res = await getInterview(id)
        const d = res?.data?.data || res?.data || res
        const questions = d.questions || []
        this.$store.commit('interview/SET_CURRENT_ID', id)
        this.$store.commit('interview/SET_INTERVIEW', d)
        this.$store.commit('interview/SET_QUESTIONS', questions)
        // 视频模式跳转到准备页重新检测设备
        const mode = d.mode || d.config?.mode || 'text'
        if (mode === 'video') {
          this.$store.commit('interview/SET_INTERVIEW_MODE', 'video')
          this.$router.push(`/interview/${id}/video-prep`)
        } else {
          this.$router.push(`/interview/${id}/doing`)
        }
      } catch (err) {
        this.$message.error('恢复面试失败，请重试')
      }
    },

    formatTime(ts) {
      if (!ts) return '-'
      const d = new Date(typeof ts === 'number' ? ts * 1000 : ts)
      if (isNaN(d.getTime())) return ts
      const pad = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
    },

    roundLabel(r) { return ROUND_MAP[r] || r || '-' },

    scoreColor(score) {
      if (score == null) return '#999'
      if (score >= 75) return '#67c23a'
      if (score >= 60) return '#ff9900'
      return '#f56c6c'
    },

    statusLabel(s) {
      const map = { completed: '已完成', paused: '已暂停', in_progress: '进行中', ongoing: '进行中', created: '未开始' }
      return map[s] || '进行中'
    },

    statusType(s) {
      const map = { completed: 'success', paused: 'warning', in_progress: 'primary', ongoing: 'primary', created: 'info' }
      return map[s] || 'primary'
    }
  }
}
</script>

<style scoped>
.layout { display: flex; min-height: calc(100vh - 90px); }
.main-content { flex: 1; padding: 20px; background: #f3f3f3; overflow: auto; }

.card {
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 20px;
}

.card-title {
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 20px;
  color: #111;
  border-left: 4px solid #ff9900;
  padding-left: 10px;
}

.state-box {
  text-align: center;
  padding: 60px 20px;
}

.op-btn { color: #ff9900 !important; }
.op-btn:hover { color: #f3a847 !important; }

.btn-orange {
  margin-top: 16px;
  background: #ff9900 !important;
  border-color: #ff9900 !important;
  color: #111 !important;
  font-weight: bold !important;
}

::v-deep .el-table th { background: #fafafa; }
</style>
