<template>
  <div class="contribution-page">
    <!-- 顶部 Banner -->
    <div class="hero">
      <div class="hero-content">
        <h1 class="hero-title">📝 贡献面试题</h1>
        <p class="hero-sub">
          你有小红书、Boss 直聘、牛客网上看到的真实面试题吗？提交上来，帮助更多人，<strong>每道题奖励 2 积分</strong>，审核通过额外奖励 3 积分。<br>
          <span class="tips">积分可用于解锁其他公司的优质真题，实现互惠交换 ♻️</span>
        </p>
        <div class="credit-box">
          <span class="credit-icon">💎</span>
          <span>当前积分：<strong>{{ credits }}</strong></span>
          <span class="credit-desc">（提交题目得积分，看他人题目也消耗积分）</span>
        </div>
      </div>
    </div>

    <div class="main-container">
      <!-- 左侧：提交表单 -->
      <div class="form-panel">
        <el-card shadow="never" class="form-card">
          <div slot="header" class="card-header">
            <span>📋 提交新题目</span>
          </div>

          <el-form :model="form" label-position="top" class="contribution-form">
            <el-form-item label="公司名称" required>
              <el-input v-model="form.company" placeholder="如：字节跳动、阿里巴巴" clearable></el-input>
            </el-form-item>

            <el-form-item label="岗位名称">
              <el-input v-model="form.jobTitle" placeholder="如：后端开发、前端工程师" clearable></el-input>
            </el-form-item>

            <div class="row-2">
              <el-form-item label="面试年份">
                <el-input-number v-model="form.year" :min="2018" :max="2030" placeholder="年份"></el-input-number>
              </el-form-item>
              <el-form-item label="面试轮次">
                <el-select v-model="form.round" placeholder="选择轮次" clearable style="width: 100%">
                  <el-option label="一面" value="一面"></el-option>
                  <el-option label="二面" value="二面"></el-option>
                  <el-option label="三面" value="三面"></el-option>
                  <el-option label="HR面" value="HR面"></el-option>
                  <el-option label="终面" value="终面"></el-option>
                </el-select>
              </el-form-item>
            </div>

            <!-- 题目列表 -->
            <div class="questions-section">
              <div class="questions-header">
                <span class="label">面试题目</span>
                <span class="count">已添加 {{ form.questions.length }} 道</span>
              </div>

              <div v-for="(q, idx) in form.questions" :key="idx" class="question-item">
                <div class="q-header">
                  <span class="q-num">题目 {{ idx + 1 }}</span>
                  <el-button type="text" class="q-remove" @click="removeQuestion(idx)" v-if="form.questions.length > 1">
                    <i class="el-icon-delete"></i> 删除
                  </el-button>
                </div>
                <el-input
                  type="textarea"
                  v-model="q.content"
                  :rows="3"
                  :placeholder="'请输入面试题目内容...'"
                ></el-input>
                <div class="q-meta">
                  <el-select v-model="q.questionType" placeholder="题目类型" size="small" style="width: 120px">
                    <el-option label="技术题" value="technical"></el-option>
                    <el-option label="行为题" value="behavioral"></el-option>
                    <el-option label="手撕算法" value="algorithm"></el-option>
                    <el-option label="系统设计" value="system-design"></el-option>
                  </el-select>
                  <el-select v-model="q.difficulty" placeholder="难度" size="small" style="width: 100px">
                    <el-option label="简单" value="easy"></el-option>
                    <el-option label="中等" value="medium"></el-option>
                    <el-option label="困难" value="hard"></el-option>
                  </el-select>
                  <el-input v-model="q.tagsInput" size="small" placeholder="标签（逗号分隔）" style="width: 160px"></el-input>
                </div>
              </div>

              <el-button class="add-btn" @click="addQuestion">
                <i class="el-icon-plus"></i> 添加一道题目
              </el-button>
            </div>

            <!-- 提交按钮 -->
            <el-button
              type="primary"
              class="submit-btn"
              :loading="submitting"
              :disabled="!canSubmit"
              @click="handleSubmit"
            >
              {{ submitting ? '提交中...' : '提交贡献' }}
            </el-button>

            <div class="submit-tips" v-if="!canSubmit">
              <span v-if="!form.company">请填写公司名称</span>
              <span v-else-if="form.questions.length === 0">请至少添加1道题目</span>
              <span v-else-if="!form.questions[0]?.content">请填写至少1道题目的内容</span>
            </div>
          </el-form>
        </el-card>
      </div>

      <!-- 右侧：贡献指南 + 我的贡献 + 积分规则 -->
      <div class="sidebar">
        <!-- 积分规则 -->
        <el-card shadow="never" class="sidebar-card">
          <div slot="header" class="card-header">
            <span>💎 积分规则</span>
          </div>
          <div class="rule-list">
            <div class="rule-item">
              <span class="rule-icon">📤</span>
              <span class="rule-text">提交每道题</span>
              <span class="rule-credit">+2 积分</span>
            </div>
            <div class="rule-item">
              <span class="rule-icon">✅</span>
              <span class="rule-text">审核通过</span>
              <span class="rule-credit">+3 积分</span>
            </div>
            <div class="rule-item">
              <span class="rule-icon">👍</span>
              <span class="rule-text">投票验证题目</span>
              <span class="rule-credit">+1 积分</span>
            </div>
            <div class="rule-item">
              <span class="rule-icon">🚀</span>
              <span class="rule-text">题目被他人使用</span>
              <span class="rule-credit">+1 积分</span>
            </div>
            <div class="rule-item">
              <span class="rule-icon">🎁</span>
              <span class="rule-text">新用户注册</span>
              <span class="rule-credit">+10 积分</span>
            </div>
          </div>
        </el-card>

        <!-- 我的贡献 -->
        <el-card shadow="never" class="sidebar-card">
          <div slot="header" class="card-header">
            <span>📊 我的贡献</span>
          </div>
          <div v-if="loading" class="loading-placeholder">
            <i class="el-icon-loading"></i> 加载中...
          </div>
          <div v-else-if="myContributions.length === 0" class="empty-state">
            <div class="empty-icon">📝</div>
            <div class="empty-text">暂无贡献记录</div>
            <div class="empty-sub">提交第一道题开始吧！</div>
          </div>
          <div v-else class="contribution-list">
            <div v-for="c in myContributions" :key="c.id" class="contribution-item">
              <div class="c-company">{{ c.company }}</div>
              <div class="c-content">{{ c.content }}</div>
              <div class="c-meta">
                <el-tag :type="statusType(c.status)" size="mini">{{ statusText(c.status) }}</el-tag>
                <span class="c-date">{{ formatDate(c.createdAt) }}</span>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 贡献指南 -->
        <el-card shadow="never" class="sidebar-card">
          <div slot="header" class="card-header">
            <span>📖 贡献指南</span>
          </div>
          <ul class="guide-list">
            <li>题目来源于真实面试经历（小红书、牛客、Boss等平台分享）</li>
            <li>请勿提交广告、软文、虚假信息</li>
            <li>题目将经过社区投票验证，高质量题目标记为"已验证"</li>
            <li>举报超过 3 次的题目将重新审核</li>
            <li>每道题必须包含完整的题目内容，技术题建议包含考察知识点</li>
          </ul>
        </el-card>
      </div>
    </div>

    <!-- 提交成功弹窗 -->
    <el-dialog title="🎉 提交成功！" :visible.sync="successVisible" width="420px" center>
      <div class="success-content">
        <div class="success-icon">✅</div>
        <p>感谢您的贡献！</p>
        <p class="success-msg">每道题奖励 <strong>2 积分</strong>，审核通过后再奖励 3 积分</p>
        <p class="credit-tip">当前积分：<strong>{{ newCredits }}</strong></p>
      </div>
      <span slot="footer">
        <el-button type="primary" @click="successVisible = false">继续添加</el-button>
        <el-button @click="$router.push('/question/contributions')">查看全部题库</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import {
  submitContributedQuestion,
  getUserContributions,
  getUserCredits
} from '@/api/contribution'

