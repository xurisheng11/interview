<template>
  <div class="company-intel-page">
    <!-- 搜索区 -->
    <el-card shadow="never" class="search-card">
      <div class="search-header">
        <div class="search-title">🏢 公司面试知识库</div>
        <div class="search-desc">输入公司名称，AI 为你生成该公司的面试流程、常见题目和备考技巧</div>
      </div>
      <div class="search-form">
        <el-input
          v-model="company"
          placeholder="输入公司名称，如：字节跳动、腾讯、阿里巴巴、美团..."
          size="large"
          clearable
          class="company-input"
          @keyup.enter.native="search"
        >
          <i slot="prefix" class="el-icon-office-building"></i>
        </el-input>
        <el-select v-model="jobTitle" placeholder="岗位方向（可选）" clearable size="large" class="job-select">
          <el-option label="后端开发" value="后端开发" />
          <el-option label="前端开发" value="前端开发" />
          <el-option label="全栈开发" value="全栈开发" />
          <el-option label="移动端开发" value="移动端开发" />
          <el-option label="算法工程师" value="算法工程师" />
          <el-option label="大数据工程师" value="大数据工程师" />
          <el-option label="测试工程师" value="测试工程师" />
          <el-option label="运维/DevOps" value="运维/DevOps" />
          <el-option label="网络安全" value="网络安全" />
          <el-option label="产品经理" value="产品经理" />
          <el-option label="UI/UX设计师" value="UI/UX设计师" />
          <el-option label="平面设计师" value="平面设计师" />
          <el-option label="数据分析师" value="数据分析师" />
          <el-option label="市场营销" value="市场营销" />
          <el-option label="运营专员" value="运营专员" />
          <el-option label="新媒体运营" value="新媒体运营" />
          <el-option label="人力资源" value="人力资源" />
          <el-option label="行政管理" value="行政管理" />
          <el-option label="销售/商务" value="销售/商务" />
          <el-option label="会计/财务" value="会计/财务" />
          <el-option label="项目管理" value="项目管理" />
          <el-option label="客户服务" value="客户服务" />
        </el-select>
        <el-button type="primary" size="large" :loading="loading" :disabled="!company.trim()" @click="search">
          🔍 AI 查询
        </el-button>
      </div>
      <div class="hot-companies">
        <span class="hot-label">热门：</span>
        <el-tag v-for="c in hotCompanies" :key="c" size="small" class="hot-tag" @click="quickSearch(c)">{{ c }}</el-tag>
      </div>
    </el-card>

    <!-- 加载中 -->
    <div v-if="loading" class="loading-wrap">
      <div class="loading-icon">🤖</div>
      <div class="loading-text">AI 正在检索 <b>{{ company }}</b> 的面试情报，请稍候...</div>
      <el-progress :percentage="loadingPercent" :show-text="false" class="loading-bar" />
    </div>

    <!-- 结果展示 -->
    <div v-if="result && !loading" class="result-wrap">
      <div class="result-header">
        <div class="result-company">{{ result.company }}</div>
        <div class="result-meta">
          <el-tag type="primary" v-if="result.jobTitle">{{ result.jobTitle }}</el-tag>
          <el-tag :type="difficultyTagType(result.difficulty)">难度：{{ result.difficulty }}</el-tag>
          <el-button type="primary" size="small" @click="startInterview">🚀 发起模拟面试</el-button>
        </div>
      </div>

      <el-row :gutter="16">
        <!-- 左列 -->
        <el-col :span="10">
          <el-card shadow="hover" class="section-card">
            <div slot="header" class="card-title">📋 面试流程</div>
            <div class="process-text">{{ result.process }}</div>
          </el-card>
          <el-card shadow="hover" class="section-card">
            <div slot="header" class="card-title">💡 面试技巧</div>
            <ul class="tips-list">
              <li v-for="(tip, i) in result.tips" :key="i">
                <span class="tip-num">{{ i + 1 }}</span>{{ tip }}
              </li>
            </ul>
          </el-card>
        </el-col>

        <!-- 右列：题目列表 -->
        <el-col :span="14">
          <el-card shadow="hover" class="section-card">
            <div slot="header" class="card-header">
              <span class="card-title">❓ 常见面试题（{{ filteredQuestions.length }} 道）</span>
              <el-radio-group v-model="qFilter" size="mini" @change="onFilterChange">
                <el-radio-button label="all">全部</el-radio-button>
                <el-radio-button label="technical">技术题</el-radio-button>
                <el-radio-button label="behavioral">行为题</el-radio-button>
              </el-radio-group>
            </div>

            <div class="questions-list">
              <div
                v-for="(q, i) in visibleQuestions"
                :key="i"
                class="question-item"
                @click="openQuestion(q)"
              >
                <div class="q-header">
                  <span class="q-num">{{ i + 1 }}</span>
                  <el-tag size="mini" :type="q.type === 'technical' ? 'primary' : 'warning'">
                    {{ q.type === 'technical' ? '技术' : '行为' }}
                  </el-tag>
                  <el-tag v-for="tag in (q.tags || []).slice(0, 2)" :key="tag" size="mini" type="info">{{ tag }}</el-tag>
                  <i class="el-icon-arrow-right q-arrow"></i>
                </div>
                <div class="q-content">{{ q.content }}</div>
              </div>
            </div>

            <!-- 查看更多 -->
            <div class="load-more-wrap">
              <div v-if="hasMore" class="load-more-info">已展示 {{ visibleQuestions.length }} / {{ filteredQuestions.length }} 道</div>
              <el-button v-if="hasMore" type="text" size="small" class="load-more-btn" @click="loadMore">
                查看更多 <i class="el-icon-arrow-down"></i>
              </el-button>
              <div v-if="!hasMore && filteredQuestions.length > defaultPageSize" class="all-loaded">✅ 已加载全部题目</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 空状态引导 -->
    <div v-if="!result && !loading" class="guide-wrap">
      <div class="guide-icon">🔍</div>
      <div class="guide-text">搜索任意公司，获取 AI 整理的面试情报</div>
      <div class="guide-examples">
        <div v-for="example in examples" :key="example.company" class="example-card" @click="quickSearch(example.company)">
          <div class="example-company">{{ example.company }}</div>
          <div class="example-desc">{{ example.desc }}</div>
        </div>
      </div>
    </div>

    <!-- 题目详情抽屉 -->
    <el-drawer
      :visible.sync="drawerVisible"
      direction="rtl"
      size="520px"
      :show-close="true"
      :wrapper-closable="true"
      class="question-drawer"
    >
      <div slot="title" class="drawer-title">
        <el-tag size="small" :type="activeQuestion && activeQuestion.type === 'technical' ? 'primary' : 'warning'">
          {{ activeQuestion && activeQuestion.type === 'technical' ? '技术题' : '行为题' }}
        </el-tag>
        <span style="margin-left:8px;">题目详情</span>
      </div>

      <div v-if="activeQuestion" class="drawer-body">
        <!-- 题目内容 -->
        <div class="drawer-question">{{ activeQuestion.content }}</div>
        <div class="drawer-tags">
          <el-tag v-for="tag in (activeQuestion.tags || [])" :key="tag" size="mini" type="info" style="margin-right:6px;">{{ tag }}</el-tag>
        </div>

        <el-divider></el-divider>

        <!-- 参考答案区 -->
        <div class="answer-section">
          <div class="answer-label">📝 参考答案</div>

          <!-- 未加载 -->
          <div v-if="!activeQuestion._answerLoaded && !answerLoading" class="answer-placeholder">
            <el-button type="primary" size="small" @click="loadAnswer">🤖 AI 生成参考答案</el-button>
          </div>

          <!-- 加载中 -->
          <div v-if="answerLoading" class="answer-loading">
            <i class="el-icon-loading"></i> AI 正在生成答案，请稍候...
          </div>

          <!-- 已加载 -->
          <div v-if="activeQuestion._answerLoaded && !answerLoading">
            <markdown-render v-if="activeQuestion.answer" :content="activeQuestion.answer" />
            <div v-else class="answer-empty">暂无参考答案</div>
          </div>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script>
