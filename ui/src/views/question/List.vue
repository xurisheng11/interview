<template>
  <div class="question-list">
    <!-- 筛选栏 -->
    <el-card shadow="never" class="filter-card">
      <el-form :inline="true" class="filter-form">
        <el-form-item label="岗位">
          <el-select v-model="filters.jobTitle" placeholder="全部岗位" clearable size="small">
            <el-option v-for="j in jobOptions" :key="j.value" :label="j.label" :value="j.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="难度">
          <el-select v-model="filters.difficulty" placeholder="全部难度" clearable size="small">
            <el-option label="初级" value="junior" />
            <el-option label="中级" value="middle" />
            <el-option label="高级" value="senior" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.type" placeholder="全部类型" clearable size="small">
            <el-option label="基础" value="basic" />
            <el-option label="算法" value="algorithm" />
            <el-option label="设计" value="design" />
            <el-option label="HR" value="hr" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input
            v-model="filters.keyword"
            placeholder="搜索题目..."
            size="small"
            style="width:200px"
            clearable
            @keyup.enter.native="loadQuestions"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="small" @click="loadQuestions">搜索</el-button>
          <el-button size="small" @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 题目列表 -->
    <el-card shadow="never" class="list-card" v-loading="loading">
      <div slot="header" class="card-header">
        <span class="card-title">📚 题库列表</span>
        <span class="total-tip">共 {{ total }} 道题目</span>
      </div>

      <div v-if="questions.length === 0 && !loading" class="empty-tip">
        暂无题目，换个条件试试
      </div>

      <div
        v-for="q in questions"
        :key="q.questionId || q.id"
        class="question-item"
        @click="goPractice(q.questionId || q.id)"
      >
        <div class="q-main">
          <div class="q-title">{{ q.content || q.title }}</div>
          <div class="q-meta">
            <el-tag size="mini" type="primary" effect="plain" class="meta-tag">{{ q.jobTitle }}</el-tag>
            <el-tag size="mini" :type="diffTagType(q.difficulty)" class="meta-tag">
              {{ diffLabel(q.difficulty) }}
            </el-tag>
            <el-tag v-for="tag in (q.tags || []).slice(0,3)" :key="tag" size="mini" type="info" class="meta-tag">
              {{ tag }}
            </el-tag>
          </div>
        </div>
        <div class="q-stats">
          <span class="stat-item">👥 {{ q.answerCount || 0 }} 人作答</span>
          <span class="stat-item">📊 均分 {{ q.avgScore != null ? q.avgScore : '—' }}</span>
          <el-button
            :type="q.collected ? 'warning' : 'default'"
            size="mini"
            icon="el-icon-star-off"
            @click.stop="toggleCollect(q)"
          >{{ q.collected ? '已收藏' : '收藏' }}</el-button>
        </div>
      </div>

      <!-- 分页 -->
      <div class="pagination-wrap">
        <el-pagination
          background
          layout="prev, pager, next, total"
          :total="total"
          :page-size="pageSize"
          :current-page.sync="page"
          @current-change="loadQuestions"
        />
      </div>
    </el-card>
  </div>
</template>

<script>
import { getQuestions, collectQuestion } from '@/api/question'

export default {
  name: 'QuestionList',
  data() {
    return {
      loading: false,
      questions: [],
      total: 0,
      page: 1,
      pageSize: 20,
      filters: {
        jobTitle: '',
        difficulty: '',
        type: '',
        keyword: ''
      },
      jobOptions: [
        { label: '后端开发', value: '后端开发' },
        { label: '前端开发', value: '前端开发' },
        { label: '大数据工程师', value: '大数据工程师' },
        { label: 'AI算法工程师', value: 'AI算法工程师' },
        { label: '会计/财务', value: '会计/财务' },
        { label: '产品经理', value: '产品经理' },
        { label: '测试工程师', value: '测试工程师' },
        { label: '运维工程师', value: '运维工程师' }
      ]
    }
  },
  created() {
    this.loadQuestions()
  },
  methods: {
    async loadQuestions() {
      this.loading = true
      try {
        const params = {
          page: this.page,
          pageSize: this.pageSize,
          ...this.filters
        }
        // 清空空值参数
        Object.keys(params).forEach(k => { if (!params[k]) delete params[k] })
        const res = await getQuestions(params)
        const d = res.data || res
        this.questions = Array.isArray(d) ? d : (d.list || d.items || [])
        this.total = d.total || (Array.isArray(d) ? d.length : 0)
      } catch (e) {
        this.$message.error('加载题目失败')
      } finally {
        this.loading = false
      }
    },
    resetFilters() {
      this.filters = { jobTitle: '', difficulty: '', type: '', keyword: '' }
      this.page = 1
      this.loadQuestions()
    },
    async toggleCollect(q) {
      const id = q.questionId || q.id
      try {
        await collectQuestion(id)
        q.collected = !q.collected
        this.$message.success(q.collected ? '收藏成功' : '已取消收藏')
      } catch (e) {
        this.$message.error('操作失败')
      }
    },
    goPractice(id) {
      this.$router.push(`/questions/${id}/practice`)
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
.question-list { max-width: 960px; margin: 0 auto; padding: 20px 16px; }
.filter-card { margin-bottom: 16px; }
.filter-form { display: flex; flex-wrap: wrap; gap: 4px; }
.list-card { }
.card-header { display: flex; align-items: center; justify-content: space-between; }
.card-title { font-size: 16px; font-weight: bold; color: #111; border-left: 4px solid #ff9900; padding-left: 10px; }
.total-tip { font-size: 13px; color: #999; }
.question-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 14px 0;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: background 0.15s;
  gap: 16px;
}
.question-item:last-child { border-bottom: none; }
.question-item:hover .q-title { color: #ff9900; text-decoration: underline; }
.q-main { flex: 1; }
.q-title { font-size: 15px; font-weight: bold; color: #0066c0; margin-bottom: 8px; line-height: 1.5; }
.q-meta { display: flex; flex-wrap: wrap; gap: 6px; }
.meta-tag { }
.q-stats { display: flex; align-items: center; gap: 12px; flex-shrink: 0; flex-wrap: wrap; }
.stat-item { font-size: 12px; color: #999; white-space: nowrap; }
.pagination-wrap { display: flex; justify-content: center; margin-top: 20px; }
.empty-tip { text-align: center; color: #999; padding: 40px 0; font-size: 14px; }
</style>
