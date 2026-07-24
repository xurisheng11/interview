<template>
  <div class="profile-page">
    <el-tabs v-model="activeTab" type="border-card">

      <!-- Tab 1: 个人信息 -->
      <el-tab-pane label="👤 个人信息" name="info">
        <el-row :gutter="24">
          <el-col :span="14">
            <el-card shadow="never">
              <div slot="header" class="card-title">基本信息</div>
              <el-form :model="profileForm" :rules="profileRules" ref="profileForm" label-width="80px" v-loading="profileLoading">

                <!-- 头像上传 -->
                <el-form-item label="头像">
                  <div class="avatar-upload-wrap">
                    <div class="avatar-preview" @click="triggerFileInput">
                      <img v-if="profileForm.avatar" :src="profileForm.avatar" class="avatar-img" />
                      <div v-else class="avatar-placeholder">
                        <i class="el-icon-camera"></i>
                        <span>点击上传</span>
                      </div>
                      <div class="avatar-overlay">
                        <i class="el-icon-camera"></i> 更换
                      </div>
                    </div>
                    <div class="avatar-tip">点击头像上传，支持 JPG / PNG，建议正方形图片</div>
                    <input
                      ref="fileInput"
                      type="file"
                      accept="image/jpeg,image/png,image/gif,image/webp"
                      style="display:none"
                      @change="onFileChange"
                    />
                  </div>
                </el-form-item>

                <el-form-item label="昵称" prop="nickname">
                  <el-input v-model="profileForm.nickname" placeholder="请输入昵称" clearable />
                </el-form-item>
                <el-form-item label="个人简介">
                  <el-input v-model="profileForm.bio" type="textarea" :rows="3" placeholder="介绍一下自己..." resize="none" />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" :loading="savingProfile" @click="saveProfile">保存修改</el-button>
                </el-form-item>
              </el-form>
            </el-card>
          </el-col>
          <el-col :span="10">
            <el-card shadow="never">
              <div slot="header" class="card-title">修改密码</div>
              <el-form :model="pwdForm" :rules="pwdRules" ref="pwdForm" label-width="90px">
                <el-form-item label="当前密码" prop="oldPassword">
                  <el-input v-model="pwdForm.oldPassword" type="password" placeholder="请输入当前密码" show-password />
                </el-form-item>
                <el-form-item label="新密码" prop="newPassword">
                  <el-input v-model="pwdForm.newPassword" type="password" placeholder="不少于8位" show-password />
                </el-form-item>
                <el-form-item label="确认新密码" prop="confirmPassword">
                  <el-input v-model="pwdForm.confirmPassword" type="password" placeholder="再次输入新密码" show-password />
                </el-form-item>
                <el-form-item>
                  <el-button type="warning" :loading="savingPwd" @click="changePassword">修改密码</el-button>
                </el-form-item>
              </el-form>
            </el-card>
          </el-col>
        </el-row>
      </el-tab-pane>

      <!-- Tab 2: 面试历史 -->
      <el-tab-pane label="📋 面试历史" name="history">
        <el-card shadow="never" v-loading="historyLoading">
          <el-table :data="interviewHistory" stripe empty-text="暂无面试记录">
            <el-table-column label="时间" width="160">
              <template slot-scope="{ row }">{{ formatTime(row.startTime || row.createdAt) }}</template>
            </el-table-column>
            <el-table-column label="岗位" prop="jobTitle" />
            <el-table-column label="轮次" width="80">
              <template slot-scope="{ row }">
                <el-tag size="mini" type="info">{{ roundLabel(row.round) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="得分" width="80">
              <template slot-scope="{ row }">
                <span :style="{ color: scoreColor(row.totalScore), fontWeight: 'bold' }">
                  {{ row.totalScore != null ? row.totalScore : '—' }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="90">
              <template slot-scope="{ row }">
                <el-tag size="mini" :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100" align="center">
              <template slot-scope="{ row }">
                <el-button
                  type="text"
                  size="mini"
                  @click="$router.push(`/report/${row.interviewId || row.id}`)"
                  v-if="row.status === 'completed'"
                >查看报告</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- Tab 3: 成长轨迹 -->
      <el-tab-pane label="📈 成长轨迹" name="trend">
        <el-card shadow="never" v-loading="trendLoading">
          <div slot="header" class="card-title">近30次面试得分趋势</div>
          <div ref="trendChart" class="trend-chart" />
          <div v-if="!trendLoading && trendEmpty" class="empty-tip">
            暂无趋势数据，完成更多面试后将在这里展示
          </div>
        </el-card>
      </el-tab-pane>

      <!-- Tab 4: 我的收藏 -->
      <el-tab-pane label="⭐ 我的收藏" name="collections">
        <el-card shadow="never" v-loading="collectionsLoading">
          <el-tabs v-model="collectTab">
            <el-tab-pane label="收藏题目" name="questions">
              <div v-if="collectedQuestions.length === 0" class="empty-tip">暂无收藏题目</div>
              <div
                v-for="q in collectedQuestions"
                :key="q.questionId || q.id"
                class="collect-item"
                @click="$router.push(`/questions/${q.questionId || q.id}/practice`)"
              >
                <div class="collect-title">{{ q.content || q.title }}</div>
                <div class="collect-meta">
                  <el-tag size="mini" type="primary" effect="plain">{{ q.jobTitle }}</el-tag>
                  <el-tag size="mini" :type="diffTagType(q.difficulty)">{{ diffLabel(q.difficulty) }}</el-tag>
                </div>
              </div>
            </el-tab-pane>
            <el-tab-pane label="收藏文章" name="articles">
              <div v-if="collectedArticles.length === 0" class="empty-tip">暂无收藏文章</div>
              <div
                v-for="a in collectedArticles"
                :key="a.articleId || a.id"
                class="collect-item"
                @click="$router.push(`/community/${a.articleId || a.id}`)"
              >
                <div class="collect-title">{{ a.title }}</div>
                <div class="collect-meta">
                  <el-tag v-for="tag in (a.tags || []).slice(0,2)" :key="tag" size="mini" type="info">{{ tag }}</el-tag>
                  <span class="collect-time">{{ formatTime(a.createdAt) }}</span>
                </div>
              </div>
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </el-tab-pane>

    </el-tabs>
  </div>
</template>

<script>
import { getProfile, updateProfile, changePassword, getScoreTrend, getCollections } from '@/api/profile'
import request from '@/api/request'
import * as echarts from 'echarts'

export default {
  name: 'ProfileIndex',
  data() {
    const validateConfirmPwd = (rule, value, callback) => {
      if (value !== this.pwdForm.newPassword) {
        callback(new Error('两次输入密码不一致'))
      } else {
        callback()
      }
    }
    return {
      activeTab: 'info',
      collectTab: 'questions',
      // 个人信息
      profileLoading: false,
      savingProfile: false,
      profileForm: { nickname: '', avatar: '', bio: '' },
      profileRules: {
        nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }]
      },
      // 密码
      savingPwd: false,
      pwdForm: { oldPassword: '', newPassword: '', confirmPassword: '' },
      pwdRules: {
        oldPassword: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
        newPassword: [
          { required: true, message: '请输入新密码', trigger: 'blur' },
          { min: 8, message: '密码不少于8位', trigger: 'blur' }
        ],
        confirmPassword: [
          { required: true, message: '请确认新密码', trigger: 'blur' },
          { validator: validateConfirmPwd, trigger: 'blur' }
        ]
      },
      // 面试历史
      historyLoading: false,
      interviewHistory: [],
      // 趋势图
      trendLoading: false,
      trendData: [],
      trendChart: null,
      // 收藏
      collectionsLoading: false,
      collectedQuestions: [],
      collectedArticles: []
    }
  },
  computed: {
    trendEmpty() {
      return !this.trendData || this.trendData.length === 0
    }
  },
  created() {
    this.loadProfile()
    this.loadHistory()
  },
  watch: {
    activeTab(val) {
      if (val === 'trend') {
        this.loadTrend()
      } else if (val === 'collections') {
        this.loadCollections()
      }
    }
  },
  beforeDestroy() {
    if (this.trendChart) {
      this.trendChart.dispose()
      this.trendChart = null
    }
  },
  methods: {
    async loadProfile() {
      this.profileLoading = true
      try {
        const res = await getProfile()
        const d = res.data || res
        this.profileForm.nickname = d.nickname || ''
        this.profileForm.avatar = d.avatar || ''
        this.profileForm.bio = d.bio || ''
      } catch (e) {
        this.$message.error('个人信息加载失败')
      } finally {
        this.profileLoading = false
      }
    },
    triggerFileInput() {
      this.$refs.fileInput.click()
    },
    onFileChange(e) {
      const file = e.target.files[0]
      if (!file) return
      if (file.size > 2 * 1024 * 1024) {
        this.$message.warning('图片大小不能超过 2MB')
        return
      }
      const reader = new FileReader()
      reader.onload = (ev) => {
        this.profileForm.avatar = ev.target.result  // base64
        this.$message.success('头像已选择，点击"保存修改"后生效')
      }
      reader.readAsDataURL(file)
      // 清空 input，允许重复选同一文件
      e.target.value = ''
    },
    async saveProfile() {
      this.$refs.profileForm.validate(async valid => {
        if (!valid) return
        this.savingProfile = true
        try {
          await updateProfile(this.profileForm)
          // 同步更新 Vuex userInfo，让 Navbar 头像实时刷新
          this.$store.commit('user/SET_USER_INFO', {
            ...this.$store.state.user.userInfo,
            nickname: this.profileForm.nickname,
            avatar: this.profileForm.avatar,
            bio: this.profileForm.bio
          })
          this.$message.success('保存成功')
        } catch (e) {
          this.$message.error('保存失败')
        } finally {
          this.savingProfile = false
        }
      })
    },
    async changePassword() {
      this.$refs.pwdForm.validate(async valid => {
        if (!valid) return
        this.savingPwd = true
        try {
          await changePassword({
            oldPassword: this.pwdForm.oldPassword,
            newPassword: this.pwdForm.newPassword
          })
          this.$message.success('密码修改成功')
          this.$refs.pwdForm.resetFields()
        } catch (e) {
          this.$message.error(e?.response?.data?.message || '密码修改失败')
        } finally {
          this.savingPwd = false
        }
      })
    },
    async loadHistory() {
      this.historyLoading = true
      try {
        const res = await request.get('/interviews', { params: { page: 1, pageSize: 100 } })
        const d = res.data || res
        this.interviewHistory = Array.isArray(d) ? d : (d.list || d.items || [])
      } catch (e) {
        // 静默失败
      } finally {
        this.historyLoading = false
      }
    },
    async loadTrend() {
      this.trendLoading = true
      try {
        const res = await getScoreTrend()
        const d = res.data || res
        this.trendData = Array.isArray(d) ? d : (d.list || d.items || [])
        this.$nextTick(() => { this.initTrendChart() })
      } catch (e) {
        // 静默失败
      } finally {
        this.trendLoading = false
      }
    },
    initTrendChart() {
      if (!this.$refs.trendChart || this.trendEmpty) return
      if (this.trendChart) this.trendChart.dispose()
      this.trendChart = echarts.init(this.$refs.trendChart)
      const dates = this.trendData.map(item => item.date || item.day || '')
      const scores = this.trendData.map(item => item.score ?? item.avgScore ?? 0)
      this.trendChart.setOption({
        tooltip: { trigger: 'axis' },
        grid: { left: 40, right: 20, top: 20, bottom: 40 },
        xAxis: { type: 'category', data: dates, axisLabel: { color: '#666' } },
        yAxis: { type: 'value', min: 0, max: 100, interval: 20 },
        series: [{
          type: 'line',
          data: scores,
          smooth: true,
          symbol: 'circle',
          symbolSize: 6,
          lineStyle: { color: '#ff9900', width: 2 },
          itemStyle: { color: '#ff9900' },
          areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [{ offset: 0, color: 'rgba(255,153,0,0.3)' }, { offset: 1, color: 'rgba(255,153,0,0.02)' }] } }
        }]
      })
    },
    async loadCollections() {
      this.collectionsLoading = true
      try {
        const res = await getCollections()
        const d = res.data || res
        this.collectedQuestions = d.questions || []
        this.collectedArticles = d.articles || []
      } catch (e) {
        // 静默失败
      } finally {
        this.collectionsLoading = false
      }
    },
    formatTime(val) {
      if (!val) return '—'
      const d = new Date(val)
      if (isNaN(d.getTime())) return val
      const pad = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
    },
    roundLabel(r) {
      const map = { round1: '一面', round2: '二面', round3: '三面' }
      return map[r] || r || '—'
    },
    scoreColor(s) {
      if (s == null) return '#999'
      if (s >= 80) return '#67c23a'
      if (s >= 60) return '#ff9900'
      return '#f56c6c'
    },
    statusTagType(s) {
      const map = { completed: 'success', ongoing: 'warning', paused: 'info' }
      return map[s] || 'info'
    },
    statusLabel(s) {
      const map = { completed: '已完成', ongoing: '进行中', paused: '已暂停' }
      return map[s] || s || '—'
    },
    diffTagType(d) {
      const map = { junior: 'success', middle: 'warning', senior: 'danger' }
      return map[d] || 'info'
    },
    diffLabel(d) {
      const map = { junior: '初级', middle: '中级', senior: '高级' }
      return map[d] || d || '—'
    }
  }
}
</script>

