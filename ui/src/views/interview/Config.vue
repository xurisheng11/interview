<template>
  <div class="layout">
    <!-- 侧边栏 -->
    <Sidebar :items="sidebarItems" />

    <!-- 主内容区 -->
    <div class="main-content">
      <!-- 练习模式选择 -->
      <div class="mode-tabs">
        <div
          class="mode-tab"
          :class="{ active: practiceMode === 'interview' }"
          @click="practiceMode = 'interview'"
        >
          <div class="mode-icon">🎯</div>
          <div class="mode-title">完整模拟面试</div>
          <div class="mode-desc">模拟真实面试流程,连贯答题</div>
        </div>
        <div
          class="mode-tab"
          :class="{ active: practiceMode === 'single' }"
          @click="practiceMode = 'single'"
        >
          <div class="mode-icon">✏️</div>
          <div class="mode-title">单题练习</div>
          <div class="mode-desc">逐个击破,打好基础</div>
        </div>
        <div
          class="mode-tab"
          :class="{ active: practiceMode === 'category' }"
          @click="practiceMode = 'category'"
        >
          <div class="mode-icon">📚</div>
          <div class="mode-title">分类练习</div>
          <div class="mode-desc">按知识点批量练习</div>
        </div>
      </div>

      <!-- 单题练习/分类练习: 跳转到题库 -->
      <div v-if="practiceMode === 'single' || practiceMode === 'category'" class="practice-hint">
        <el-card shadow="hover">
          <div class="hint-content">
            <div class="hint-icon">💡</div>
            <div class="hint-text">
              <template v-if="practiceMode === 'single'">
                <p><strong>单题练习</strong>适合面试小白从零开始,逐题积累经验。</p>
                <p>每道题都有 AI 点评,帮你了解答题要点和改进方向。</p>
              </template>
              <template v-else>
                <p><strong>分类练习</strong>让你按知识点系统性地提升。</p>
                <p>例如:先把"哈希表"相关的题都练一遍,再准备下一个知识点。</p>
              </template>
            </div>
            <el-button type="primary" @click="goToQuestionBank">
              前往题库 →
            </el-button>
          </div>
        </el-card>
      </div>

      <!-- 完整模拟面试配置 -->
      <div v-if="practiceMode === 'interview'" class="card">
        <div class="card-title">🎯 发起模拟面试</div>

        <!-- 目标公司（可选） -->
        <div class="form-group">
          <label>🏢 目标公司 <span class="optional">（可选，帮助匹配更精准的题目）</span></label>
          <el-autocomplete
            v-model="config.company"
            class="company-input"
            :fetch-suggestions="searchCompany"
            placeholder="输入目标公司名称，如：腾讯、阿里..."
            :trigger-on-focus="false"
            clearable
            @select="handleCompanySelect"
          >
            <template slot="suffix">
              <i class="el-icon-search"></i>
            </template>
          </el-autocomplete>
        </div>

        <div class="config-grid">
          <!-- 左列：目标岗位 + 面试难度 -->
          <div>
            <!-- 目标岗位 -->
            <div class="form-group">
              <label>💼 目标岗位 <span class="required">*</span></label>
              <el-select
                v-model="config.jobTitle"
                placeholder="选择或搜索岗位"
                filterable
                class="job-select"
              >
                <el-option-group label="技术类">
                  <el-option v-for="j in jobOptions.tech" :key="j" :label="j" :value="j" />
                </el-option-group>
                <el-option-group label="产品与设计">
                  <el-option v-for="j in jobOptions.product" :key="j" :label="j" :value="j" />
                </el-option-group>
                <el-option-group label="运营与市场">
                  <el-option v-for="j in jobOptions.operation" :key="j" :label="j" :value="j" />
                </el-option-group>
                <el-option-group label="职能类">
                  <el-option v-for="j in jobOptions.admin" :key="j" :label="j" :value="j" />
                </el-option-group>
              </el-select>
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
                >
                  <span class="tag-label">{{ opt.label }}</span>
                  <span class="tag-desc">{{ opt.desc }}</span>
                </div>
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

        <!-- 面试形式 -->
        <div class="form-group">
          <label>📋 面试形式 <span class="optional">（可多选）</span></label>
          <div class="tag-select interview-type">
            <div
              v-for="opt in interviewTypeOptions"
              :key="opt.value"
              class="tag type-tag"
              :class="{ selected: config.interviewTypes.includes(opt.value) }"
              @click="toggleInterviewType(opt.value)"
            >
              <span class="type-icon">{{ opt.icon }}</span>
              <span class="type-label">{{ opt.label }}</span>
            </div>
          </div>
          <div class="type-desc">
            <template v-if="config.interviewTypes.includes('structured')">
              <el-alert type="info" :closable="false" show-icon>
                <strong>结构化面试:</strong> 固定题目、统一评分标准,公平性强
              </el-alert>
            </template>
            <template v-else-if="config.interviewTypes.includes('semi-structured')">
              <el-alert type="info" :closable="false" show-icon>
                <strong>半结构化面试:</strong> 主干问题固定,追问灵活,最常见的面试形式
              </el-alert>
            </template>
            <template v-else-if="config.interviewTypes.includes('random')">
              <el-alert type="info" :closable="false" show-icon>
                <strong>随机问答:</strong> 不固定题目,考验临场反应和真实能力
              </el-alert>
            </template>
          </div>
        </div>

        <!-- 重点方向（多选，跨全宽） -->
        <div class="form-group">
          <label>🏷️ 重点方向 <span class="optional">（可多选）</span></label>
          <div class="tag-select">
            <div
              v-for="opt in focusOptions"
              :key="opt.value"
              class="tag"
              :class="{ selected: config.focusAreas.includes(opt.value) }"
              @click="toggleFocus(opt.value)"
            >{{ opt.label }}</div>
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

        <!-- 视频面试选项（仅视频模式显示） -->
        <div v-if="config.mode === 'video'" class="form-group video-options">
          <label>📹 视频面试选项</label>
          <div class="option-row">
            <div class="option-item">
              <span class="option-label">思考时间</span>
              <el-slider
                v-model="config.thinkTime"
                :min="0"
                :max="60"
                :step="5"
                :marks="thinkTimeMarks"
                class="option-slider"
              ></el-slider>
              <span class="option-value">{{ config.thinkTime }}秒</span>
            </div>
          </div>
          <div class="option-row">
            <div class="option-item">
              <span class="option-label">虚拟背景</span>
              <el-switch v-model="config.virtualBackground" active-text="开启" inactive-text="关闭"></el-switch>
            </div>
            <div class="option-item" v-if="config.virtualBackground">
              <el-select v-model="config.bgStyle" placeholder="选择背景" size="small">
                <el-option label="办公室" value="office" />
                <el-option label="纯色蓝" value="blue" />
                <el-option label="纯色灰" value="gray" />
                <el-option label="模糊背景" value="blur" />
              </el-select>
            </div>
          </div>
        </div>

        <!-- 补充说明（选填） -->
        <div class="form-group">
          <label>📝 补充说明 <span class="optional">（选填）</span></label>
          <textarea
            v-model="config.remark"
            class="remark-input"
            placeholder="描述你的项目经验、技术栈、目标公司的特殊要求等，帮助 AI 生成更精准的题目..."
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
import { createInterview, searchCompanies } from '@/api/interview'

