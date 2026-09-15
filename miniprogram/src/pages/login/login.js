const app = getApp()
const api = require('../../api/index')

Page({
  data: {
    loading: false,
    accountLoading: false,
    account: '',
    password: ''
  },

  // 微信登录
  handleWxLogin() {
    this.setData({ loading: true })
    
    // 调用微信登录
    wx.login({
      success: (res) => {
        if (res.code) {
          // 发送 code 到后端
          api.auth.wxLogin(res.code).then(result => {
            this.setData({ loading: false })
            
            // 保存登录信息
            app.globalData.token = result.data.token
            app.globalData.userInfo = result.data.user
            wx.setStorageSync('token', result.data.token)
            wx.setStorageSync('userInfo', result.data.user)
            
            wx.showToast({
              title: '登录成功',
              icon: 'success'
            })
            
            // 跳转回上一页或首页
            setTimeout(() => {
              wx.navigateBack()
            }, 1500)
          }).catch(err => {
            this.setData({ loading: false })
            wx.showToast({
              title: err.message || '登录失败',
              icon: 'none'
            })
          })
        } else {
          this.setData({ loading: false })
          wx.showToast({
            title: '获取微信授权失败',
            icon: 'none'
          })
        }
      },
      fail: () => {
        this.setData({ loading: false })
        wx.showToast({
          title: '微信登录失败',
          icon: 'none'
        })
      }
    })
  },

  // 账号密码登录
  onAccountInput(e) {
    this.setData({ account: e.detail.value })
  },

  onPasswordInput(e) {
    this.setData({ password: e.detail.value })
  },

  handleAccountLogin() {
    const { account, password } = this.data
    
    if (!account) {
      wx.showToast({ title: '请输入账号', icon: 'none' })
      return
    }
    if (!password) {
      wx.showToast({ title: '请输入密码', icon: 'none' })
      return
    }

    this.setData({ accountLoading: true })

    api.auth.login(account, password).then(result => {
      this.setData({ accountLoading: false })
      
      // 保存登录信息
      app.globalData.token = result.data.token
      app.globalData.userInfo = result.data.user
      wx.setStorageSync('token', result.data.token)
      wx.setStorageSync('userInfo', result.data.user)
      
      wx.showToast({
        title: '登录成功',
        icon: 'success'
      })
      
      setTimeout(() => {
        wx.navigateBack()
      }, 1500)
    }).catch(err => {
      this.setData({ accountLoading: false })
      wx.showToast({
        title: err.message || '登录失败',
        icon: 'none'
      })
    })
  },

  // 查看用户协议
  viewAgreement() {
    wx.showModal({
      title: '用户协议',
      content: '面试模拟系统用户协议\n\n1. 服务条款\n2. 用户权利与义务\n3. 隐私保护\n4. 免责声明',
      showCancel: false
    })
  },

  // 查看隐私政策
  viewPrivacy() {
    wx.showModal({
      title: '隐私政策',
      content: '我们重视您的隐私保护\n\n1. 信息收集\n2. 信息使用\n3. 信息存储\n4. Cookie政策',
      showCancel: false
    })
  },

  // 跳转到注册页
  goToRegister() {
    wx.navigateTo({
      url: '/pages/register/index'
    })
  }
})
