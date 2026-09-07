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

        <!-- 简历关联（可选） -->
        <div class="form-group">
          <label>📄 关联简历 <span class="optional">（可选，AI 将围绕你的简历深挖提问）</span></label>
          <div class="resume-select-row">
            <el-select
              v-model="config.resumeId"
              placeholder="选择已上传的简历（让面试围绕你的经历展开）"
              clearable
              filterable
              class="resume-select"
              :loading="resumeLoading"
              @focus="loadUserResumes"
            >
              <el-option
                v-for="r in userResumes"
                :key="r.id"
                :label="r.filename"
                :value="r.id"
              >
                <div class="resume-option">
                  <span class="resume-name">{{ r.filename }}</span>
                  <el-tag v-if="r.analysisStatus === 'done'" type="success" size="mini">已分析</el-tag>
                  <el-tag v-else type="info" size="mini">{{ r.analysisStatus }}</el-tag>
                </div>
              </el-option>
            </el-select>
            <el-button size="small" @click="$router.push('/resumes')">
              上传新简历
            </el-button>
          </div>
          <div v-if="config.resumeId" class="resume-tip">
            <i class="el-icon-info"></i>
            关联简历后，系统将围绕你的项目经历和技术栈出题，面试更有针对性
          </div>
        </div>

        <!-- 求职类型 -->
        <div class="form-group">
          <label>🎯 求职类型 <span class="optional">（决定面试题目的侧重方向）</span></label>
          <div class="tag-select job-type-select">
            <div
              v-for="opt in jobTypeOptions"
              :key="opt.value"
              class="tag type-tag"
              :class="{ selected: config.jobType === opt.value }"
              @click="config.jobType = opt.value; onJobTypeChange(opt.value)"
            >
              <span class="type-icon">{{ opt.icon }}</span>
              <div class="type-info">
                <span class="type-label">{{ opt.label }}</span>
                <span class="type-desc">{{ opt.desc }}</span>
              </div>
            </div>
          </div>
          <!-- 求职类型说明 -->
          <div v-if="currentJobTypeInfo" class="jobtype-hint">
            <el-alert type="info" :closable="false" show-icon>
              <strong>{{ currentJobTypeInfo.label }}：</strong>{{ currentJobTypeInfo.hint }}
            </el-alert>
          </div>
        </div>

        <div class="config-grid">
          <!-- 左列：目标岗位 + 面试难度 -->
          <div>
            <!-- 目标岗位 - 左右两栏布局 -->
            <div class="form-group">
              <label>💼 目标岗位 <span class="required">*</span></label>
              <div class="job-selector-new">
                <!-- 左侧：大类列表 -->
                <div class="job-category-list">
                  <div
                    v-for="cat in jobCategories"
                    :key="cat.name"
                    class="job-category-item"
                    :class="{ active: selectedCategory === cat.name }"
                    @click="selectCategory(cat.name)"
                  >
                    <span class="category-icon">{{ cat.icon }}</span>
                    <span class="category-label">{{ cat.label }}</span>
                  </div>
                  <!-- 自定义选项 -->
                  <div
                    class="job-category-item custom-category"
                    :class="{ active: selectedCategory === 'custom' }"
                    @click="selectCategory('custom')"
                  >
                    <span class="category-icon">✏️</span>
                    <span class="category-label">自定义岗位</span>
                  </div>
                </div>

                <!-- 右侧：岗位列表 -->
                <div class="job-list-panel">
                  <!-- 搜索框 -->
                  <div class="job-search-box">
                    <el-input
                      v-model="jobSearchQuery"
                      placeholder="搜索岗位..."
                      size="small"
                      clearable
                      prefix-icon="el-icon-search"
                      @input="onJobSearch"
                    >
                    </el-input>
                  </div>

                  <!-- 搜索结果 -->
                  <template v-if="jobSearchQuery">
                    <div v-if="filteredJobResults.length > 0" class="job-items">
                      <div
                        v-for="job in filteredJobResults"
                        :key="job"
                        class="job-item"
                        :class="{ selected: config.jobTitle === job }"
                        @click="selectJob(job)"
                      >
                        {{ job }}
                      </div>
                    </div>
                    <div v-else class="job-empty">
                      <div class="custom-input-area">
                        <p>未找到"{{ jobSearchQuery }}"</p>
                        <el-button size="mini" type="primary" @click="selectJob(jobSearchQuery)">
                          <i class="el-icon-plus"></i> 使用此岗位
                        </el-button>
                      </div>
                    </div>
                  </template>

                  <!-- 当前选中分类的岗位列表 -->
                  <template v-else-if="selectedCategory !== 'custom'">
                    <div class="category-jobs">
                      <div
                        v-for="job in currentCategoryJobs"
                        :key="job"
                        class="job-item"
                        :class="{ selected: config.jobTitle === job }"
                        @click="selectJob(job)"
                      >
                        {{ job }}
                      </div>
                    </div>
                  </template>

                  <!-- 自定义输入模式 -->
                  <template v-else>
                    <div class="custom-input-area">
                      <p class="custom-tip">输入你想要的岗位名称，系统将围绕该岗位生成面试题目</p>
                      <el-input
                        v-model="customJobInput"
                        placeholder="例如：党建专员、内容编辑..."
                        size="small"
                        clearable
                      >
                        <template slot="append">
                          <el-button @click="applyCustomJob">确定</el-button>
                        </template>
                      </el-input>
                    </div>
                  </template>
                </div>
              </div>
              <!-- 已选岗位提示 -->
              <div v-if="config.jobTitle" class="selected-job-hint">
                <i class="el-icon-check"></i> 已选择：<strong>{{ config.jobTitle }}</strong>
                <span v-if="isCustomJob" class="custom-badge">自定义</span>
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
                >
                  <span class="tag-label">{{ opt.label }}</span>
                  <span class="tag-desc">{{ opt.desc }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 右列：工作经验 + 面试轮次 -->
          <div>
            <!-- 工作经验 - 只读显示，从个人中心读取 -->
            <div class="form-group">
              <label>💼 工作经验 <span class="optional">（来自个人资料）</span></label>
              <div v-if="profileExperience" class="experience-display">
                <div class="experience-value">
                  <i class="el-icon-user"></i>
                  {{ experienceLabel }}
                </div>
                <div class="experience-tip">
                  <i class="el-icon-info"></i>
                  如需修改，请前往
                  <el-button type="text" size="mini" @click="$router.push('/profile')">
                    个人中心 - 求职状态
                  </el-button>
                </div>
              </div>
              <div v-else class="experience-empty">
                <div class="experience-empty-icon">⚠️</div>
                <p>您还未设置工作经验</p>
                <el-button type="primary" size="small" @click="$router.push('/profile')">
                  去设置
                </el-button>
              </div>
            </div>

            <!-- 面试轮次 -->
            <div class="form-group">
              <label>🔄 面试轮次 <span class="required">*</span></label>
              <div class="tag-select round-select">
                <div
                  v-for="opt in roundOptions"
                  :key="opt.value"
                  class="tag round-tag"
                  :class="{ selected: config.round === opt.value }"
                  @click="config.round = opt.value"
                >
                  <div class="round-info">
                    <span class="round-label">{{ opt.label }}</span>
                    <span class="round-desc">{{ opt.desc }}</span>
                  </div>
                </div>
              </div>
              <!-- 选中轮次后的引导提示：覆盖范围 + 适合人群 + 风险提示 -->
              <el-alert
                v-if="selectedRoundOption"
                :type="config.round === 'round3' ? 'warning' : 'info'"
                :closable="false"
                show-icon
                class="round-guide-alert"
              >
                <strong>覆盖范围：</strong>{{ selectedRoundOption.coverage }}<br>
                <strong>适合人群：</strong>{{ selectedRoundOption.audience }}<br>
                <strong>{{ config.round === 'round3' ? '⚠️ 风险提示' : '注意' }}：</strong>{{ selectedRoundOption.risk }}
              </el-alert>
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
              <span class="option-label">最长思考上限</span>
              <el-select v-model="config.thinkTime" size="small" class="option-select">
                <el-option label="60秒（快节奏练习）" :value="60" />
                <el-option label="2分钟（推荐，接近真实面试）" :value="120" />
                <el-option label="3分钟（充分思考）" :value="180" />
                <el-option label="不限制" :value="0" />
              </el-select>
            </div>
          </div>
          <div class="think-time-explain">
            <el-alert type="warning" :closable="false" show-icon>
              <strong>思考时间说明：</strong>系统会自动记录你每道题的思考时间（从题目展示到首次开口/输入）。超过设定时长会温和提醒你开始作答，不会强制中断。
            </el-alert>
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
          请完成所有必填项（岗位、难度、轮次）后再发起面试
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
import { getProfile } from '@/api/profile'

export default {
  name: 'InterviewConfig',
  components: { Sidebar },

  data() {
    return {
      practiceMode: 'interview', // 'interview' | 'single' | 'category'
      loading: false,
      showHint: false,
      profileExperience: '', // 从用户资料读取的工作经验
      isCustomJob: false, // 是否选择了自定义岗位

      config: {
        company: '',
        companyId: '',
        jobTitle: '',
        jobType: '',       // 求职类型
        difficulty: '',
        experience: '',
        round: '',
        interviewTypes: ['semi-structured'], // 默认为半结构化
        focusAreas: [],
        remark: '',
        mode: 'text',
        thinkTime: 120, // 最长思考上限（秒），0 表示不限制
        virtualBackground: false,
        bgStyle: 'blur',
        resumeId: ''  // 关联简历ID
      },

      // 用户简历列表
      userResumes: [],
      resumeLoading: false,

      // 岗位选择器状态
      selectedCategory: 'tech', // 默认选中的大类
      jobSearchQuery: '',
      customJobInput: '',

      jobTypeOptions: [
        { label: '校园招聘', value: 'campus', icon: '🎓', desc: '校招/应届', hint: '重点考察基础知识、算法、逻辑思维，题目相对基础但覆盖面广' },
        { label: '社会招聘', value: 'social', icon: '💼', desc: '社招/跳槽', hint: '围绕项目经验深挖，考察技术深度、架构思维、解决问题能力' },
        { label: '事业单位', value: 'institution', icon: '🏛️', desc: '公务员/国企', hint: '结构化面试为主，重点考综合分析、计划组织、应急应变、人际沟通。包含漫画题、情景题等特殊题型。' },
        { label: '实习', value: 'intern', icon: '🌱', desc: '日常实习', hint: '题目最友好，考察学习能力、基础认知和实习动机' }
      ],

      sidebarItems: [
        {
          title: '面试中心',
          children: [
            { icon: '🚀', label: '发起面试', path: '/interview/config' },
            { icon: '📁', label: '我的记录', path: '/interview/history' }
          ]
        }
      ],

      // 岗位选项(按类别分组) - 重新组织为数组格式
      jobCategories: [
        {
          name: 'tech',
          label: '技术类',
          icon: '💻',
          jobs: [
            '后端开发', '前端开发', '全栈开发', '移动端开发(Android)',
            '移动端开发(iOS)', '大数据工程师', 'AI算法工程师', '测试工程师',
            '运维/DevOps', '网络安全', '嵌入式开发', '游戏开发',
            '游戏客户端开发', '游戏服务端开发', '数据分析', '数据工程',
            '机器学习工程师', '深度学习工程师', 'NLP工程师', '推荐算法工程师'
          ]
        },
        {
          name: 'product',
          label: '产品与设计',
          icon: '🎨',
          jobs: [
            '产品经理', '产品助理', '高级产品经理', '数据产品经理',
            'AI产品经理', 'C端产品经理', 'B端产品经理', '平台产品经理',
            'UI设计师', 'UX设计师', '视觉设计师', '交互设计师',
            '平面设计师', '品牌设计师', '视频设计师'
          ]
        },
        {
          name: 'operation',
          label: '运营与市场',
          icon: '📈',
          jobs: [
            '运营专员', '内容运营', '用户运营', '活动运营',
            '新媒体运营', '电商运营', '社群运营', '游戏运营',
            '市场策划', '市场营销', '商务拓展', '销售代表',
            '客户经理', '渠道运营', '增长运营'
          ]
        },
        {
          name: 'admin',
          label: '职能类',
          icon: '📋',
          jobs: [
            '会计/财务', '人力资源', '行政管理', '法务/合规',
            '采购/供应链', '质量管理', '项目协调', '总裁助理',
            '投资关系', '公关媒介', '党建专员', '行政前台'
          ]
        }
      ],

      difficultyOptions: [
        { label: '🟢 初级', value: 'easy', desc: '校招/入门级' },
        { label: '🟡 中级', value: 'medium', desc: '1-3年经验' },
        { label: '🔴 高级', value: 'hard', desc: '资深/专家级' }
      ],

      roundOptions: [
        {
          label: '一面',
          value: 'round1',
          desc: '基础能力考察',
          coverage: '自我介绍 + 基础知识 + 简单算法 + 项目概述',
          audience: '正在准备第一轮面试，或想夯实基础',
          risk: '不含深度技术题和系统设计'
        },
        {
          label: '二面',
          value: 'round2',
          desc: '技术深度考察',
          coverage: '技术深挖 + 系统设计 + 项目追问',
          audience: '已通过一面，准备技术深面',
          risk: '不含基础题和HR类问题'
        },
        {
          label: '三面',
          value: 'round3',
          desc: '综合能力/HR面',
          coverage: '综合素质 + 职业规划 + HR问答',
          audience: '准备终面/HR面',
          risk: '不含基础和技术深度题。如果目标面试只有一轮，建议选择「综合面试」'
        },
        {
          label: '综合面试',
          value: 'comprehensive',
          desc: '全链路覆盖，一轮打尽',
          coverage: '自我介绍 + 基础 + 技术深度 + 项目经验 + 综合/HR 均衡分布',
          audience: '目标单位只有一轮面试，或想全面练习',
          risk: '题量有限，各方向覆盖不如单轮深入'
        }
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

      // 缓存搜索结果
      companyCache: []
    }
  },

  computed: {
    // 当前选中的轮次选项（用于展示引导提示）
    selectedRoundOption() {
      if (!this.config.round) return null
      return this.roundOptions.find(opt => opt.value === this.config.round) || null
    },

    currentJobTypeInfo() {
      if (!this.config.jobType) return null
      return this.jobTypeOptions.find(opt => opt.value === this.config.jobType) || null
    },

    // 当前选中分类的岗位列表
    currentCategoryJobs() {
      const cat = this.jobCategories.find(c => c.name === this.selectedCategory)
      return cat ? cat.jobs : []
    },

    // 过滤后的搜索结果
    filteredJobResults() {
      if (!this.jobSearchQuery) return []
      const query = this.jobSearchQuery.toLowerCase()
      const allJobs = this.jobCategories.flatMap(c => c.jobs)
      return [...new Set(allJobs.filter(job => job.toLowerCase().includes(query)))]
    },

    // 工作经验标签显示
    experienceLabel() {
      const map = {
        fresh: '应届生',
        '1-3': '1-3年经验',
        '3-5': '3-5年经验',
        '5+': '5年以上'
      }
      return map[this.profileExperience] || this.profileExperience || '未设置'
    },

    isFormValid() {
      return (
        !!this.config.jobTitle &&
        !!this.config.difficulty &&
        !!this.profileExperience && // 现在依赖个人中心设置
        !!this.config.round
      )
    }
  },

  mounted() {
    // 从用户资料加载默认工作经验
    this.loadUserExperience()
  },

  methods: {
    onJobTypeChange(value) {
      if (value === 'intern' && !this.config.difficulty) {
        this.config.difficulty = 'easy'
      }
    },

    // 选择大类
    selectCategory(name) {
      this.selectedCategory = name
      this.jobSearchQuery = ''
      this.customJobInput = ''
    },

    // 选择岗位
    selectJob(job) {
      this.config.jobTitle = job
      this.isCustomJob = false
      this.jobSearchQuery = ''
    },

    // 应用自定义岗位
    applyCustomJob() {
      if (this.customJobInput.trim()) {
        this.config.jobTitle = this.customJobInput.trim()
        this.isCustomJob = true
        this.customJobInput = ''
      }
    },

    // 搜索岗位
    onJobSearch() {
      // 搜索时会自动显示搜索结果
    },

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

    async loadUserResumes() {
      if (this.userResumes.length > 0) return
      this.resumeLoading = true
      try {
        const { getResumeList } = await import('@/api/resume')
        const res = await getResumeList()
        const list = res.data || res || []
        this.userResumes = list.filter(r => r.analysisStatus === 'done')
      } catch (e) {
        this.userResumes = []
      } finally {
        this.resumeLoading = false
      }
    },

    goToQuestionBank() {
      this.$router.push('/questions')
    },

    async handleStart() {
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
          experience: this.profileExperience, // 使用个人中心设置的Experience
          round: this.config.round,
          focusAreas: [...this.config.focusAreas],
          remark: this.config.remark || '',
          mode: this.config.mode || 'text'
        }

        if (this.config.companyId) {
          payload.companyId = this.config.companyId
          payload.companyName = this.config.company
        }
        if (this.config.interviewTypes.length) {
          payload.interviewTypes = this.config.interviewTypes
        }
        if (this.config.jobType) {
          payload.jobType = this.config.jobType
        }
        if (this.config.mode === 'video') {
          payload.thinkTime = this.config.thinkTime
          payload.virtualBackground = this.config.virtualBackground
          payload.bgStyle = this.config.bgStyle
        }
        if (this.config.resumeId) {
          payload.resumeId = this.config.resumeId
        }

        const res = await createInterview(payload)

        const data = res.data || res
        const interviewId = data.id || data.interviewId || data.data?.id || data.data?.interviewId

        if (!interviewId) {
          throw new Error('未获取到面试 ID，请重试')
        }

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
    },

    async loadUserExperience() {
      try {
        const res = await getProfile()
        const d = res.data || res
        if (d.experience) {
          this.profileExperience = d.experience
          this.config.experience = d.experience
        }
      } catch (e) {
        // 静默失败
      }
    }
  }
}
</script>

<style scoped>
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

/* ===== 岗位选择器 - 新布局 ===== */
.job-selector-new {
  display: flex;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  overflow: hidden;
  min-height: 200px;
}

/* 左侧大类列表 */
.job-category-list {
  width: 120px;
  background: #f5f7fa;
  border-right: 1px solid #e4e7ed;
  flex-shrink: 0;
}

.job-category-item {
  padding: 12px 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s;
  border-bottom: 1px solid #e4e7ed;
  font-size: 13px;
  color: #606266;
}

.job-category-item:hover {
  background: #ecf5ff;
  color: #409eff;
}

.job-category-item.active {
  background: #fff;
  color: #ff9900;
  font-weight: bold;
  border-left: 3px solid #ff9900;
}

.job-category-item.custom-category {
  color: #909399;
  border-top: 1px dashed #dcdfe6;
  margin-top: 8px;
}

.job-category-item.custom-category.active {
  color: #ff9900;
}

.category-icon {
  font-size: 16px;
}

.category-label {
  white-space: nowrap;
}

/* 右侧岗位列表 */
.job-list-panel {
  flex: 1;
  padding: 12px;
  overflow-y: auto;
  max-height: 220px;
}

.job-search-box {
  margin-bottom: 10px;
}

.category-jobs,
.job-items {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.job-item {
  padding: 6px 14px;
  background: #f4f4f5;
  border-radius: 16px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
  color: #606266;
}

.job-item:hover {
  background: #fdf6ec;
  color: #ff9900;
  border-color: #ff9900;
}

.job-item.selected {
  background: #ff9900;
  color: #111;
  font-weight: bold;
}

.job-empty {
  text-align: center;
  padding: 20px;
  color: #909399;
}

.custom-input-area {
  text-align: center;
  padding: 10px;
}

.custom-tip {
  font-size: 12px;
  color: #909399;
  margin-bottom: 10px;
}

/* 已选岗位提示 */
.selected-job-hint {
  margin-top: 10px;
  font-size: 13px;
  color: #67c23a;
  display: flex;
  align-items: center;
  gap: 6px;
}

.custom-badge {
  background: #e6a23c;
  color: #fff;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
}

/* ===== 工作经验显示 ===== */
.experience-display {
  background: #f0f9eb;
  border: 1px solid #c2e7b0;
  border-radius: 8px;
  padding: 12px 16px;
}

.experience-value {
  font-size: 15px;
  font-weight: bold;
  color: #67c23a;
  display: flex;
  align-items: center;
  gap: 8px;
}

.experience-tip {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
  display: flex;
  align-items: center;
  gap: 4px;
}

.experience-empty {
  background: #fdf6ec;
  border: 1px dashed #f5dab1;
  border-radius: 8px;
  padding: 16px;
  text-align: center;
}

.experience-empty-icon {
  font-size: 24px;
  margin-bottom: 8px;
}

.experience-empty p {
  margin: 0 0 10px 0;
  font-size: 13px;
  color: #e6a23c;
}

/* 面试轮次 */
.round-select {
  flex-direction: column;
  align-items: flex-start;
}

.round-tag {
  width: 100%;
  padding: 10px 16px;
  border-radius: 8px;
}

.round-info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.round-label {
  font-weight: bold;
  font-size: 14px;
}

.round-desc {
  font-size: 11px;
  color: #999;
  margin-top: 2px;
}

.round-tag.selected .round-desc {
  color: #333;
}

/* ===== 通用标签样式 ===== */
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

/* 求职类型 */
.job-type-select .type-tag {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 16px;
  border-radius: 10px;
  flex: 1;
  min-width: 200px;
}

.jobtype-hint {
  margin-top: 10px;
}

/* 公司输入 */
.company-input {
  width: 100%;
  max-width: 400px;
}

/* 简历关联 */
.resume-select-row {
  display: flex;
  gap: 10px;
  align-items: center;
}

.resume-select {
  flex: 1;
  max-width: 400px;
}

.resume-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.resume-name {
  font-size: 14px;
}

.resume-tip {
  margin-top: 6px;
  font-size: 12px;
  color: #67c23a;
  background: #f0f9eb;
  border: 1px solid #c2e7b0;
  border-radius: 4px;
  padding: 5px 10px;
  display: flex;
  align-items: center;
  gap: 6px;
}

/* 视频面试选项 */
.video-options {
  background: #f9f9f9;
  border-radius: 8px;
  padding: 16px;
  margin-top: -8px;
}

.think-time-explain {
  margin-bottom: 16px;
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

.option-select {
  width: 220px;
}

.round-guide-alert {
  margin-top: 8px;
  width: 100%;
  line-height: 1.7;
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

/* 面试形式 */
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

  .job-selector-new {
    flex-direction: column;
  }

  .job-category-list {
    width: 100%;
    display: flex;
    flex-wrap: wrap;
    border-right: none;
    border-bottom: 1px solid #e4e7ed;
  }

  .job-category-item {
    border-bottom: none;
    border-right: 1px solid #e4e7ed;
    padding: 8px 12px;
  }

  .job-category-item.active {
    border-left: none;
    border-bottom: 2px solid #ff9900;
  }
}
</style>
