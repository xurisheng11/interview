<template>
  <div class="resume-detail-page">
    <!-- 顶部导航 -->
    <div class="page-nav">
      <el-button size="small" icon="el-icon-arrow-left" @click="$router.push('/resume')">返回列表</el-button>
    </div>

    <div v-loading="loading">
      <div v-if="!loading && !resume" class="not-found">简历不存在或无权访问</div>

      <div v-else-if="!loading && resume">
        <!-- 文件信息卡片 -->
        <el-card shadow="never" class="info-card">
          <div class="resume-header">
            <div class="resume-icon-lg">📄</div>
            <div class="resume-meta">
              <div class="resume-filename-lg">{{ resume.filename }}</div>
              <div class="resume-tags">
                <el-tag size="mini" type="info">{{ formatFileSize(resume.fileSize) }}</el-tag>
                <el-tag size="mini" type="info">{{ formatDate(resume.uploadedAt) }}</el-tag>
                <el-tag size="mini" :type="statusTagType(resume.analysisStatus)">
                  {{ statusLabel(resume.analysisStatus) }}
                </el-tag>
              </div>
            </div>
          </div>
        </el-card>

        <el-row :gutter="16">

          <!-- 左列：分析结果 -->
          <el-col :span="16">
            <!-- 分析结果 -->
            <el-card shadow="never" class="section-card" v-loading="analysisLoading">
              <div slot="header" class="card-header">
                <span class="card-title">📊 AI 简历分析</span>
                <el-button
                  v-if="resume.analysisStatus === 'failed'"
                  type="warning" size="mini"
                  :loading="retrying"
                  @click="handleRetry"
                >
                  重新分析
                </el-button>
              </div>

              <!-- 分析中 -->
              <div v-if="resume.analysisStatus === 'pending' || resume.analysisStatus === 'analyzing'" class="analyzing-box">
                <div class="analyzing-icon">⏳</div>
                <div class="analyzing-text">AI 正在分析简历，请稍候…</div>
              </div>

              <!-- 分析失败 -->
              <div v-else-if="resume.analysisStatus === 'failed'" class="failed-box">
                <div class="failed-icon">⚠️</div>
                <div class="failed-text">简历分析失败，请点击上方按钮重试</div>
              </div>

              <!-- 分析完成 -->
              <div v-else-if="resume.analysisStatus === 'done' && resume.analysis">
                <!-- 总分 + 雷达图 -->
                <el-row :gutter="20">
                  <el-col :span="10">
                    <div class="total-score-wrap">
                      <ScoreCircle :score="resume.analysis.totalScore" :size="140" />
                      <div class="total-score-label">简历综合评分</div>
                    </div>
                  </el-col>
                  <el-col :span="14">
                    <div class="radar-wrap">
                      <RadarChart :modules="radarData" />
                    </div>
                  </el-col>
                </el-row>

                <!-- 改进建议 -->
                <div class="suggestions-section">
                  <div class="suggestions-title">💡 改进建议</div>
                  <ul class="suggestions-list">
                    <li v-for="(s, i) in resume.analysis.suggestions" :key="i" class="suggestion-item">
                      {{ s }}
                    </li>
                  </ul>
                </div>
              </div>
            </el-card>

            <!-- 简历结构化内容 -->
            <el-card shadow="never" class="section-card" style="margin-top:16px">
              <div slot="header"><span class="card-title">📋 简历内容摘要</span></div>
              <div class="parsed-content">
                <!-- 目标岗位 -->
                <div v-if="resume.parsedContent.jobTitle" class="content-section">
                  <div class="content-label">🎯 目标岗位</div>
                  <div class="content-value">{{ resume.parsedContent.jobTitle }}</div>
                </div>

                <!-- 技能 -->
                <div v-if="resume.parsedContent.skills && resume.parsedContent.skills.length" class="content-section">
                  <div class="content-label">🛠️ 技能关键词</div>
                  <div class="skill-tags">
                    <el-tag
                      v-for="s in resume.parsedContent.skills" :key="s"
                      size="small" type="info" class="skill-tag"
                    >{{ s }}</el-tag>
                  </div>
                </div>

                <!-- 工作经历 -->
                <div v-if="resume.parsedContent.workExperience && resume.parsedContent.workExperience.length" class="content-section">
                  <div class="content-label">💼 工作经历</div>
                  <div v-for="(w, i) in resume.parsedContent.workExperience" :key="i" class="entry-item">
                    <div class="entry-title">{{ w.position }} @ {{ w.company }}</div>
                    <div class="entry-duration">{{ w.duration }}</div>
                    <div class="entry-desc">{{ w.desc }}</div>
                  </div>
                </div>

                <!-- 项目经历 -->
                <div v-if="resume.parsedContent.projects && resume.parsedContent.projects.length" class="content-section">
                  <div class="content-label">🚀 项目经历</div>
                  <div v-for="(p, i) in resume.parsedContent.projects" :key="i" class="entry-item">
                    <div class="entry-title">{{ p.name }} <span class="entry-role">({{ p.role }})</span></div>
                    <div class="entry-stack">技术栈：{{ p.stack }}</div>
                    <div class="entry-desc">{{ p.desc }}</div>
                  </div>
                </div>
              </div>
            </el-card>
          </el-col>

          <!-- 右列：操作 -->
          <el-col :span="8">
            <el-card shadow="never" class="action-card">
              <div slot="header"><span class="card-title">🚀 发起面试</span></div>
              <div class="action-tip">
                基于简历内容生成个性化面试题目，进行针对性练习。
              </div>
              <el-button
                type="primary"
                style="width:100%;margin-top:14px"
                :loading="startingInterview"
                :disabled="resume.analysisStatus !== 'done'"
                @click="handleStartInterview"
              >
                {{ resume.analysisStatus === 'done' ? '开始简历面试' : '分析完成后可发起' }}
              </el-button>
              <div v-if="resume.analysisStatus !== 'done'" class="action-hint">
                当前简历分析{{ statusLabel(resume.analysisStatus) }}，请等待分析完成后再发起面试。
              </div>
            </el-card>
          </el-col>

        </el-row>
      </div>
    </div>
  </div>
