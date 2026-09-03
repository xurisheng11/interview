<template>
  <div class="contributions-page">
    <!-- 顶部 -->
    <div class="page-header">
      <h2 class="page-title">🏢 公司面试题库</h2>
      <p class="page-sub">
        浏览其他用户贡献的真实面试题，投票帮助筛选高质量题目。<br>
        <span class="highlight">提交题目获得积分，用积分解锁已验证的优质真题 ♻️</span>
      </p>
    </div>

    <!-- 搜索栏 -->
    <div class="search-bar">
      <el-input
        v-model="searchCompany"
        placeholder="输入公司名称搜索（如：字节跳动）"
        clearable
        @keyup.enter.native="handleSearch"
        class="search-input"
      >
        <i slot="prefix" class="el-icon-search"></i>
      </el-input>
      <el-button type="primary" class="search-btn" @click="handleSearch">搜索</el-button>
      <el-button @click="handleContribute">
        <i class="el-icon-plus"></i> 我要贡献题目
      </el-button>
    </div>

    <!-- 降级提示 -->
    <div class="fallback-tip" v-if="fallback">
      <div class="fallback-icon">🔍</div>
      <div class="fallback-content">
        <div class="fallback-title">{{ fallback.message }}</div>
        <div class="fallback-similar" v-if="fallback.similar && fallback.similar.length">
          <span>相似公司题目：</span>
          <el-tag
            v-for="q in fallback.similar.slice(0, 5)"
            :key="q.id"
            size="mini"
            class="similar-tag"
            @click="goToCompany(q.company)"
          >
            {{ q.company }}
          </el-tag>
        </div>
        <div class="fallback-actions">
          <el-button size="mini" type="primary" @click="handleContribute">
            成为第一个贡献者 →
          </el-button>
        </div>
      </div>
    </div>

    <!-- 结果 -->
    <div v-if="!fallback || results.length">
      <!-- 切换 tabs -->
      <el-tabs v-model="activeTab" class="result-tabs">
        <el-tab-pane label="已验证真题" name="verified">
          <span slot="label">
            ✅ 已验证 <el-badge :value="verifiedList.length" type="success" />
          </span>
        </el-tab-pane>
        <el-tab-pane label="待验证" name="pending">
          <span slot="label">
            ⏳ 待验证 <el-badge :value="pendingList.length" type="warning" />
          </span>
        </el-tab-pane>
      </el-tabs>

      <!-- 积分提示条 -->
      <div class="credits-bar" v-if="activeTab === 'verified'">
        <span>💎 当前积分：<strong>{{ credits }}</strong></span>
        <span class="credits-tip" v-if="credits < 3">积分不足？<el-link type="primary" @click="handleContribute">贡献题目赚积分</el-link></span>
      </div>

      <!-- 已验证列表 -->
      <div v-if="activeTab === 'verified'" class="questions-grid">
        <div v-if="verifiedList.length === 0" class="empty-state">
          <div class="empty-icon">📭</div>
          <div class="empty-title">暂无已验证题目</div>
          <div class="empty-sub">还没有用户通过验证的题目，成为第一个贡献者吧！</div>
          <el-button type="primary" size="small" @click="handleContribute">贡献第一道题</el-button>
        </div>
        <div
          v-for="q in verifiedList"
          :key="q.id"
          class="question-card verified"
        >
          <div class="q-company">{{ q.company }} · {{ q.round || '通用' }}</div>
          <div class="q-content">{{ q.content }}</div>
          <div class="q-meta">
            <el-tag size="mini" :type="tagType(q.difficulty)">{{ difficultyText(q.difficulty) }}</el-tag>
            <el-tag size="mini" type="info">{{ typeText(q.questionType) }}</el-tag>
            <span class="q-year" v-if="q.year"> {{ q.year }}年</span>
          </div>
          <div class="q-stats">
            <span class="stat-item">👍 {{ q.helpfulCount || 0 }}</span>
            <span class="stat-item">👁 {{ q.useCount || 0 }} 次使用</span>
          </div>
        </div>
      </div>

      <!-- 待验证列表 -->
      <div v-if="activeTab === 'pending'" class="questions-grid">
        <div v-if="pendingList.length === 0" class="empty-state">
          <div class="empty-icon">✅</div>
          <div class="empty-title">全部题目已验证完毕</div>
        </div>
        <div
          v-for="q in pendingList"
          :key="q.id"
          class="question-card pending"
        >
          <div class="q-company">{{ q.company }} · {{ q.round || '通用' }}</div>
          <div class="q-content">{{ q.content }}</div>
          <div class="q-meta">
            <el-tag size="mini" :type="tagType(q.difficulty)">{{ difficultyText(q.difficulty) }}</el-tag>
            <el-tag size="mini" type="info">{{ typeText(q.questionType) }}</el-tag>
            <span class="q-year" v-if="q.year"> {{ q.year }}年</span>
          </div>
          <div class="q-stats">
            <span class="stat-item">👍 {{ q.helpfulCount || 0 }} 有用</span>
            <span class="stat-item">👎 {{ q.uselessCount || 0 }} 没用</span>
          </div>
          <!-- 投票按钮 -->
          <div class="vote-actions">
            <el-button
              size="mini"
              type="success"
              :disabled="votedMap[q.id] === 'helpful'"
              @click="handleVote(q.id, 'helpful')"
            >
              👍 有用
            </el-button>
            <el-button
              size="mini"
              type="danger"
              :disabled="votedMap[q.id] === 'useless'"
              @click="handleVote(q.id, 'useless')"
            >
              👎 没用
            </el-button>
            <el-button size="mini" type="text" class="report-btn" @click="showReport(q)">
              🚩 举报
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 举报弹窗 -->
    <el-dialog title="🚩 举报题目" :visible.sync="reportVisible" width="400px">
      <p class="report-q">{{ reportQuestion?.content }}</p>
      <el-input type="textarea" v-model="reportReason" :rows="3" placeholder="请说明举报原因（选填）" style="margin-top: 12px"></el-input>
      <span slot="footer">
        <el-button @click="reportVisible = false">取消</el-button>
        <el-button type="danger" @click="handleReport">确认举报</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import {
  getCompanyContributedQuestions,
  voteContributedQuestion,
  reportContributedQuestion,
  getUserCredits,
  searchQuestionsFallback
} from '@/api/contribution'

