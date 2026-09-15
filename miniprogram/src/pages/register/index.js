const api = require('../../api/index')
const store = require('../../store/index')

Page({
  data: {
    username: '',
    phone: '',
    email: '',
    password: '',
    confirmPassword: '',
    agreed: false,
    loading: false,
    // 注册按钮可用态（wxml 只能绑 data 字段，getters 在模板里取不到，必须入 data）
    canRegister: false
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

  // 计算是否可以注册（结果写入 data 供 wxml 绑定）
  updateCanRegister() {
    const { username, phone, email, password, confirmPassword, agreed } = this.data
    const canRegister = (
      username.trim().length >= 2 &&
      this.validatePhone(phone) &&
      this.validateEmail(email) &&
      password.length >= 8 &&
      password === confirmPassword &&
      agreed
    )
    if (canRegister !== this.data.canRegister) {
      this.setData({ canRegister })
    }
  },

  // 输入处理
  onUsernameInput(e) {
    this.setData({ username: e.detail.value })
    this.updateCanRegister()
  },

  onPhoneInput(e) {
    this.setData({ phone: e.detail.value })
    this.updateCanRegister()
  },

  onEmailInput(e) {
    this.setData({ email: e.detail.value })
    this.updateCanRegister()
  },

  onPasswordInput(e) {
    this.setData({ password: e.detail.value })
    this.updateCanRegister()
  },

  onConfirmPasswordInput(e) {
    this.setData({ confirmPassword: e.detail.value })
    this.updateCanRegister()
  },

  onAgreementChange(e) {
    this.setData({ agreed: e.detail.value.length > 0 })
    this.updateCanRegister()
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
    if (!this.data.canRegister) {
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
      if (password.length < 8) {
        wx.showToast({ title: '密码至少8位', icon: 'none' })
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
      // 后端账号字段：单一 account 或分开的 phone/email 都接受
      account: this.data.phone.trim(),
      phone: this.data.phone.trim(),
      email: this.data.email.trim(),
      password: this.data.password,
      confirmPassword: this.data.confirmPassword
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
          // 同步 store 内存态，保持会话内登录状态一致
          store.setToken(loginRes.data.token)
          store.setUser(loginRes.data.user)
          
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
