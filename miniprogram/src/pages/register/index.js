const api = require('../../api/index')

Page({
  data: {
    username: '',
    phone: '',
    email: '',
    password: '',
    confirmPassword: '',
    agreed: false,
    loading: false
  },

  // 验证手机号格式
  validatePhone(phone) {
    return /^1[3-9]\d{9}$/.test(phone)
  },

  // 验证邮箱格式
  validateEmail(email) {
    if (!email) return true // 邮箱选填
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
  },

  // 计算是否可以注册
  get canRegister() {
    const { username, phone, password, confirmPassword, agreed } = this.data
    return (
      username.trim().length >= 2 &&
      this.validatePhone(phone) &&
      this.validateEmail(this.data.email) &&
      password.length >= 6 &&
      password === confirmPassword &&
      agreed
    )
  },

  // 输入处理
  onUsernameInput(e) {
    this.setData({ username: e.detail.value })
  },

  onPhoneInput(e) {
    this.setData({ phone: e.detail.value })
  },

  onEmailInput(e) {
    this.setData({ email: e.detail.value })
  },

  onPasswordInput(e) {
    this.setData({ password: e.detail.value })
  },

  onConfirmPasswordInput(e) {
    this.setData({ confirmPassword: e.detail.value })
  },

  onAgreementChange(e) {
    this.setData({ agreed: e.detail.value.length > 0 })
  },

  // 查看用户协议
  viewAgreement() {
    wx.showModal({
      title: '用户协议',
      content: '这里是用户协议的内容...\n\n1. 服务条款\n2. 用户权利与义务\n3. 隐私保护\n4. 免责声明',
      showCancel: false
    })
  },

  // 查看隐私政策
  viewPrivacy() {
    wx.showModal({
      title: '隐私政策',
      content: '我们重视您的隐私保护...\n\n1. 信息收集\n2. 信息使用\n3. 信息共享\n4. 信息安全',
      showCancel: false
    })
  },

  // 注册
  handleRegister() {
    if (!this.canRegister) {
      const { username, phone, password, confirmPassword, agreed } = this.data
      
      if (!agreed) {
        wx.showToast({ title: '请阅读并同意用户协议', icon: 'none' })
        return
      }
      if (username.trim().length < 2) {
        wx.showToast({ title: '用户名至少2个字符', icon: 'none' })
        return
      }
      if (!this.validatePhone(phone)) {
        wx.showToast({ title: '请输入正确的手机号', icon: 'none' })
        return
      }
      if (password.length < 6) {
        wx.showToast({ title: '密码至少6位', icon: 'none' })
        return
      }
      if (password !== confirmPassword) {
        wx.showToast({ title: '两次密码不一致', icon: 'none' })
        return
      }
      return
    }

    this.setData({ loading: true })

    api.auth.register({
      username: this.data.username.trim(),
      phone: this.data.phone.trim(),
      email: this.data.email.trim(),
      password: this.data.password
    }).then(res => {
      this.setData({ loading: false })
      
      wx.showToast({
        title: '注册成功',
        icon: 'success'
      })

      // 自动登录
      setTimeout(() => {
        api.auth.login(this.data.username.trim(), this.data.password).then(loginRes => {
          const app = getApp()
          app.globalData.token = loginRes.data.token
          app.globalData.userInfo = loginRes.data.user
          wx.setStorageSync('token', loginRes.data.token)
          wx.setStorageSync('userInfo', loginRes.data.user)
          
          wx.navigateBack()
        }).catch(() => {
          wx.redirectTo({
            url: '/pages/login/login'
          })
        })
      }, 1500)
    }).catch(err => {
      this.setData({ loading: false })
      wx.showToast({
        title: err.message || '注册失败',
        icon: 'none'
      })
    })
  },

  // 跳转到登录页
  goToLogin() {
    wx.navigateBack()
  }
})