<style scoped>
.profile-page { max-width: 960px; margin: 0 auto; padding: 20px 16px; }
.card-title { font-size: 15px; font-weight: bold; color: #111; border-left: 4px solid #ff9900; padding-left: 10px; }

/* 头像上传 */
.avatar-upload-wrap { display: flex; align-items: center; gap: 16px; }
.avatar-preview {
  width: 80px; height: 80px; border-radius: 50%;
  border: 2px dashed #ddd; cursor: pointer;
  position: relative; overflow: hidden;
  transition: border-color 0.2s;
  flex-shrink: 0;
}
.avatar-preview:hover { border-color: #ff9900; }
.avatar-img { width: 100%; height: 100%; object-fit: cover; display: block; }
.avatar-placeholder {
  width: 100%; height: 100%;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  color: #bbb; font-size: 11px; gap: 4px;
}
.avatar-placeholder i { font-size: 22px; }
.avatar-overlay {
  position: absolute; inset: 0;
  background: rgba(0,0,0,0.45); color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-size: 12px; gap: 4px;
  opacity: 0; transition: opacity 0.2s;
}
.avatar-preview:hover .avatar-overlay { opacity: 1; }
.avatar-tip { font-size: 12px; color: #aaa; line-height: 1.6; }
.trend-chart { width: 100%; height: 280px; }
.empty-tip { text-align: center; color: #999; padding: 30px 0; font-size: 14px; }
.collect-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #f5f5f5;
  cursor: pointer;
  gap: 12px;
}
.collect-item:last-child { border-bottom: none; }
.collect-item:hover .collect-title { color: #ff9900; text-decoration: underline; }
.collect-title { font-size: 14px; font-weight: bold; color: #0066c0; flex: 1; line-height: 1.5; }
.collect-meta { display: flex; align-items: center; gap: 6px; flex-shrink: 0; flex-wrap: wrap; }
.collect-time { font-size: 12px; color: #bbb; }
</style>
