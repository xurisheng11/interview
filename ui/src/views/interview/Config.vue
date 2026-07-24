<template>
  <div class="layout">
    <!-- 侧边栏 -->
    <Sidebar :items="sidebarItems" />

    <!-- 主内容区 -->
    <div class="main-content">
      <div class="card">
        <div class="card-title">🚀 发起模拟面试</div>

        <div class="config-grid">
          <!-- 左列：目标岗位 + 面试难度 -->
          <div>
            <!-- 目标岗位 -->
            <div class="form-group">
              <label>🏢 目标岗位 <span class="required">*</span></label>
              <div class="tag-select">
                <div
                  v-for="opt in jobTitleOptions"
                  :key="opt.value"
                  class="tag"
                  :class="{ selected: config.jobTitle === opt.value }"
                  @click="config.jobTitle = opt.value"
                >{{ opt.label }}</div>
              </div>
            </div>

            <!-- 面试难度 -->
            <div class="form-group">
              <label>📊 面试难度 <span class="required">*</span></label>
              <div class="tag-select">
                <div
                  v-for="opt in difficultyOptions"
                  :key="opt.value"
                  class="tag"
                  :class="{ selected: config.difficulty === opt.value }"
                  @click="config.difficulty = opt.value"
                >{{ opt.label }}</div>
              </div>
            </div>
          </div>

          <!-- 右列：工作经验 + 面试轮次 -->
          <div>
            <!-- 工作经验 -->
            <div class="form-group">
              <label>💼 工作经验 <span class="required">*</span></label>
              <div class="tag-select">
                <div
                  v-for="opt in experienceOptions"
                  :key="opt.value"
                  class="tag"
                  :class="{ selected: config.experience === opt.value }"
                  @click="config.experience = opt.value"
                >{{ opt.label }}</div>
              </div>
            </div>

            <!-- 面试轮次 -->
            <div class="form-group">
              <label>🔄 面试轮次 <span class="required">*</span></label>
              <div class="tag-select">
                <div
                  v-for="opt in roundOptions"
                  :key="opt.value"
                  class="tag"
                  :class="{ selected: config.round === opt.value }"
                  @click="config.round = opt.value"
                >{{ opt.label }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- 重点方向（多选，跨全宽） -->
        <div class="form-group">
          <label>🏷️ 重点方向 <span class="optional">（可多选）</span></label>
          <div class="tag-select">
            <div
              v-for="opt in focusOptions"
              :key="opt"
              class="tag"
              :class="{ selected: config.focusAreas.includes(opt) }"
              @click="toggleFocus(opt)"
            >{{ opt }}</div>
          </div>
        </div>

        <!-- 答题方式 -->
        <div class="form-group">
          <label>🎥 答题方式</label>
          <div class="tag-select">
            <div
              class="tag"
              :class="{ selected: config.mode === 'text' }"
              @click="config.mode = 'text'"
            >📝 文字输入</div>
            <div
              class="tag"
              :class="{ selected: config.mode === 'video' }"
              @click="config.mode = 'video'"
            >📹 视频面试</div>
          </div>
          <div v-if="config.mode === 'video'" class="video-hint">
            <i class="el-icon-video-camera"></i>
            需要摄像头和麦克风权限，建议使用 Chrome 浏览器
          </div>
        </div>

        <!-- 补充说明（选填） -->
        <div class="form-group">
          <label>📝 补充说明 <span class="optional">（选填）</span></label>
          <textarea
            v-model="config.remark"
            class="remark-input"
            placeholder="描述你的项目经验、技术栈等，帮助 AI 生成更精准的题目..."
          ></textarea>
        </div>

        <!-- 必填项提示 -->
        <div v-if="showHint" class="hint-bar">
          <i class="el-icon-warning-outline"></i>
          请完成所有必填项（岗位、难度、经验、轮次）后再发起面试
        </div>

        <!-- 提交按钮 -->
        <div class="submit-row">
          <el-button
            type="primary"
            class="submit-btn"
            :disabled="!isFormValid"
            :loading="loading"
            @click="handleStart"
          >🚀 发起面试</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import Sidebar from '@/components/layout/Sidebar.vue'
import { createInterview } from '@/api/interview'

export default {
  name: 'InterviewConfig',
  components: { Sidebar },

  data() {
    return {
      loading: false,
      showHint: false,

      config: {
        jobTitle: '',
        difficulty: '',
        experience: '',
        round: '',
        focusAreas: [],
        remark: '',
        mode: 'text'
      },

      sidebarItems: [
        {
          title: '面试中心',
          children: [
            { icon: '🚀', label: '发起面试', path: '/interview/config' },
            { icon: '📁', label: '我的记录', path: '/interview/history' }
          ]
        }
      ],

      jobTitleOptions: [
        { label: '后端开发', value: '后端开发' },
        { label: '前端开发', value: '前端开发' },
        { label: '全栈开发', value: '全栈开发' },
        { label: '移动端开发', value: '移动端开发' },
        { label: '大数据工程师', value: '大数据工程师' },
        { label: 'AI算法工程师', value: 'AI算法工程师' },
        { label: '测试工程师', value: '测试工程师' },
        { label: '运维/DevOps', value: '运维/DevOps' },
        { label: '网络安全', value: '网络安全' },
        { label: '嵌入式开发', value: '嵌入式开发' },
        { label: '产品经理', value: '产品经理' },
        { label: 'UI/UX设计师', value: 'UI/UX设计师' },
        { label: '平面设计师', value: '平面设计师' },
        { label: '数据分析师', value: '数据分析师' },
        { label: '会计/财务', value: '会计/财务' },
        { label: '市场营销', value: '市场营销' },
        { label: '运营专员', value: '运营专员' },
        { label: '新媒体运营', value: '新媒体运营' },
        { label: '人力资源', value: '人力资源' },
        { label: '行政管理', value: '行政管理' },
        { label: '销售/商务', value: '销售/商务' },
        { label: '项目管理', value: '项目管理' },
        { label: '法务/合规', value: '法务/合规' },
        { label: '客户服务', value: '客户服务' }
      ],

      difficultyOptions: [
        { label: '初级', value: '初级' },
        { label: '中级', value: '中级' },
        { label: '高级', value: '高级' }
      ],

      experienceOptions: [
        { label: '应届生', value: '应届生' },
        { label: '1-3年', value: '1-3年' },
        { label: '3-5年', value: '3-5年' },
        { label: '5年以上', value: '5年以上' }
      ],

      roundOptions: [
        { label: '一面（基础）', value: 'round1' },
        { label: '二面（技术深度）', value: 'round2' },
        { label: '三面（综合/HR）', value: 'round3' }
      ],

      focusOptions: ['算法', '系统设计', '项目经验', '基础知识', '场景题']
    }
  },

  computed: {
    isFormValid() {
      return (
        !!this.config.jobTitle &&
        !!this.config.difficulty &&
        !!this.config.experience &&
        !!this.config.round
      )
    }
  },

  methods: {
    toggleFocus(opt) {
      const idx = this.config.focusAreas.indexOf(opt)
      if (idx === -1) {
        this.config.focusAreas.push(opt)
      } else {
        this.config.focusAreas.splice(idx, 1)
      }
    },

    async handleStart() {
      // 如果必填项未完成，显示提示并阻止
      if (!this.isFormValid) {
        this.showHint = true
        return
      }
      if (this.loading) return

      this.showHint = false
      this.loading = true

      try {
        const payload = {
          jobTitle: this.config.jobTitle,
          difficulty: this.config.difficulty,
          experience: this.config.experience,
          round: this.config.round,
          focusAreas: [...this.config.focusAreas],
          remark: this.config.remark || '',
          mode: this.config.mode || 'text'
        }

        const res = await createInterview(payload)

        // 兼容不同后端响应结构
        const data = res.data || res
        const interviewId = data.id || data.interviewId || data.data?.id || data.data?.interviewId

        if (!interviewId) {
          throw new Error('未获取到面试 ID，请重试')
        }

        // 存储面试信息到 Vuex store
        this.$store.commit('interview/SET_CURRENT_ID', interviewId)
        if (data.interview) {
          this.$store.commit('interview/SET_INTERVIEW', data.interview)
        }
        if (data.questions && data.questions.length) {
          this.$store.commit('interview/SET_QUESTIONS', data.questions)
        }
        this.$store.commit('interview/SET_INTERVIEW_MODE', this.config.mode || 'text')

        this.$message.success('面试配置成功，AI 正在生成题目...')
        if (this.config.mode === 'video') {
          this.$router.push(`/interview/${interviewId}/video-prep`)
        } else {
          this.$router.push('/interview/loading')
        }
      } catch (err) {
        const msg =
          err?.response?.data?.message ||
          err?.response?.data?.msg ||
          err?.message ||
          '发起面试失败，请稍后重试'
        this.$message.error(msg)
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style scoped>
/* 整体布局继承 App.vue 中的 layout flex 容器 */
.layout {
  display: flex;
  min-height: calc(100vh - 90px);
}

.main-content {
  flex: 1;
  padding: 20px;
  background: #f3f3f3;
  overflow: auto;
}

/* 卡片 */
.card {
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 20px;
  max-width: 960px;
}

/* 卡片标题：左侧橙色竖线 */
.card-title {
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 20px;
  color: #111;
  border-left: 4px solid #ff9900;
  padding-left: 10px;
}

/* 两列网格 */
.config-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0 24px;
}

/* 表单组 */
.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  font-size: 13px;
  font-weight: bold;
  margin-bottom: 8px;
  color: #333;
}

.required {
  color: #f56c6c;
  margin-left: 2px;
}

.optional {
  font-size: 12px;
  font-weight: normal;
  color: #999;
  margin-left: 4px;
}

/* 标签选择区域 */
.tag-select {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

/* 单个标签 */
.tag {
  padding: 6px 14px;
  border-radius: 20px;
  border: 1px solid #ddd;
  background: #fff;
  cursor: pointer;
  font-size: 13px;
  color: #555;
  transition: all 0.2s;
  user-select: none;
}

.tag:hover {
  border-color: #ff9900;
  color: #ff9900;
}

.tag.selected {
  background: #ff9900;
  border-color: #ff9900;
  color: #111;
  font-weight: bold;
}

/* 补充说明文本域 */
.remark-input {
  width: 100%;
  padding: 10px;
  border: 1px solid #aaa;
  border-radius: 3px;
  font-size: 14px;
  min-height: 80px;
  resize: vertical;
  outline: none;
  transition: border 0.2s;
  font-family: inherit;
  box-sizing: border-box;
}

.remark-input:focus {
  border-color: #ff9900;
  box-shadow: 0 0 0 2px rgba(255, 153, 0, 0.15);
}

/* 必填项提示条 */
.hint-bar {
  background: #fff9f0;
  border: 1px solid #febd69;
  border-radius: 4px;
  padding: 10px 14px;
  font-size: 13px;
  color: #a05c00;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 6px;
}

/* 提交行 */
.submit-row {
  text-align: center;
  margin-top: 4px;
}

/* 发起面试按钮 */
.submit-btn {
  padding: 12px 60px !important;
  font-size: 16px !important;
  font-weight: bold !important;
  height: auto !important;
}

/* 覆盖 Element UI 主色为橙色 */
::v-deep .el-button--primary {
  background: #ff9900;
  border-color: #ff9900;
  color: #111;
}

::v-deep .el-button--primary:hover,
::v-deep .el-button--primary:focus {
  background: #f3a847;
  border-color: #f3a847;
  color: #111;
}

::v-deep .el-button--primary.is-disabled,
::v-deep .el-button--primary.is-disabled:hover {
  background: #d3d3d3;
  border-color: #d3d3d3;
  color: #888;
  cursor: not-allowed;
}

/* 视频面试提示 */
.video-hint {
  margin-top: 8px;
  font-size: 12px;
  color: #a05c00;
  background: #fff9f0;
  border: 1px solid #febd69;
  border-radius: 4px;
  padding: 6px 10px;
  display: flex;
  align-items: center;
  gap: 6px;
}

/* 响应式：小屏切单列 */
@media (max-width: 768px) {
  .config-grid {
    grid-template-columns: 1fr;
  }

  .submit-btn {
    padding: 12px 40px !important;
  }
}
</style>
