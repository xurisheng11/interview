<template>
  <div class="admin-login-page">
    <div class="admin-login-card">
      <div class="admin-login-header">
        <svg width="36" height="36" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
          <rect x="2" y="3" width="22" height="16" rx="4" fill="#ff9900"/>
          <path d="M8 19 L6 25 L14 19Z" fill="#ff9900"/>
          <polyline points="7,11 11,15 19,8" stroke="#131921" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
          <rect x="18" y="17" width="12" height="9" rx="3" fill="#febd69"/>
          <line x1="21" y1="20" x2="27" y2="20" stroke="#131921" stroke-width="1.5" stroke-linecap="round"/>
          <line x1="21" y1="23" x2="25" y2="23" stroke="#131921" stroke-width="1.5" stroke-linecap="round"/>
        </svg>
        <div class="admin-login-title">
          <span class="admin-login-main">MockInterview</span>
          <span class="admin-login-sub">后台管理系统</span>
        </div>
      </div>

      <el-form ref="loginForm" :model="form" :rules="rules" @submit.native.prevent="handleLogin">
        <el-form-item prop="account">
          <el-input
            v-model="form.account"
            placeholder="管理员用户名"
            prefix-icon="el-icon-user"
            size="medium"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            prefix-icon="el-icon-lock"
            size="medium"
            show-password
            @keyup.enter.native="handleLogin"
          />
        </el-form-item>
        <el-button
          type="primary"
          class="admin-login-btn"
          :loading="loading"
          @click="handleLogin"
        >
          登 录
        </el-button>
      </el-form>
    </div>
  </div>
</template>

<script>
import { adminLogin } from '@/api/admin'

export default {
  name: 'AdminLogin',
  data() {
    return {
      loading: false,
      form: {
        account: '',
        password: ''
      },
      rules: {
        account: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
        password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
      }
    }
  },
  methods: {
    handleLogin() {
      this.$refs.loginForm.validate(async valid => {
        if (!valid) return
        this.loading = true
        try {
          const res = await adminLogin({ account: this.form.account, password: this.form.password })
          const user = res.data.user
          if (user.role !== 'admin') {
            this.$message.error('权限不足，请使用管理员账号登录')
            return
          }
          // 只存 admin 专用 key，不覆盖主应用 token
          localStorage.setItem('admin_token', res.data.token)
          localStorage.setItem('admin_user', JSON.stringify(user))
          this.$message.success('登录成功')
          this.$router.push('/admin/dashboard')
        } catch (e) {
          // 错误由 request 拦截器统一处理
        } finally {
          this.loading = false
        }
      })
    }
  }
}
</script>

<style scoped>
.admin-login-page {
  min-height: 100vh;
  background: #131921;
  display: flex;
  align-items: center;
  justify-content: center;
}
.admin-login-card {
  background: #fff;
  border-radius: 10px;
  padding: 40px 36px 32px;
  width: 360px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.3);
}
.admin-login-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 28px;
}
.admin-login-title {
  display: flex;
  flex-direction: column;
}
.admin-login-main {
  font-size: 20px;
  font-weight: 800;
  color: #131921;
  line-height: 1.2;
}
.admin-login-sub {
  font-size: 12px;
  color: #888;
  letter-spacing: 1px;
}
.admin-login-btn {
  width: 100%;
  margin-top: 8px;
  background: #ff9900;
  border-color: #ff9900;
  color: #131921;
  font-weight: 700;
  letter-spacing: 2px;
}
.admin-login-btn:hover, .admin-login-btn:focus {
  background: #e68a00;
  border-color: #e68a00;
  color: #131921;
}
</style>