import { getCompanyIntel, getCompanyQuestionAnswer } from '@/api/company'
import MarkdownRender from '@/components/common/MarkdownRender.vue'

export default {
  name: 'CompanyIntel',
  components: { MarkdownRender },
  data() {
    return {
      company: '',
      jobTitle: '',
      loading: false,
      loadingPercent: 0,
      result: null,
      qFilter: 'all',
      defaultPageSize: 11,
      currentPage: 1,
      pageIncrement: 10,
      // 抽屉
      drawerVisible: false,
      activeQuestion: null,
      answerLoading: false,
      // 答案缓存 content -> answer
      answerCache: {},
      hotCompanies: ['字节跳动', '腾讯', '阿里巴巴', '美团', '京东', '百度', '华为', '网易', '滴滴', '快手'],
      examples: [
        { company: '字节跳动', desc: '以算法题著称，多轮技术面' },
        { company: '腾讯', desc: '强调项目深度和技术原理' },
        { company: '阿里巴巴', desc: '注重系统设计和业务理解' },
        { company: '华为', desc: '笔试+技术面+HR面流程完整' }
      ]
    }
  },
  computed: {
    filteredQuestions() {
      if (!this.result) return []
      if (this.qFilter === 'all') return this.result.questions
      return this.result.questions.filter(q => q.type === this.qFilter)
    },
    visibleQuestions() {
      const limit = this.defaultPageSize + (this.currentPage - 1) * this.pageIncrement
      return this.filteredQuestions.slice(0, limit)
    },
    hasMore() {
      return this.visibleQuestions.length < this.filteredQuestions.length
    }
  },
  methods: {
    async search() {
      if (!this.company.trim()) return
      this.loading = true
      this.result = null
      this.loadingPercent = 0
      this.currentPage = 1
      this.qFilter = 'all'
      this.answerCache = {}

      const timer = setInterval(() => {
        if (this.loadingPercent < 90) this.loadingPercent += 10
      }, 800)

      try {
        const res = await getCompanyIntel(this.company.trim(), this.jobTitle)
        this.result = res.data || res
        this.loadingPercent = 100
      } catch (e) {
        this.$message.error('查询失败，请重试')
      } finally {
        clearInterval(timer)
        this.loading = false
      }
    },
    quickSearch(c) {
      this.company = c
      this.search()
    },
    startInterview() {
      this.$router.push({ path: '/interview/config', query: { jobTitle: this.result.jobTitle || '' } })
    },
    difficultyTagType(d) {
      const map = { '初级': 'success', '中级': 'warning', '高级': 'danger' }
      return map[d] || 'info'
    },
    onFilterChange() {
      this.currentPage = 1
    },
    loadMore() {
      this.currentPage += 1
    },
    // 点击题目打开抽屉
    openQuestion(q) {
      // 创建一个带响应式标记的副本
      this.activeQuestion = {
        ...q,
        _answerLoaded: !!this.answerCache[q.content],
        answer: this.answerCache[q.content] || ''
      }
      this.drawerVisible = true
    },
    // 加载当前题目答案
    async loadAnswer() {
      if (!this.activeQuestion || this.answerLoading) return
      this.answerLoading = true
      try {
        const res = await getCompanyQuestionAnswer(
          this.result.company,
          this.result.jobTitle || '',
          this.activeQuestion.content
        )
        const answer = (res.data || res).answer || ''
        // 缓存
        this.answerCache[this.activeQuestion.content] = answer
        // 更新 activeQuestion（需重新赋值触发响应式）
        this.activeQuestion = { ...this.activeQuestion, answer, _answerLoaded: true }
      } catch (e) {
        this.$message.error('答案获取失败，请重试')
      } finally {
        this.answerLoading = false
      }
    }
  }
}
</script>

