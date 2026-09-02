<template>
  <div class="resume-list-page">
    <div class="page-header">
      <h2 class="page-title">📄 我的简历</h2>
      <el-button type="primary" size="small" @click="showUpload = true">
        <i class="el-icon-upload2"></i> 上传简历
      </el-button>
    </div>

    <!-- 简历列表 -->
    <el-card shadow="never" class="list-card">
      <div v-loading="loading">
        <!-- 空状态 -->
        <div v-if="!loading && resumes.length === 0" class="empty-state">
          <div class="empty-icon">📄</div>
          <div class="empty-title">还没有上传过简历</div>
          <div class="empty-sub">上传简历后，AI 将自动分析并为你生成针对性面试题目</div>
          <el-button type="primary" size="default" @click="showUpload = true">
            上传第一份简历
          </el-button>
        </div>

        <!-- 简历列表 -->
        <div v-else class="resume-grid">
          <div
            v-for="r in resumes"
            :key="r.id"
            class="resume-card"
            @click="goDetail(r.id)"
          >
            <div class="resume-card-top">
              <div class="resume-icon">📄</div>
              <div class="resume-info">
                <div class="resume-filename">{{ r.filename }}</div>
                <div class="resume-date">{{ formatDate(r.uploadedAt) }}</div>
              </div>
              <el-tag size="mini" :type="statusTagType(r.analysisStatus)" class="resume-status-tag">
                {{ statusLabel(r.analysisStatus) }}
              </el-tag>
            </div>

            <div class="resume-card-bottom">
              <div v-if="r.totalScore != null" class="resume-score">
                <span class="score-label">AI 评分</span>
                <span class="score-value" :class="scoreClass(r.totalScore)">{{ r.totalScore }}</span>
              </div>
              <div v-else class="resume-score">
                <span class="score-label text-muted">AI 评分</span>
                <span class="score-value text-muted">—</span>
              </div>
              <div class="resume-actions" @click.stop>
                <el-button type="text" size="mini" @click="goDetail(r.id)">查看详情</el-button>
                <el-button type="text" size="mini" class="delete-btn" @click="handleDelete(r)">删除</el-button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 上传弹窗 -->
    <el-dialog title="上传简历" :visible.sync="showUpload" width="480px" :close-on-click-modal="false">
      <div class="upload-tip">支持 PDF、Word 文档，最大 10 MB</div>
      <el-upload
        ref="upload"
        class="resume-upload"
        drag
        action="#"
        :auto-upload="false"
        :limit="1"
        :on-change="handleFileChange"
        :on-remove="handleFileRemove"
        accept=".pdf,.doc,.docx"
      >
        <i class="el-icon-upload"></i>
        <div class="el-upload__text">将文件拖到此处，或 <em>点击选择</em></div>
        <div slot="tip" class="el-upload__tip">只能上传 PDF / DOC / DOCX 文件，且不超过 10MB</div>
      </el-upload>
      <div v-if="uploadError" class="upload-error">{{ uploadError }}</div>
      <div slot="footer" class="dialog-footer">
        <el-button @click="showUpload = false" :disabled="uploading">取消</el-button>
        <el-button type="primary" :loading="uploading" :disabled="!selectedFile" @click="handleUpload">
          {{ uploading ? '解析中…' : '上传并解析' }}
        </el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { uploadResume, getResumeList, deleteResume } from '@/api/resume'