export default {
  name: 'ContributeQuestion',

  data() {
    return {
      credits: 0,
      newCredits: 0,
      submitting: false,
      loading: true,
      successVisible: false,
      form: {
        company: '',
        jobTitle: '',
        year: new Date().getFullYear(),
        round: '',
        questions: [
          this.newQuestion()
        ]
      },
      myContributions: []
    }
  },

  computed: {
    canSubmit() {
      const q = this.form.questions[0]
      return this.form.company && q && q.content && q.content.trim()
    }
  },

  created() {
    this.loadData()
  },

  methods: {
    newQuestion() {
      return {
        content: '',
        questionType: 'technical',
        difficulty: 'medium',
        tagsInput: ''
      }
    },

    addQuestion() {
      if (this.form.questions.length >= 20) {
        this.$message.warning('每次最多提交 20 道题目')
        return
      }
      this.form.questions.push(this.newQuestion())
    },

    removeQuestion(idx) {
      this.form.questions.splice(idx, 1)
    },

    async loadData() {
      this.loading = true
      try {
        const [contribRes, creditRes] = await Promise.all([
          getUserContributions(),
          getUserCredits()
        ])
        this.myContributions = contribRes?.contributions || []
        this.credits = creditRes?.credits || 0
      } catch (e) {
        // ignore
      } finally {
        this.loading = false
      }
    },

    async handleSubmit() {
      if (!this.canSubmit) return
      this.submitting = true

      const questions = this.form.questions.map(q => ({
        content: q.content.trim(),
        questionType: q.questionType,
        difficulty: q.difficulty,
        tags: q.tagsInput ? q.tagsInput.split(/[,，]/).map(t => t.trim()).filter(Boolean) : []
      }))

      try {
        const res = await submitContributedQuestion({
          company: this.form.company.trim(),
          jobTitle: this.form.jobTitle.trim(),
          year: this.form.year,
          round: this.form.round,
          questions
        })

        this.newCredits = res.credits || this.credits + 2
        this.credits = this.newCredits

        // 重置表单
        this.form.questions = [this.newQuestion()]
        this.successVisible = true

        // 刷新贡献列表
        this.loadData()
      } catch (err) {
        this.$message.error(err?.message || '提交失败，请重试')
      } finally {
        this.submitting = false
      }
    },

    statusType(status) {
      const map = {
        pending: 'warning',
        approved: 'success',
        verified: 'primary',
        rejected: 'danger'
      }
      return map[status] || 'info'
    },

    statusText(status) {
      const map = {
        pending: '待审核',
        approved: '已通过',
        verified: '已验证',
        rejected: '已驳回'
      }
      return map[status] || status
    },

    formatDate(ts) {
      if (!ts) return ''
      const d = new Date(ts * 1000)
      return `${d.getMonth() + 1}/${d.getDate()}`
    }
  }
}
</script>

