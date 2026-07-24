<template>
  <div class="article-page" v-loading="loading">
    <div v-if="article">
      <!-- 文章主体 -->
      <el-card shadow="hover" class="article-card">
        <div class="article-header">
          <h1 class="article-title">{{ article.title }}</h1>
          <div class="article-meta">
            <el-tag v-for="tag in (article.tags || [])" :key="tag" size="mini" type="info" class="meta-tag">{{ tag }}</el-tag>
            <el-tag v-if="article.isAiGenerated" size="mini" type="warning">AI生成</el-tag>
            <span class="meta-time">{{ formatTime(article.createdAt) }}</span>
          </div>
        </div>
        <el-divider />
        <markdown-render :content="article.content" class="article-body" />
        <el-divider />
        <!-- 操作按钮 -->
        <div class="action-row">
          <el-button
            :type="article.liked ? 'primary' : 'default'"
            icon="el-icon-thumb"
            @click="handleLike"
          >👍 点赞 {{ article.likeCount || 0 }}</el-button>
          <el-button
            :type="article.collected ? 'warning' : 'default'"
            icon="el-icon-star-off"
            @click="handleCollect"
          >⭐ 收藏 {{ article.collectCount || 0 }}</el-button>
          <el-button @click="$router.push('/community')">← 返回社区</el-button>
        </div>
      </el-card>

      <!-- 评论区 -->
      <el-card shadow="never" class="comment-card">
        <div slot="header" class="card-header">
          <span class="card-title">💬 评论 ({{ comments.length }})</span>
        </div>

        <!-- 发表评论 -->
        <div class="comment-input-wrap">
          <el-input
            v-model="commentText"
            type="textarea"
            :rows="3"
            placeholder="发表你的评论..."
            resize="none"
            maxlength="500"
            show-word-limit
          />
          <div class="comment-actions">
            <el-button
              type="primary"
              size="small"
              :disabled="!commentText.trim()"
              :loading="submittingComment"
              @click="submitComment"
            >发表评论</el-button>
          </div>
        </div>

        <el-divider />

        <!-- 评论列表 -->
        <div v-if="commentsLoading" v-loading="true" style="height:80px"></div>
        <div v-else>
          <div v-if="comments.length === 0" class="empty-tip">暂无评论，来发表第一条吧</div>
          <div v-for="c in comments" :key="c.commentId || c.id" class="comment-item">
            <div class="comment-author">{{ c.username || c.nickname || '匿名用户' }}</div>
            <div class="comment-content">{{ c.content }}</div>
            <div class="comment-time">{{ formatTime(c.createdAt) }}</div>
          </div>
        </div>
      </el-card>
    </div>

    <el-empty v-if="!loading && !article" description="文章不存在或加载失败">
      <el-button @click="$router.push('/community')">返回社区</el-button>
    </el-empty>
  </div>
</template>

<script>
import { getArticle, likeArticle, collectArticle, getComments, addComment } from '@/api/community'
import MarkdownRender from '@/components/common/MarkdownRender.vue'

export default {
  name: 'ArticleDetail',
  components: { MarkdownRender },
  data() {
    return {
      loading: true,
      article: null,
      commentsLoading: false,
      comments: [],
      commentText: '',
      submittingComment: false
    }
  },
  mounted() {
    this.loadArticle()
  },
  methods: {
    async loadArticle() {
      const id = this.$route.params.id
      try {
        const res = await getArticle(id)
        this.article = res.data || res
        this.loadComments()
      } catch (e) {
        this.$message.error('文章加载失败')
      } finally {
        this.loading = false
      }
    },
    async loadComments() {
      const id = this.$route.params.id
      this.commentsLoading = true
      try {
        const res = await getComments(id)
        const d = res.data || res
        this.comments = Array.isArray(d) ? d : (d.list || d.items || [])
      } catch (e) {
        // 评论加载失败不影响主流程
      } finally {
        this.commentsLoading = false
      }
    },
    async handleLike() {
      const id = this.$route.params.id
      try {
        await likeArticle(id)
        if (this.article.liked) {
          this.article.likeCount = Math.max(0, (this.article.likeCount || 0) - 1)
        } else {
          this.article.likeCount = (this.article.likeCount || 0) + 1
        }
        this.article.liked = !this.article.liked
      } catch (e) {
        this.$message.error('操作失败')
      }
    },
    async handleCollect() {
      const id = this.$route.params.id
      try {
        await collectArticle(id)
        if (this.article.collected) {
          this.article.collectCount = Math.max(0, (this.article.collectCount || 0) - 1)
        } else {
          this.article.collectCount = (this.article.collectCount || 0) + 1
        }
        this.article.collected = !this.article.collected
        this.$message.success(this.article.collected ? '收藏成功' : '已取消收藏')
      } catch (e) {
        this.$message.error('操作失败')
      }
    },
    async submitComment() {
      if (!this.commentText.trim()) return
      const id = this.$route.params.id
      this.submittingComment = true
      try {
        await addComment(id, { content: this.commentText })
        this.$message.success('评论发表成功')
        this.commentText = ''
        this.loadComments()
        this.article.commentCount = (this.article.commentCount || 0) + 1
      } catch (e) {
        this.$message.error('评论发表失败')
      } finally {
        this.submittingComment = false
      }
    },
    formatTime(val) {
      if (!val) return ''
      const d = new Date(val)
      if (isNaN(d.getTime())) return val
      const pad = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
    }
  }
}
</script>

<style scoped>
.article-page { max-width: 800px; margin: 0 auto; padding: 20px 16px; }
.article-card { margin-bottom: 16px; }
.article-header { margin-bottom: 8px; }
.article-title { font-size: 22px; font-weight: bold; color: #222; margin: 0 0 12px; line-height: 1.4; }
.article-meta { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
.meta-tag { }
.meta-time { font-size: 12px; color: #bbb; }
.article-body { }
.action-row { display: flex; gap: 12px; flex-wrap: wrap; }
.comment-card { }
.card-header { display: flex; align-items: center; }
.card-title { font-size: 15px; font-weight: bold; color: #111; border-left: 4px solid #ff9900; padding-left: 10px; }
.comment-input-wrap { margin-bottom: 12px; }
.comment-actions { display: flex; justify-content: flex-end; margin-top: 8px; }
.comment-item { padding: 12px 0; border-bottom: 1px solid #f5f5f5; }
.comment-item:last-child { border-bottom: none; }
.comment-author { font-size: 13px; font-weight: bold; color: #333; margin-bottom: 4px; }
.comment-content { font-size: 14px; color: #444; line-height: 1.6; margin-bottom: 4px; }
.comment-time { font-size: 12px; color: #bbb; }
.empty-tip { text-align: center; color: #999; padding: 20px 0; font-size: 13px; }
</style>