</template>

<script>
import { getResume, retryAnalyze, createResumeInterview } from '@/api/resume'
import ScoreCircle from '@/components/common/ScoreCircle.vue'
import RadarChart from '@/components/common/RadarChart.vue'

export default {
  name: 'ResumeDetail',
  components: { ScoreCircle, RadarChart },
  data() {
    return {
      loading: true,
      analysisLoading: false,
      resume: null,
      retrying: false,
      startingInterview: false
    }
  },
  computed: {
    radarData() {
      if (!this.resume?.analysis?.dimensions) return []
      return this.resume.analysis.dimensions.map(d => ({
        name: d.name,
        score: d.score
      }))
    }
  },
  created() {
    this.load()
  },
  methods: {
    async load() {
      this.loading = true
      try {
        const res = await getResume(this.$route.params.id)
        this.resume = res.data || res
        // 分析完成时轮询状态（万一还在分析中）
        if (this.resume.analysisStatus === 'pending' || this.resume.analysisStatus === 'analyzing') {
          this.pollAnalysis()
        }
      } catch (e) {
        // 错误已在拦截器处理
      }
      this.loading = false
    },
    pollAnalysis() {
      const interval = setInterval(async () => {
        try {
          const res = await getResume(this.$route.params.id)
          this.resume = res.data || res
          if (this.resume.analysisStatus === 'done' || this.resume.analysisStatus === 'failed') {
            clearInterval(interval)
          }
        } catch {}
      }, 5000)
      this.$once('hook:beforeDestroy', () => clearInterval(interval))
    },
    async handleRetry() {
      this.retrying = true
      try {
        await retryAnalyze(this.$route.params.id)
        this.resume.analysisStatus = 'pending'
        this.pollAnalysis()
        this.$message.success('重新分析已启动')
      } catch (e) {
        this.$message.error('重试失败')
      }
      this.retrying = false
    },
    async handleStartInterview() {
      if (this.resume.analysisStatus !== 'done') return
      this.startingInterview = true
      try {
        const res = await createResumeInterview(this.$route.params.id)
        const id = res.data?.interviewId || res.interviewId
        this.$message.success('面试会话已创建，正在跳转…')
        this.$router.push(`/interview/${id}/doing`)
      } catch (e) {
        this.$message.error(e.message || '发起面试失败')
      }
      this.startingInterview = false
    },
    formatDate(val) {
      if (!val) return '—'
      const d = new Date(val)
      const p = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${p(d.getMonth()+1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
    },
    formatFileSize(bytes) {
      if (!bytes) return '未知大小'
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      return (bytes / 1024 / 1024).toFixed(1) + ' MB'
    },
    statusTagType(s) {
      return { pending: 'info', analyzing: 'warning', done: 'success', failed: 'danger' }[s] || 'info'
    },
    statusLabel(s) {
      return { pending: '待分析', analyzing: '分析中', done: '已分析', failed: '分析失败' }[s] || s || '未知'
    }
  }
}
</script>

<style scoped>
.resume-detail-page { padding: 20px; max-width: 1100px; margin: 0 auto; }
.page-nav { margin-bottom: 14px; }
.not-found { text-align: center; padding: 60px; color: #999; font-size: 15px; }

/* 简历头部 */
.resume-header { display: flex; align-items: center; gap: 16px; }
.resume-icon-lg { font-size: 52px; flex-shrink: 0; }
.resume-filename-lg { font-size: 18px; font-weight: bold; color: #111; margin-bottom: 8px; }
.resume-tags { display: flex; flex-wrap: wrap; gap: 6px; }

/* 卡片 */
.info-card { margin-bottom: 16px; }
.section-card, .action-card { height: 100%; }
.card-header { display: flex; align-items: center; justify-content: space-between; }
.card-title { font-size: 15px; font-weight: bold; color: #111; border-left: 4px solid #ff9900; padding-left: 10px; }

/* 分析状态 */
.analyzing-box, .failed-box { text-align: center; padding: 40px 20px; }
.analyzing-icon, .failed-icon { font-size: 48px; margin-bottom: 12px; }
.analyzing-text { font-size: 14px; color: #888; }
.failed-text { font-size: 14px; color: #c7511f; }

/* 分数和雷达图 */
.total-score-wrap { text-align: center; }
.total-score-label { font-size: 13px; color: #888; margin-top: 8px; }
.radar-wrap { height: 220px; }

/* 建议 */
.suggestions-section { margin-top: 24px; border-top: 1px solid #f0f0f0; padding-top: 16px; }
.suggestions-title { font-size: 14px; font-weight: bold; color: #111; margin-bottom: 10px; }
.suggestions-list { list-style: none; padding: 0; margin: 0; }
.suggestion-item {
  padding: 8px 12px; font-size: 13px; color: #555;
  border-left: 3px solid #ff9900; background: #fffbf5; margin-bottom: 8px; border-radius: 0 4px 4px 0;
  line-height: 1.6;
}

/* 解析内容 */
.parsed-content {}
.content-section { margin-bottom: 20px; }
.content-label { font-size: 13px; font-weight: bold; color: #111; margin-bottom: 8px; }
.content-value { font-size: 14px; color: #555; }
.skill-tags { display: flex; flex-wrap: wrap; gap: 6px; }
.skill-tag { margin: 0; }
.entry-item {
  padding: 10px 14px; border: 1px solid #f0f0f0; border-radius: 8px; margin-bottom: 10px; background: #fafafa;
}
.entry-title { font-size: 14px; font-weight: 600; color: #111; margin-bottom: 4px; }
.entry-role { font-weight: normal; color: #888; font-size: 13px; }
.entry-duration, .entry-stack { font-size: 12px; color: #aaa; margin-bottom: 4px; }
.entry-desc { font-size: 13px; color: #666; line-height: 1.6; margin-top: 4px; }

/* 操作卡片 */
.action-tip { font-size: 13px; color: #888; line-height: 1.6; }
.action-hint { font-size: 12px; color: #bbb; margin-top: 10px; line-height: 1.5; }

::v-deep .el-card { border: 1px solid #eee; }
::v-deep .el-button--primary { background: #ff9900; border-color: #ff9900; color: #111; }
::v-deep .el-button--primary:hover { background: #f3a847; border-color: #f3a847; }
::v-deep .el-button--primary.is-disabled { background: #eee; border-color: #eee; color: #aaa; }
</style>