export default {
  name: 'ResumeList',
  data() {
    return {
      loading: true,
      resumes: [],
      showUpload: false,
      selectedFile: null,
      uploadError: '',
      uploading: false,
      fileRaw: null
    }
  },
  created() {
    this.load()
  },
  methods: {
    async load() {
      this.loading = true
      try {
        const res = await getResumeList()
        this.resumes = res.data || res || []
      } catch (e) {
        // 错误已在拦截器处理
      }
      this.loading = false
    },
    handleFileChange(file, fileList) {
      this.uploadError = ''
      // 扩展名校验
      const name = file.name.toLowerCase()
      if (!name.endsWith('.pdf') && !name.endsWith('.doc') && !name.endsWith('.docx')) {
        this.uploadError = '只支持 PDF 或 Word 文件（.pdf / .doc / .docx）'
        this.$refs.upload.clearFiles()
        this.selectedFile = null
        return
      }
      // 大小校验
      if (file.size > 10 * 1024 * 1024) {
        this.uploadError = '文件大小不能超过 10 MB'
        this.$refs.upload.clearFiles()
        this.selectedFile = null
        return
      }
      this.selectedFile = file.raw
    },
    handleFileRemove() {
      this.selectedFile = null
      this.uploadError = ''
    },
    async handleUpload() {
      if (!this.selectedFile) return
      this.uploading = true
      this.uploadError = ''
      try {
        const fd = new FormData()
        fd.append('file', this.selectedFile)
        const res = await uploadResume(fd)
        const id = res.data?.id || res.id
        this.showUpload = false
        this.$message.success('简历上传成功，正在分析中…')
        this.$nextTick(() => { this.$router.push(`/resume/${id}`) })
      } catch (e) {
        this.uploadError = e.message || '上传失败，请重试'
      }
      this.uploading = false
    },
    async handleDelete(r) {
      try {
        await this.$confirm(`确定要删除简历「${r.filename}」吗？此操作不可恢复。`, '确认删除', {
          confirmButtonText: '删除',
          cancelButtonText: '取消',
          type: 'warning'
        })
        await deleteResume(r.id)
        this.resumes = this.resumes.filter(x => x.id !== r.id)
        this.$message.success('删除成功')
      } catch (e) {
        if (e !== 'cancel') this.$message.error('删除失败')
      }
    },
    goDetail(id) { this.$router.push(`/resume/${id}`) },
    formatDate(val) {
      if (!val) return '—'
      const d = new Date(val)
      const p = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${p(d.getMonth()+1)}-${p(d.getDate())}`
    },
    statusTagType(s) {
      return { pending: 'info', analyzing: 'warning', done: 'success', failed: 'danger' }[s] || 'info'
    },
    statusLabel(s) {
      return { pending: '待分析', analyzing: '分析中', done: '已分析', failed: '分析失败' }[s] || s || '未知'
    },
    scoreClass(s) {
      if (s == null) return ''
      if (s >= 80) return 'score-high'
      if (s >= 60) return 'score-mid'
      return 'score-low'
    }
  }
}
</script>

<style scoped>
.resume-list-page { padding: 20px; max-width: 960px; margin: 0 auto; }

.page-header {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 16px;
}
.page-title { font-size: 20px; font-weight: bold; color: #111; }

.list-card { min-height: 300px; }

/* 空状态 */
.empty-state {
  text-align: center; padding: 60px 20px; color: #999;
}
.empty-icon { font-size: 56px; margin-bottom: 16px; }
.empty-title { font-size: 16px; font-weight: bold; color: #555; margin-bottom: 8px; }
.empty-sub { font-size: 13px; margin-bottom: 20px; color: #aaa; }

/* 简历卡片 */
.resume-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 14px; }
.resume-card {
  border: 1px solid #eee; border-radius: 10px; padding: 16px;
  cursor: pointer; transition: all 0.2s; background: #fff;
}
.resume-card:hover {
  border-color: #ff9900; box-shadow: 0 3px 14px rgba(255,153,0,0.15); transform: translateY(-2px);
}
.resume-card-top { display: flex; align-items: flex-start; gap: 12px; margin-bottom: 14px; }
.resume-icon { font-size: 32px; flex-shrink: 0; }
.resume-info { flex: 1; min-width: 0; }
.resume-filename {
  font-size: 14px; font-weight: 600; color: #111;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin-bottom: 4px;
}
.resume-date { font-size: 12px; color: #aaa; }
.resume-status-tag { flex-shrink: 0; }

.resume-card-bottom { display: flex; align-items: center; justify-content: space-between; border-top: 1px solid #f5f5f5; padding-top: 12px; }
.resume-score { display: flex; align-items: baseline; gap: 6px; }
.score-label { font-size: 12px; color: #888; }
.score-value { font-size: 22px; font-weight: bold; }
.score-high { color: #067d62; }
.score-mid  { color: #ff9900; }
.score-low  { color: #c7511f; }
.text-muted { color: #ccc; }

.resume-actions { display: flex; gap: 4px; }
.delete-btn { color: #f56c6c !important; }

/* 上传弹窗 */
.upload-tip { font-size: 13px; color: #888; margin-bottom: 12px; }
.resume-upload { width: 100%; }
.upload-error { color: #f56c6c; font-size: 13px; margin-top: 8px; }
.dialog-footer { display: flex; justify-content: flex-end; gap: 10px; }

::v-deep .el-card { border: 1px solid #eee; }
::v-deep .el-button--primary { background: #ff9900; border-color: #ff9900; color: #111; }
::v-deep .el-button--primary:hover { background: #f3a847; border-color: #f3a847; }
</style>