<style scoped>
.company-intel-page { max-width: 1000px; margin: 0 auto; padding: 20px 16px; }

.search-card { margin-bottom: 20px; }
.search-header { text-align: center; margin-bottom: 20px; }
.search-title { font-size: 22px; font-weight: bold; color: #111; margin-bottom: 6px; }
.search-desc { font-size: 14px; color: #888; }
.search-form { display: flex; gap: 10px; align-items: center; margin-bottom: 14px; }
.company-input { flex: 1; }
.job-select { width: 160px; }
.hot-companies { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
.hot-label { font-size: 13px; color: #888; }
.hot-tag { cursor: pointer; transition: all 0.15s; }
.hot-tag:hover { color: #ff9900; border-color: #ff9900; background: #fff8ee; }

.loading-wrap { text-align: center; padding: 60px 0; }
.loading-icon { font-size: 48px; margin-bottom: 16px; animation: bounce 1s infinite; }
@keyframes bounce { 0%,100%{transform:translateY(0)} 50%{transform:translateY(-10px)} }
.loading-text { font-size: 16px; color: #555; margin-bottom: 20px; }
.loading-bar { max-width: 400px; margin: 0 auto; }

.result-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; flex-wrap: wrap; gap: 10px; }
.result-company { font-size: 22px; font-weight: bold; color: #111; }
.result-meta { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.section-card { margin-bottom: 16px; }
.card-title { font-size: 15px; font-weight: bold; color: #111; border-left: 4px solid #ff9900; padding-left: 10px; }
.card-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 8px; }
.process-text { font-size: 14px; color: #444; line-height: 1.8; }
.tips-list { list-style: none; padding: 0; margin: 0; }
.tips-list li { font-size: 14px; color: #444; padding: 8px 0; border-bottom: 1px solid #f5f5f5; display: flex; align-items: flex-start; gap: 8px; line-height: 1.6; }
.tips-list li:last-child { border-bottom: none; }
.tip-num { background: #ff9900; color: #fff; border-radius: 50%; width: 20px; height: 20px; display: inline-flex; align-items: center; justify-content: center; font-size: 12px; flex-shrink: 0; }

/* 题目列表 */
.questions-list { }
.question-item { padding: 10px 8px; border-bottom: 1px solid #f0f0f0; cursor: pointer; border-radius: 6px; transition: background 0.15s; }
.question-item:last-child { border-bottom: none; }
.question-item:hover { background: #fff8ee; }
.q-header { display: flex; align-items: center; gap: 6px; margin-bottom: 5px; flex-wrap: wrap; }
.q-num { background: #f0f0f0; color: #555; border-radius: 4px; padding: 1px 6px; font-size: 12px; font-weight: bold; }
.q-arrow { margin-left: auto; color: #ccc; font-size: 12px; }
.question-item:hover .q-arrow { color: #ff9900; }
.q-content { font-size: 14px; color: #333; line-height: 1.6; padding-left: 2px; }
.question-item:hover .q-content { color: #ff9900; }

.load-more-wrap { text-align: center; padding: 14px 0 4px; }
.load-more-info { font-size: 12px; color: #aaa; margin-bottom: 6px; }
.load-more-btn { font-size: 13px; color: #ff9900; }
.load-more-btn:hover { color: #e68800; }
.all-loaded { font-size: 12px; color: #bbb; }

/* 空状态 */
.guide-wrap { text-align: center; padding: 40px 0; }
.guide-icon { font-size: 56px; margin-bottom: 12px; }
.guide-text { font-size: 16px; color: #888; margin-bottom: 24px; }
.guide-examples { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; max-width: 700px; margin: 0 auto; }
.example-card { background: #fff; border: 1px solid #e8e8e8; border-radius: 8px; padding: 16px; cursor: pointer; transition: all 0.15s; text-align: left; }
.example-card:hover { border-color: #ff9900; box-shadow: 0 2px 8px rgba(255,153,0,0.15); }
.example-company { font-size: 15px; font-weight: bold; color: #111; margin-bottom: 4px; }
.example-desc { font-size: 12px; color: #888; }

/* 抽屉 */
.drawer-title { font-size: 15px; font-weight: bold; display: flex; align-items: center; }
.drawer-body { padding: 4px 8px; }
.drawer-question { font-size: 16px; color: #111; line-height: 1.7; font-weight: 500; margin-bottom: 10px; }
.drawer-tags { margin-bottom: 4px; }
.answer-section { margin-top: 4px; }
.answer-label { font-size: 14px; font-weight: bold; color: #ff9900; margin-bottom: 12px; }
.answer-placeholder { text-align: center; padding: 30px 0; }
.answer-loading { text-align: center; padding: 30px 0; color: #888; font-size: 14px; }
.answer-empty { color: #bbb; font-size: 14px; }
</style>
