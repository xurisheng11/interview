const api = require('../../api/index')

Page({
  data: {
    loading: true,
    uploading: false,
    resumes: []
  },

  onLoad() {
    this.loadList()
  },

  onShow() {
    this.loadList()
  },

  onPullDownRefresh() {
    this.loadList().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  /** 加载简历列表 */
  loadList() {
    this.setData({ loading: true })
    return api.resume.list().then(res => {
      const list = res.data || []
      this.setData({
        loading: false,
        resumes: list.map(item => ({
          ...item,
          uploadedAtText: this.formatDate(item.uploadedAt),
          statusText: this.statusText(item.analysisStatus),
          statusClass: this.statusClass(item.analysisStatus),
          scoreClass: this.scoreClass(item.totalScore)
        }))
      })
    }).catch(err => {
      console.log('加载简历列表失败', err)
      this.setData({ loading: false, resumes: [] })
    })
  },

  /** 选择并上传简历 */
  handleUpload() {
    if (this.data.uploading) return
    wx.chooseMessageFile({
      count: 1,
      type: 'file',
      extension: ['.pdf', '.doc', '.docx'],
      success: (res) => {
        const file = res.tempFiles[0]
        // 校验文件大小（10 MB）
        if (file.size > 10 * 1024 * 1024) {
          wx.showToast({ title: '文件大小不能超过 10 MB', icon: 'none' })
          return
        }
        // 校验扩展名
        const name = (file.name || '').toLowerCase()
        if (!name.endsWith('.pdf') && !name.endsWith('.doc') && !name.endsWith('.docx')) {
          wx.showToast({ title: '只支持 PDF 或 Word 文件', icon: 'none' })
          return
        }
        this.doUpload(file.path)
      }
    })
  },

  /** 执行上传 */
  doUpload(filePath) {
    this.setData({ uploading: true })
    wx.showLoading({ title: '上传中…', mask: true })
    api.resume.upload(filePath).then(res => {
      wx.hideLoading()
      this.setData({ uploading: false })
      const id = (res.data && res.data.id) || res.id
      wx.showToast({ title: '上传成功，正在分析', icon: 'success' })
      // 跳转详情页
      if (id) {
        wx.navigateTo({ url: '/pages/resume/detail?id=' + id })
      } else {
        this.loadList()
      }
    }).catch(err => {
      wx.hideLoading()
      this.setData({ uploading: false })
      wx.showToast({ title: err.message || '上传失败', icon: 'none' })
    })
  },

  /** 点击简历卡片 → 跳转详情 */
  goDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: '/pages/resume/detail?id=' + id })
  },

  /** 长按删除 */
  handleLongPress(e) {
    const id = e.currentTarget.dataset.id
    const filename = e.currentTarget.dataset.filename || '该简历'
    wx.showModal({
      title: '确认删除',
      content: '确定要删除简历「' + filename + '」吗？此操作不可恢复。',
      confirmColor: '#ff4d4f',
      confirmText: '删除',
      success: (res) => {
        if (res.confirm) {
          this.doDelete(id)
        }
      }
    })
  },

  /** 执行删除 */
  doDelete(id) {
    wx.showLoading({ title: '删除中…' })
    api.resume.delete(id).then(() => {
      wx.hideLoading()
      wx.showToast({ title: '删除成功', icon: 'success' })
      this.loadList()
    }).catch(err => {
      wx.hideLoading()
      wx.showToast({ title: '删除失败', icon: 'none' })
    })
  },

  /** 日期格式化 */
  formatDate(val) {
    if (!val) return '—'
    var d = new Date(val)
    var p = function (n) { return String(n).padStart(2, '0') }
    return d.getFullYear() + '-' + p(d.getMonth() + 1) + '-' + p(d.getDate())
  },

  /** 状态文案 */
  statusText(s) {
    var map = { pending: '待分析', analyzing: '分析中', done: '已分析', failed: '分析失败' }
    return map[s] || s || '未知'
  },

  /** 状态样式类 */
  statusClass(s) {
    var map = { pending: 'pending', analyzing: 'analyzing', done: 'done', failed: 'failed' }
    return map[s] || 'pending'
  },

  /** 评分样式类 */
  scoreClass(score) {
    if (score == null) return 'score-none'
    if (score >= 80) return 'score-high'
    if (score >= 60) return 'score-mid'
    return 'score-low'
  }
})
