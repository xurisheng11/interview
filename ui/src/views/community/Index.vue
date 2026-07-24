<template>
  <div class="community-page">
    <!-- 顶部分类 Tab -->
    <el-card shadow="never" class="tab-card">
      <div class="tab-row">
        <el-tabs v-model="activeCategory" @tab-click="handleCategoryChange">
          <el-tab-pane v-for="cat in categories" :key="cat.value" :label="cat.label" :name="cat.value" />
        </el-tabs>
        <div class="tab-actions">
          <el-radio-group v-model="sortBy" size="small" @change="loadArticles">
            <el-radio-button label="hot">🔥 热度</el-radio-button>
            <el-radio-button label="new">🕐 最新</el-radio-button>
          </el-radio-group>
          <el-button type="primary" size="small" icon="el-icon-plus" @click="showAIDialog = true">
            AI 知识获取
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 文章列表 -->
    <el-card shadow="never" class="list-card">
      <!-- 骨架屏：首次加载时显示 -->
      <div v-if="loading && articles.length === 0">
        <div v-for="i in 8" :key="i" class="skeleton-item">
          <div class="skeleton-main">
            <div class="skeleton-line skeleton-title"></div>
            <div class="skeleton-line skeleton-tags"></div>
          </div>
          <div class="skeleton-stats">
            <div class="skeleton-line skeleton-stat"></div>
          </div>
        </div>
      </div>
      <!-- 加载更多时的 loading 遮罩 -->
      <div v-if="loading && articles.length > 0" class="loading-mask">
        <i class="el-icon-loading"></i> 加载中...
      </div>
      <!-- 正常内容 -->
      <div v-if="!loading && articles.length === 0" class="empty-tip">
        <i class="el-icon-document"></i>
        <p>正在为你准备内容，稍后刷新试试</p>
        <el-button type="primary" size="small" @click="loadArticles">刷新</el-button>
      </div>
      <div
        v-for="article in articles"
        :key="article.articleId || article.id"
        class="article-item"
        @click="goArticle(article.articleId || article.id)"
      >
        <div class="article-main">
          <div class="article-title">{{ article.title }}</div>
          <div class="article-meta">
            <el-tag
              v-for="tag in (article.tags || []).slice(0, 3)"
              :key="tag"
              size="mini"
              type="info"
              class="meta-tag"
            >{{ tag }}</el-tag>
            <el-tag v-if="article.isAiGenerated" size="mini" type="warning" class="meta-tag">AI生成</el-tag>
          </div>
        </div>
        <div class="article-stats">
          <span class="stat-item">👍 {{ article.likeCount || 0 }}</span>
          <span class="stat-item">⭐ {{ article.collectCount || 0 }}</span>
          <span class="stat-item">💬 {{ article.commentCount || 0 }}</span>
          <span class="stat-item time">{{ formatTime(article.createdAt) }}</span>
        </div>
      </div>

      <div class="pagination-wrap">
        <el-pagination
          background
          layout="prev, pager, next"
          :total="total"
          :page-size="pageSize"
          :current-page.sync="page"
          @current-change="loadArticles"
        />
      </div>
    </el-card>

    <!-- AI 知识获取对话框 -->
    <el-dialog title="🤖 AI 知识获取" :visible.sync="showAIDialog" width="480px" @close="resetAIForm">
      <el-form :model="aiForm" label-width="80px">
        <el-form-item label="知识点">
          <el-input
            v-model="aiForm.topic"
            placeholder="请输入想了解的知识点，如：Redis持久化机制"
            clearable
          />
        </el-form-item>
        <el-form-item label="岗位方向">
          <el-select v-model="aiForm.jobCategory" placeholder="请选择岗位方向" style="width:100%">
            <el-option v-for="cat in categories" :key="cat.value" :label="cat.label" :value="cat.value" />
          </el-select>
        </el-form-item>
      </el-form>
      <div slot="footer">
        <el-button @click="showAIDialog = false">取消</el-button>
        <el-button
          type="primary"
          :loading="aiGenerating"
          :disabled="!aiForm.topic || !aiForm.jobCategory"
          @click="generateAIArticle"
        >生成文章</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getArticles, generateAIArticle } from '@/api/community'

