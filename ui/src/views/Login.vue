<template>
  <div class="login-page">
    <div class="login-wrap">
      <!-- 标题 -->
      <div class="login-logo">面试模拟系统</div>

      <el-card class="login-card" shadow="always">
        <!-- Tab 切换 -->
        <div class="login-tabs">
          <div
            class="login-tab"
            :class="{ active: activeTab === 'login' }"
            @click="switchTab('login')"
          >登录</div>
          <div
            class="login-tab"
            :class="{ active: activeTab === 'register' }"
            @click="switchTab('register')"
          >注册</div>
        </div>

        <!-- 登录表单 -->
        <el-form
          v-if="activeTab === 'login'"
          ref="loginForm"
          :model="loginForm"
          :rules="loginRules"
          label-position="top"
          @submit.native.prevent="handleLogin"
        >
          <el-form-item label="账号（手机号 / 邮箱）" prop="account">
            <el-input
              v-model="loginForm.account"
              placeholder="请输入手机号或邮箱"
              prefix-icon="el-icon-user"
              clearable
            />
          </el-form-item>

          <el-form-item label="密码" prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="请输入密码"
              prefix-icon="el-icon-lock"
              show-password
              clearable
            />
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              native-type="submit"
              class="submit-btn"
              :loading="loginLoading"
              @click="handleLogin"
            >登录</el-button>
          </el-form-item>
        </el-form>

        <!-- 注册表单 -->
        <el-form
          v-if="activeTab === 'register'"
          ref="registerForm"
          :model="registerForm"
          :rules="registerRules"
          label-position="top"
          @submit.native.prevent="handleRegister"
        >
          <el-form-item label="用户名" prop="username">
            <el-input
              v-model="registerForm.username"
              placeholder="请输入用户名"
              prefix-icon="el-icon-user"
              clearable
            />
          </el-form-item>

          <el-form-item label="账号（手机号 / 邮箱）" prop="account">
            <el-input
              v-model="registerForm.account"
              placeholder="请输入手机号或邮箱"
              prefix-icon="el-icon-message"
              clearable
            />
          </el-form-item>

          <el-form-item label="密码" prop="password">
            <el-input
              v-model="registerForm.password"
              type="password"
              placeholder="至少 8 位"
              prefix-icon="el-icon-lock"
              show-password
              clearable
            />
          </el-form-item>

          <el-form-item label="确认密码" prop="confirmPassword">
            <el-input
              v-model="registerForm.confirmPassword"
              type="password"
              placeholder="请再次输入密码"
              prefix-icon="el-icon-lock"
              show-password
              clearable
            />
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              native-type="submit"
              class="submit-btn"
              :loading="registerLoading"
              @click="handleRegister"
            >注册</el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script>
import { register } from '@/api/auth'

