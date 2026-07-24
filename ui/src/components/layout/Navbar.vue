<template>
  <div class="navbar">
    <div class="navbar-left">
      <router-link to="/dashboard" class="logo">
        <!-- SVG Logo：对话气泡 + 对勾，寓意面试问答与通过 -->
        <svg width="32" height="32" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg" class="logo-svg">
          <!-- 主气泡 -->
          <rect x="2" y="3" width="22" height="16" rx="4" fill="#ff9900"/>
          <!-- 气泡尾 -->
          <path d="M8 19 L6 25 L14 19Z" fill="#ff9900"/>
          <!-- 对勾 -->
          <polyline points="7,11 11,15 19,8" stroke="#131921" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
          <!-- 小气泡（右下，表示回应） -->
          <rect x="18" y="17" width="12" height="9" rx="3" fill="#febd69"/>
          <!-- 小气泡三条横线 -->
          <line x1="21" y1="20" x2="27" y2="20" stroke="#131921" stroke-width="1.5" stroke-linecap="round"/>
          <line x1="21" y1="23" x2="25" y2="23" stroke="#131921" stroke-width="1.5" stroke-linecap="round"/>
        </svg>
        <span class="logo-text">
          <span class="logo-main">MockInterview</span>
          <span class="logo-sub">AI 面试模拟</span>
        </span>
      </router-link>
    </div>
    <div class="navbar-right">
      <el-dropdown @command="handleCommand">
        <span class="user-info">
          <el-avatar :size="32" :src="userInfo && userInfo.avatar">
            {{ userInfo && userInfo.nickname ? userInfo.nickname[0] : '用' }}
          </el-avatar>
          <span class="username">{{ userInfo && (userInfo.nickname || userInfo.username) }}</span>
          <i class="el-icon-arrow-down"></i>
        </span>
        <el-dropdown-menu slot="dropdown">
          <el-dropdown-item command="profile">个人中心</el-dropdown-item>
          <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
        </el-dropdown-menu>
      </el-dropdown>
    </div>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'

export default {
  name: 'Navbar',
  computed: {
    ...mapGetters('user', ['userInfo'])
  },
  methods: {
    handleCommand(cmd) {
      if (cmd === 'logout') {
        this.$store.dispatch('user/logout')
        this.$router.push('/login')
      } else if (cmd === 'profile') {
        this.$router.push('/profile')
      }
    }
  }
}
</script>

<style scoped>
.navbar {
  background: #131921;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  position: fixed;
  top: 0; left: 0; right: 0;
  z-index: 1000;
  height: 52px;
}
.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
}
.logo-svg {
  flex-shrink: 0;
  filter: drop-shadow(0 1px 3px rgba(255,153,0,0.4));
}
.logo-text {
  display: flex;
  flex-direction: column;
  line-height: 1.15;
}
.logo-main {
  font-size: 17px;
  font-weight: 800;
  color: #ff9900;
  letter-spacing: 0.3px;
}
.logo-sub {
  font-size: 10px;
  color: #aab7c4;
  font-weight: 400;
  letter-spacing: 1px;
}
.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #fff;
}
.username { font-size: 14px; }
</style>