export default {
  name: 'ContributionsPage',

  data() {
    return {
      searchCompany: '',
      credits: 0,
      activeTab: 'verified',
      results: [],
      fallback: null,
      votedMap: {},
      reportVisible: false,
      reportQuestion: null,
      reportReason: ''
    }
  },

  computed: {
    verifiedList() {
      return this.results.filter(q => q.status === 'verified' || q.status === 'approved')
    },
    pendingList() {
      return this.results.filter(q => q.status === 'pending')
    }
  },

  created() {
    // 如果路由有 company 参数则加载
    const company = this.$route.query.company
    if (company) {
      this.searchCompany = company
      this.loadQuestions(company)
    }
    this.loadCredits()
  },

  methods: {
    async loadCredits() {
      try {
        const res = await getUserCredits()
        this.credits = res?.credits || 0
      } catch (e) {}
    },

    async handleSearch() {
      const c = this.searchCompany.trim()
      if (!c) {
        this.$message.warning('请输入公司名称')
        return
      }
      this.loadQuestions(c)
    },

    async loadQuestions(company) {
      try {
        const res = await getCompanyContributedQuestions(company, '')
        const all = [...(res?.verified || []), ...(res?.approved || [])]
        this.results = all
        this.credits = res?.credits || this.credits
        this.fallback = null

        if (all.length === 0) {
          // 搜不到，调用降级接口
          await this.loadFallback(company)
        }
      } catch (e) {
        this.$message.error('加载失败')
      }
    },

    async loadFallback(company) {
      try {
        const res = await searchQuestionsFallback(company, '')
        this.fallback = res
      } catch (e) {
        this.fallback = {
          message: `未找到【${company}】的专属题库`,
          similar: []
        }
      }
    },

    handleContribute() {
      this.$router.push('/questions/contribute')
    },

    goToCompany(company) {
      this.searchCompany = company
      this.loadQuestions(company)
    },

    async handleVote(questionId, vote) {
      try {
        await voteContributedQuestion(questionId, vote)
        this.$set(this.votedMap, questionId, vote)
        // 更新本地计数
        const q = this.results.find(q => q.id === questionId)
        if (q) {
          if (vote === 'helpful') q.helpfulCount = (q.helpfulCount || 0) + 1
          else q.uselessCount = (q.uselessCount || 0) + 1
        }
        this.$message.success('投票成功 +1 积分')
        this.loadCredits()
      } catch (err) {
        this.$message.warning(err?.message || '投票失败')
      }
    },

    showReport(q) {
      this.reportQuestion = q
      this.reportReason = ''
      this.reportVisible = true
    },

    async handleReport() {
      if (!this.reportQuestion) return
      try {
        await reportContributedQuestion(this.reportQuestion.id, this.reportReason)
        this.$message.success('举报已收到，感谢反馈')
        this.reportVisible = false
      } catch (err) {
        this.$message.warning(err?.message || '举报失败')
      }
    },

    tagType(d) {
      return { easy: 'success', medium: 'warning', hard: 'danger' }[d] || 'info'
    },
    difficultyText(d) {
      return { easy: '简单', medium: '中等', hard: '困难' }[d] || d
    },
    typeText(t) {
      return {
        technical: '技术',
        behavioral: '行为',
        algorithm: '算法',
        'system-design': '系统设计'
      }[t] || t
    }
  }
}
</script>