export default {
  name: 'Login',

  data() {
    // 确认密码校验器
    const validateConfirmPassword = (rule, value, callback) => {
      if (!value) {
        callback(new Error('请再次输入密码'))
      } else if (value !== this.registerForm.password) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    }

    // 账号格式校验（手机号或邮箱）
    const validateAccount = (rule, value, callback) => {
      if (!value) {
        callback(new Error('请输入手机号或邮箱'))
        return
      }
      const phoneReg = /^1[3-9]\d{9}$/
      const emailReg = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      if (!phoneReg.test(value) && !emailReg.test(value)) {
        callback(new Error('请输入有效的手机号或邮箱'))
      } else {
        callback()
      }
    }

    return {
      activeTab: 'login',

      // 登录表单
      loginForm: {
        account: '',
        password: ''
      },
      loginRules: {
        account: [
          { required: true, message: '请输入账号', trigger: 'blur' }
        ],
        password: [
          { required: true, message: '请输入密码', trigger: 'blur' }
        ]
      },
      loginLoading: false,

      // 注册表单
      registerForm: {
        username: '',
        account: '',
        password: '',
        confirmPassword: ''
      },
      registerRules: {
        username: [
          { required: true, message: '请输入用户名', trigger: 'blur' },
          { min: 2, max: 20, message: '用户名长度 2~20 个字符', trigger: 'blur' }
        ],
        account: [
          { required: true, validator: validateAccount, trigger: 'blur' }
        ],
        password: [
          { required: true, message: '请输入密码', trigger: 'blur' },
          { min: 8, message: '密码至少 8 位', trigger: 'blur' }
        ],
        confirmPassword: [
          { required: true, validator: validateConfirmPassword, trigger: 'blur' }
        ]
      },
      registerLoading: false
    }
  },

  methods: {
    switchTab(tab) {
      this.activeTab = tab
      // 切换时重置表单校验状态
      this.$nextTick(() => {
        const formRef = tab === 'login' ? this.$refs.loginForm : this.$refs.registerForm
        if (formRef) formRef.clearValidate()
      })
    },

    handleLogin() {
      this.$refs.loginForm.validate(async valid => {
        if (!valid) return
        this.loginLoading = true
        try {
          await this.$store.dispatch('user/login', {
            account: this.loginForm.account,
            password: this.loginForm.password
          })
          const redirect = this.$route.query.redirect
          // 避免跳回 /login 自身造成循环
          if (redirect && redirect !== '/login' && !redirect.startsWith('/login')) {
            this.$router.replace(redirect)
          } else {
            this.$router.replace('/dashboard')
          }
        } catch (err) {
          const msg = err?.response?.data?.message || err?.message || '登录失败，请检查账号或密码'
          this.$message.error(msg)
        } finally {
          this.loginLoading = false
        }
      })
    },

    handleRegister() {
      this.$refs.registerForm.validate(async valid => {
        if (!valid) return
        this.registerLoading = true
        try {
          await register({
            username: this.registerForm.username,
            account: this.registerForm.account,
            password: this.registerForm.password,
            confirmPassword: this.registerForm.confirmPassword
          })
          // 注册成功后自动登录
          await this.$store.dispatch('user/login', {
            account: this.registerForm.account,
            password: this.registerForm.password
          })
          this.$message.success('注册成功，欢迎使用！')
          this.$router.replace('/dashboard')
        } catch (err) {
          const msg = err?.response?.data?.message || err?.message || '注册失败，请稍后重试'
          this.$message.error(msg)
        } finally {
          this.registerLoading = false
        }
      })
    }
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: #232f3e;
  display: flex;
  align-items: center;
  justify-content: center;
}

.login-wrap {
  width: 420px;
}

.login-logo {
  text-align: center;
  font-size: 28px;
  font-weight: bold;
  color: #ff9900;
  margin-bottom: 24px;
  letter-spacing: 2px;
}

.login-card {
  border-radius: 8px;
  padding: 8px 16px 0;
}

/* Tab 切换 */
.login-tabs {
  display: flex;
  border-bottom: 2px solid #ff9900;
  margin-bottom: 24px;
}

.login-tab {
  flex: 1;
  text-align: center;
  padding: 10px;
  cursor: pointer;
  font-weight: bold;
  font-size: 15px;
  color: #888;
  transition: color 0.2s;
  user-select: none;
}

.login-tab.active {
  color: #ff9900;
  border-bottom: 3px solid #ff9900;
  margin-bottom: -2px;
}

.login-tab:hover:not(.active) {
  color: #333;
}

/* 提交按钮 */
.submit-btn {
  width: 100%;
  background: #ff9900;
  border-color: #ff9900;
  color: #111;
  font-weight: bold;
  font-size: 15px;
  height: 42px;
  border-radius: 4px;
  transition: background 0.2s;
}

.submit-btn:hover,
.submit-btn:focus {
  background: #f3a847;
  border-color: #f3a847;
  color: #111;
}

/* Element UI 覆盖 */
::v-deep .el-form-item__label {
  font-weight: bold;
  color: #333;
  line-height: 1.4;
  padding-bottom: 4px;
}

::v-deep .el-input__inner:focus {
  border-color: #ff9900;
  box-shadow: 0 0 0 2px rgba(255, 153, 0, 0.2);
}

::v-deep .el-button--primary {
  background: #ff9900;
  border-color: #ff9900;
  color: #111;
}

::v-deep .el-button--primary:hover,
::v-deep .el-button--primary:focus {
  background: #f3a847;
  border-color: #f3a847;
  color: #111;
}
</style>
