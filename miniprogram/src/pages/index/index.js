const api = require('../../api/index')
const store = require('../../store/index')

const STATUS_NAMES = { ongoing: '进行中', paused: '已暂停', completed: '已完成' }

// ISO 时间（带 T/Z）→ 本地可读格式
function fmtTime(val) {
  if (!val) return '--'
  const d = new Date(val)
  if (!isNaN(d.getTime())) {
    const pad = n => String(n).padStart(2, '0')
    return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) +
      ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes())
  }
  const m = String(val).match(/^(\d{4}-\d{2}-\d{2})[T ](\d{2}:\d{2})/)
  return m ? m[1] + ' ' + m[2] : String(val)
}

function scoreClass(score) {
  if (score >= 90) return 's-exc'
  if (score >= 75) return 's-good'
  if (score >= 60) return 's-pass'
  return 's-low'
}

Page({
  data: {
    isLoggedIn: false,
    userInfo: null,
    // 自定义顶栏状态栏高度
    statusBarHeight: 20,
    // 门禁重定向进行中的防抖标记，避免 onLoad/onShow 重复 reLaunch
    loginRedirecting: false,
    stats: null,
    recentInterviews: []
  },

  onLoad() {
    try {
      const info = wx.getWindowInfo ? wx.getWindowInfo() : wx.getSystemInfoSync()
      this.setData({ statusBarHeight: info.statusBarHeight || 20 })
    } catch (e) { /* 保持默认 */ }
    this.checkLoginStatus()
    this.gateLogin()
  },

  onShow() {
    this.checkLoginStatus()
    // 门禁未拦截且已登录才拉数据
    if (!this.gateLogin() && store.isLoggedIn()) {
      this.loadData()
    }
  },

  // 登录门禁：未登录一律 reLaunch 到登录页（无 tabBar），防未登录浏览 tab 内容
  gateLogin() {
    if (this.data.loginRedirecting) return true
    if (store.isLoggedIn()) return false
    this.setData({ loginRedirecting: true })
    wx.reLaunch({ url: '/pages/login/login' })
    return true
  },

  checkLoginStatus() {
    const userInfo = wx.getStorageSync('userInfo')
    const token = wx.getStorageSync('token')
    
    this.setData({
      isLoggedIn: !!token,
      userInfo: userInfo
    })
  },

  loadData() {
    // 加载统计数据
    api.profile.getStats().then(res => {
      this.setData({ stats: res.data })
    }).catch(err => {
      console.log('加载统计数据失败', err)
    })

    // 加载最近面试（预计算展示字段，wxml 不支持函数调用）
    api.interview.list().then(res => {
      const list = (res.data || []).slice(0, 5).map(it => Object.assign({}, it, {
        _time: fmtTime(it.startTime),
        _statusText: STATUS_NAMES[it.status] || it.status || '',
        _isDone: it.status === 'completed',
        _scoreCls: scoreClass(it.totalScore || 0),
        _modeText: it.mode === 'video' ? '视频' : '文字'
      }))
      this.setData({ recentInterviews: list })
    }).catch(err => {
      console.log('加载面试记录失败', err)
    })
  },

  goLogin() {
    wx.navigateTo({
      url: '/pages/login/login'
    })
  },

  goAccountLogin() {
    // 跳转到登录页使用账号密码登录
    wx.navigateTo({
      url: '/pages/login/login'
    })
  },

  goRegister() {
    wx.navigateTo({
      url: '/pages/register/index'
    })
  },

  goAgreement() {
    wx.showModal({
      title: '用户协议',
      content: '面试模拟系统用户协议\n\n1. 服务条款\n2. 用户权利与义务\n3. 隐私保护\n4. 免责声明',
      showCancel: false
    })
  },

  goPrivacy() {
    wx.showModal({
      title: '隐私政策',
      content: '我们重视您的隐私保护\n\n1. 信息收集\n2. 信息使用\n3. 信息存储\n4. Cookie政策',
      showCancel: false
    })
  },

  goToInterview() {
    wx.navigateTo({
      url: '/pages/interview/create'
    })
  },

  goToPractice() {
    // 跳转到题库页面
    wx.navigateTo({
      url: '/pages/questions/index'
    })
  },

  goToReport() {
    // 面试列表是 tabBar 页面，必须用 switchTab
    wx.switchTab({
      url: '/pages/interview/list'
    })
  },

  goToInterviewDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/interview/detail?id=${id}`
    })
  },

  // ===== 功能导航 =====

  goToLearn() {
    wx.navigateTo({ url: '/pages/learn/index' })
  },

  goToResume() {
    wx.navigateTo({ url: '/pages/resume/list' })
  },

  goToCommunity() {
    wx.navigateTo({ url: '/pages/community/index' })
  },

  goToCompanyIntel() {
    wx.navigateTo({ url: '/pages/company/intel' })
  },

  goToContributions() {
    wx.navigateTo({ url: '/pages/contribution/company' })
  },

  goToMyContributions() {
    wx.navigateTo({ url: '/pages/contribution/my' })
  },

  goToCollections() {
    wx.navigateTo({ url: '/pages/collections/index' })
  },

  goToProfile() {
    wx.navigateTo({ url: '/pages/profile/index' })
  },

  goToHelp() {
    wx.navigateTo({ url: '/pages/help/index' })
  }
})