<style scoped>
.contributions-page { max-width: 900px; margin: 0 auto; padding: 20px; }

.page-header { margin-bottom: 20px; }
.page-title { font-size: 20px; font-weight: bold; color: #333; margin-bottom: 8px; }
.page-sub { font-size: 13px; color: #666; line-height: 1.8; }
.highlight { color: #ff9900; }

/* Search */
.search-bar {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
  align-items: center;
}
.search-input { flex: 1; max-width: 400px; }
.search-btn { background: #ff9900 !important; border-color: #ff9900 !important; color: #111 !important; }

/* Fallback tip */
.fallback-tip {
  display: flex;
  gap: 16px;
  background: #fffbe6;
  border: 1px solid #ffe58f;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
  align-items: flex-start;
}
.fallback-icon { font-size: 28px; }
.fallback-title { font-weight: bold; color: #333; margin-bottom: 6px; }
.fallback-similar { display: flex; flex-wrap: wrap; gap: 6px; align-items: center; font-size: 13px; color: #555; }
.similar-tag { cursor: pointer; }
.similar-tag:hover { opacity: 0.8; }
.fallback-actions { margin-top: 10px; }

/* Credits bar */
.credits-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  background: #f5f5f5;
  border-radius: 6px;
  padding: 8px 14px;
  margin-bottom: 16px;
  font-size: 13px;
  color: #555;
}
.credits-tip { margin-left: auto; }

/* Tabs */
.result-tabs { margin-bottom: 16px; }

/* Questions grid */
.questions-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.question-card {
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 16px;
}

.question-card.verified {
  border-color: #b7eb8f;
  background: #f6ffed;
}

.question-card.pending {
  border-color: #ffe58f;
  background: #fffbe6;
}

.q-company { font-weight: bold; font-size: 13px; color: #ff9900; margin-bottom: 8px; }
.q-content { font-size: 14px; color: #222; line-height: 1.7; margin-bottom: 10px; }
.q-meta { display: flex; gap: 6px; align-items: center; flex-wrap: wrap; margin-bottom: 8px; }
.q-year { font-size: 12px; color: #999; }
.q-stats { display: flex; gap: 14px; font-size: 12px; color: #888; margin-bottom: 8px; }
.stat-item { display: flex; align-items: center; gap: 4px; }

/* Vote actions */
.vote-actions { display: flex; gap: 8px; align-items: center; border-top: 1px solid #f0f0f0; padding-top: 10px; }
.report-btn { color: #999 !important; font-size: 12px !important; margin-left: auto; }
.report-btn:hover { color: #f56c6c !important; }

/* Empty */
.empty-state { text-align: center; padding: 48px 0; background: #fafafa; border-radius: 8px; }
.empty-icon { font-size: 48px; }
.empty-title { font-size: 16px; font-weight: bold; color: #555; margin: 12px 0 6px; }
.empty-sub { font-size: 13px; color: #999; margin-bottom: 16px; }

/* Report */
.report-q { font-size: 13px; color: #555; background: #f5f5f5; padding: 10px; border-radius: 4px; }
</style>