export default {
  name: 'InterviewConfig',
  components: { Sidebar },

  data() {
    return {
      practiceMode: 'interview', // 'interview' | 'single' | 'category'
      loading: false,
      showHint: false,

      config: {
        company: '',
        companyId: '',
        jobTitle: '',
        difficulty: '',
        experience: '',
        round: '',
        interviewTypes: ['semi-structured'], // 默认为半结构化
        focusAreas: [],
        remark: '',
        mode: 'text',
        thinkTime: 15, // 思考时间(秒)
        virtualBackground: false,
        bgStyle: 'blur'
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

      // 岗位选项(按类别分组)
      jobOptions: {
        tech: [
          '后端开发', '前端开发', '全栈开发', '移动端开发(Android)',
          '移动端开发(iOS)', '大数据工程师', 'AI算法工程师', '测试工程师',
          '运维/DevOps', '网络安全', '嵌入式开发', '游戏开发',
          '游戏客户端开发', '游戏服务端开发', '数据分析', '数据工程',
          '机器学习工程师', '深度学习工程师', 'NLP工程师', '推荐算法工程师'
        ],
        product: [
          '产品经理', '产品助理', '高级产品经理', '数据产品经理',
          'AI产品经理', 'C端产品经理', 'B端产品经理', '平台产品经理',
          'UI设计师', 'UX设计师', '视觉设计师', '交互设计师',
          '平面设计师', '品牌设计师', '视频设计师'
        ],
        operation: [
          '运营专员', '内容运营', '用户运营', '活动运营',
          '新媒体运营', '电商运营', '社群运营', '游戏运营',
          '市场策划', '市场营销', '商务拓展', '销售代表',
          '客户经理', '渠道运营', '增长运营'
        ],
        admin: [
          '会计/财务', '人力资源', '行政管理', '法务/合规',
          '采购/供应链', '质量管理', '项目协调', ' CEO/总裁助理',
          '投资关系', '公关媒介'
        ]
      },

      difficultyOptions: [
        { label: '初级', value: 'junior', desc: '校招/入门级' },
        { label: '中级', value: 'middle', desc: '1-3年经验' },
        { label: '高级', value: 'senior', desc: '资深/专家级' }
      ],

      experienceOptions: [
        { label: '应届生', value: 'fresh' },
        { label: '1-3年', value: '1-3' },
        { label: '3-5年', value: '3-5' },
        { label: '5年以上', value: '5+' }
      ],

      roundOptions: [
        { label: '一面（基础）', value: 'round1' },
        { label: '二面（技术深度）', value: 'round2' },
        { label: '三面（综合/HR）', value: 'round3' }
      ],

      interviewTypeOptions: [
        { value: 'structured', label: '结构化面试', icon: '📋' },
        { value: 'semi-structured', label: '半结构化面试', icon: '🔄' },
        { value: 'random', label: '随机问答', icon: '🎲' }
      ],

      focusOptions: [
        { value: '算法', label: '算法与数据结构' },
        { value: 'system_design', label: '系统设计' },
        { value: 'project', label: '项目经验' },
        { value: 'basic', label: '基础知识' },
        { value: 'scenario', label: '场景题' },
        { value: 'behavior', label: '行为面试' }
      ],

      thinkTimeMarks: {
        0: '0秒',
        15: '15秒',
        30: '30秒',
        60: '60秒'
      },

      // 缓存搜索结果
      companyCache: []
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

    toggleInterviewType(type) {
      const idx = this.config.interviewTypes.indexOf(type)
      if (idx === -1) {
        this.config.interviewTypes.push(type)
      } else if (this.config.interviewTypes.length > 1) {
        // 至少保留一个
        this.config.interviewTypes.splice(idx, 1)
      }
    },

    async searchCompany(queryString, cb) {
      if (!queryString || queryString.length < 2) {
        cb([])
        return
      }
      try {
        const res = await searchCompanies(queryString)
        const companies = (res.data || res || []).map(c => ({
          value: c.name || c.companyName,
          id: c.id || c.companyId,
          questionCount: c.questionCount || 0
        }))
        this.companyCache = companies
        cb(companies)
      } catch (e) {
        cb([])
      }
    },

    handleCompanySelect(item) {
      this.config.companyId = item.id || ''
    },

    goToQuestionBank() {
      this.$router.push('/questions')
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

        // 添加新字段
        if (this.config.companyId) {
          payload.companyId = this.config.companyId
          payload.companyName = this.config.company
        }
        if (this.config.interviewTypes.length) {
          payload.interviewTypes = this.config.interviewTypes
        }
        if (this.config.mode === 'video') {
          payload.thinkTime = this.config.thinkTime
          payload.virtualBackground = this.config.virtualBackground
          payload.bgStyle = this.config.bgStyle
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
  width: 100%;
}

.main-content {
  flex: 1;
  padding: 20px;
  background: #fff;
  overflow: auto;
  min-width: 0;
}

/* ===== 练习模式选择 ===== */
.mode-tabs {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

.mode-tab {
  background: #fff;
  border: 2px solid #e8e8e8;
  border-radius: 12px;
  padding: 20px;
  text-align: center;
  cursor: pointer;
  transition: all 0.25s;
}

.mode-tab:hover {
  border-color: #ff9900;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(255, 153, 0, 0.15);
}

.mode-tab.active {
  border-color: #ff9900;
  background: linear-gradient(135deg, #fff9f0, #fff);
}

.mode-icon {
  font-size: 36px;
  margin-bottom: 8px;
}

.mode-title {
  font-size: 16px;
  font-weight: bold;
  color: #333;
  margin-bottom: 4px;
}

.mode-desc {
  font-size: 13px;
  color: #999;
}

/* ===== 练习提示卡片 ===== */
.practice-hint {
  max-width: 700px;
}

.hint-content {
  display: flex;
  align-items: center;
  gap: 20px;
}

.hint-icon {
  font-size: 40px;
  flex-shrink: 0;
}

.hint-text {
  flex: 1;
}

.hint-text p {
  margin: 4px 0;
  font-size: 14px;
  color: #666;
  line-height: 1.6;
}

/* ===== 卡片 ===== */
.card {
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 20px;
  max-width: 960px;
}

.card-title {
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 20px;
  color: #111;
  border-left: 4px solid #ff9900;
  padding-left: 10px;
}

.config-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0 24px;
}

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

/* 岗位选择器 */
.job-select {
  width: 100%;
}

::v-deep .el-select-group__title {
  font-weight: bold;
  color: #333;
}

/* 公司输入 */
.company-input {
  width: 100%;
  max-width: 400px;
}

.company-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.company-name {
  font-size: 14px;
}

.company-tag {
  font-size: 12px;
  color: #ff9900;
  background: #fff7e6;
  padding: 2px 6px;
  border-radius: 4px;
}

/* 标签选择区域 */
.tag-select {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag {
  padding: 8px 16px;
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

/* 难度标签带描述 */
.tag .tag-label {
  font-weight: bold;
}

.tag .tag-desc {
  font-size: 11px;
  color: #999;
  margin-left: 4px;
}

.tag.selected .tag-desc {
  color: #333;
}

/* 面试形式标签 */
.interview-type .type-tag {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px;
}

.type-icon {
  font-size: 16px;
}

.type-desc {
  margin-top: 10px;
}

.type-desc .el-alert {
  border-radius: 6px;
}

/* 视频面试选项 */
.video-options {
  background: #f9f9f9;
  border-radius: 8px;
  padding: 16px;
  margin-top: -8px;
}

.option-row {
  display: flex;
  align-items: center;
  gap: 24px;
  margin-bottom: 12px;
}

.option-row:last-child {
  margin-bottom: 0;
}

.option-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.option-label {
  font-size: 13px;
  color: #666;
  white-space: nowrap;
}

.option-slider {
  width: 200px;
}

.option-value {
  font-size: 14px;
  font-weight: bold;
  color: #ff9900;
  min-width: 45px;
}

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

.submit-row {
  text-align: center;
  margin-top: 4px;
}

.submit-btn {
  padding: 12px 60px !important;
  font-size: 16px !important;
  font-weight: bold !important;
  height: auto !important;
}

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

@media (max-width: 768px) {
  .config-grid {
    grid-template-columns: 1fr;
  }

  .mode-tabs {
    grid-template-columns: 1fr;
  }

  .submit-btn {
    padding: 12px 40px !important;
  }

  .option-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}
</style>