export default {
  name: 'CommunityIndex',
  data() {
    return {
      loading: false,
      articles: [],
      total: 0,
      page: 1,
      pageSize: 15,
      activeCategory: 'all',
      sortBy: 'hot',
      showAIDialog: false,
      aiGenerating: false,
      aiForm: { topic: '', jobCategory: '' },
      categories: [
        { label: '全部', value: 'all' },
        { label: '后端开发', value: 'backend' },
        { label: '前端开发', value: 'frontend' },
        { label: '大数据', value: 'bigdata' },
        { label: 'AI/算法', value: 'ai' },
        { label: '会计/财务', value: 'accounting' },
        { label: '通用', value: 'general' }
      ]
    }
  },
  created() {
    this.initCategory()
    this.loadArticles()
  },
  methods: {
    // 根据用户信息自动选择分类
    initCategory() {
      const userInfo = this.$store.getters['user/userInfo']
      if (!userInfo) return
      // 尝试从用户 bio 或 role 里匹配分类关键词（简单启发式）
      const hint = ((userInfo.bio || '') + ' ' + (userInfo.nickname || '')).toLowerCase()
      const mapping = [
        { keywords: ['前端', 'frontend', 'vue', 'react', 'js', 'css'], cat: 'frontend' },
        { keywords: ['后端', 'backend', 'java', 'go', 'python', 'spring', 'mysql'], cat: 'backend' },
        { keywords: ['大数据', 'bigdata', 'spark', 'hadoop', 'flink', 'hive'], cat: 'bigdata' },
        { keywords: ['ai', '算法', '机器学习', '深度学习', 'ml', 'nlp'], cat: 'ai' },
        { keywords: ['会计', '财务', '审计', 'accounting', '税务'], cat: 'accounting' },
      ]
      for (const { keywords, cat } of mapping) {
        if (keywords.some(k => hint.includes(k))) {
          this.activeCategory = cat
          return
        }
      }
    },
    async loadArticles() {
      this.loading = true
      try {
        const params = {
          page: this.page,
          pageSize: this.pageSize,
          sortBy: this.sortBy
        }
        if (this.activeCategory !== 'all') params.jobCategory = this.activeCategory
        const res = await getArticles(params)
        const d = res.data || res
        this.articles = Array.isArray(d) ? d : (d.list || d.items || [])
        this.total = d.total || (Array.isArray(d) ? d.length : 0)
      } catch (e) {
        this.$message.error('加载文章失败')
      } finally {
        this.loading = false
      }
    },
    handleCategoryChange() {
      this.page = 1
      this.loadArticles()
    },
    async generateAIArticle() {
      this.aiGenerating = true
      try {
        await generateAIArticle({ topic: this.aiForm.topic, jobCategory: this.aiForm.jobCategory })
        this.$message.success('AI 文章生成成功！')
        this.showAIDialog = false
        this.page = 1
        this.loadArticles()
      } catch (e) {
        const code = e?.response?.data?.code || e?.response?.status
        if (code === 429) {
          this.$message.error('今日 AI 生成次数已用完，明天再试')
        } else {
          this.$message.error('AI 生成失败，请重试')
        }
      } finally {
        this.aiGenerating = false
      }
    },
    resetAIForm() {
      this.aiForm = { topic: '', jobCategory: '' }
    },
    goArticle(id) {
      this.$router.push(`/community/${id}`)
    },
    formatTime(val) {
      if (!val) return ''
      const d = new Date(val)
      if (isNaN(d.getTime())) return val
      const pad = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
    }
  }
}
</script>

<style scoped>
.community-page { max-width: 960px; margin: 0 auto; padding: 20px 16px; }
.tab-card { margin-bottom: 16px; }
.tab-row { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 12px; }
.tab-row :deep(.el-tabs__header) { margin: 0; }
.tab-actions { display: flex; align-items: center; gap: 12px; }
.list-card { }
.article-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  gap: 16px;
}
.article-item:last-child { border-bottom: none; }
.article-item:hover .article-title { color: #ff9900; text-decoration: underline; }
.article-main { flex: 1; }
.article-title { font-size: 15px; font-weight: bold; color: #0066c0; margin-bottom: 8px; line-height: 1.5; }
.article-meta { display: flex; flex-wrap: wrap; gap: 6px; }
.meta-tag { }
.article-stats { display: flex; align-items: center; gap: 12px; flex-shrink: 0; flex-wrap: wrap; }
.stat-item { font-size: 13px; color: #888; white-space: nowrap; }
.stat-item.time { color: #bbb; }
.pagination-wrap { display: flex; justify-content: center; margin-top: 20px; }
.empty-tip { text-align: center; color: #999; padding: 40px 0; font-size: 14px; }
.empty-tip p { margin: 12px 0; }
.loading-mask { text-align: center; color: #999; padding: 12px 0; font-size: 13px; }

/* 骨架屏 */
.skeleton-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
  gap: 16px;
}
.skeleton-main { flex: 1; }
.skeleton-stats { flex-shrink: 0; width: 120px; }
.skeleton-line {
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
  border-radius: 4px;
}
.skeleton-title { height: 18px; width: 70%; margin-bottom: 10px; }
.skeleton-tags { height: 14px; width: 40%; }
.skeleton-stat { height: 14px; width: 100%; }
@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>