<style scoped>
.contribution-page {
  max-width: 1100px;
  margin: 0 auto;
  padding: 20px;
}

/* Hero */
.hero {
  background: linear-gradient(135deg, #ff9900 0%, #ff6b00 100%);
  border-radius: 12px;
  padding: 32px;
  color: #fff;
  margin-bottom: 24px;
}

.hero-title {
  font-size: 24px;
  font-weight: bold;
  margin-bottom: 10px;
  color: #fff;
}

.hero-sub {
  font-size: 14px;
  line-height: 1.8;
  opacity: 0.95;
  margin-bottom: 14px;
}

.tips {
  font-size: 13px;
  opacity: 0.85;
}

.credit-box {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: rgba(255,255,255,0.2);
  border-radius: 20px;
  padding: 6px 16px;
  font-size: 14px;
}

.credit-icon { font-size: 16px; }
.credit-desc { font-size: 12px; opacity: 0.8; }

/* Main layout */
.main-container {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}

/* Form panel */
.form-panel { flex: 1; min-width: 0; }

.form-card {
  border: 1px solid #e8e8e8;
  border-radius: 8px;
}

.card-header {
  font-weight: bold;
  font-size: 15px;
  color: #333;
}

.contribution-form .el-form-item { margin-bottom: 14px; }
.contribution-form .el-form-item__label { padding-bottom: 4px !important; font-weight: 600; }

.row-2 {
  display: flex;
  gap: 16px;
}
.row-2 .el-form-item { flex: 1; }

/* Questions */
.questions-section {
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 16px;
  background: #fafafa;
  margin-bottom: 16px;
}

.questions-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.questions-header .label { font-weight: 600; font-size: 14px; }
.questions-header .count { font-size: 12px; color: #999; }

.question-item {
  background: #fff;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 10px;
}

.q-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.q-num { font-weight: 600; font-size: 13px; color: #ff9900; }
.q-remove { color: #f56c6c !important; font-size: 12px; padding: 0; }

.q-meta {
  display: flex;
  gap: 8px;
  margin-top: 8px;
  flex-wrap: wrap;
}

.add-btn {
  width: 100%;
  border-style: dashed;
  color: #ff9900;
  border-color: #ff9900;
  background: transparent;
}
.add-btn:hover { background: #fff3e0; }

.submit-btn {
  width: 100%;
  height: 44px;
  font-size: 15px;
  background: #ff9900 !important;
  border-color: #ff9900 !important;
  color: #111 !important;
  font-weight: bold !important;
}

.submit-tips {
  text-align: center;
  font-size: 12px;
  color: #999;
  margin-top: 8px;
}

/* Sidebar */
.sidebar {
  width: 300px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.sidebar-card { border: 1px solid #e8e8e8; border-radius: 8px; }

/* Rules */
.rule-list { display: flex; flex-direction: column; gap: 10px; }

.rule-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
}

.rule-icon { font-size: 16px; }
.rule-text { flex: 1; color: #555; }
.rule-credit { color: #ff9900; font-weight: bold; font-size: 13px; }

/* My contributions */
.loading-placeholder { text-align: center; color: #999; padding: 20px; font-size: 13px; }

.empty-state { text-align: center; padding: 20px 0; }
.empty-icon { font-size: 32px; }
.empty-text { font-size: 14px; color: #555; margin-top: 6px; }
.empty-sub { font-size: 12px; color: #999; margin-top: 4px; }

.contribution-list { display: flex; flex-direction: column; gap: 10px; }

.contribution-item {
  border: 1px solid #f0f0f0;
  border-radius: 6px;
  padding: 10px;
}

.c-company { font-weight: 600; font-size: 13px; color: #333; }
.c-content { font-size: 12px; color: #666; margin-top: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.c-meta { display: flex; justify-content: space-between; align-items: center; margin-top: 6px; }
.c-date { font-size: 11px; color: #999; }

/* Guide */
.guide-list {
  padding-left: 18px;
  font-size: 12px;
  color: #555;
  line-height: 2;
}

/* Success dialog */
.success-content { text-align: center; }
.success-icon { font-size: 48px; margin-bottom: 12px; }
.success-msg { font-size: 14px; color: #555; margin-top: 4px; }
.credit-tip { font-size: 14px; color: #ff9900; margin-top: 6px; }
</style>
