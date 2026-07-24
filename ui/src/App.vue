<template>
  <div id="app">
    <!-- 管理后台路由不显示主应用导航栏 -->
    <template v-if="!isAdminRoute">
      <template v-if="isLoggedIn">
        <Navbar />
        <Subnav />
      </template>
      <template v-else>
        <div class="navbar-guest">
          <router-link to="/" class="logo-guest">
            <svg width="28" height="28" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
              <rect x="2" y="3" width="22" height="16" rx="4" fill="#ff9900"/>
              <path d="M8 19 L6 25 L14 19Z" fill="#ff9900"/>
              <polyline points="7,11 11,15 19,8" stroke="#131921" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
              <rect x="18" y="17" width="12" height="9" rx="3" fill="#febd69"/>
              <line x1="21" y1="20" x2="27" y2="20" stroke="#131921" stroke-width="1.5" stroke-linecap="round"/>
              <line x1="21" y1="23" x2="25" y2="23" stroke="#131921" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            <span class="logo-guest-text">
              <span class="logo-guest-main">MockInterview</span>
              <span class="logo-guest-sub">AI 面试模拟</span>
            </span>
          </router-link>
          <router-link to="/login">
            <el-button type="warning" size="small">👤 登录/注册</el-button>
          </router-link>
        </div>
      </template>
    </template>
    <div :class="contentClass">
      <router-view />
    </div>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import Navbar from '@/components/layout/Navbar.vue'
import Subnav from '@/components/layout/Subnav.vue'

export default {
  name: 'App',
  components: { Navbar, Subnav },
  computed: {
    ...mapGetters('user', ['isLoggedIn']),
    isAdminRoute() {
      return this.$route.path.startsWith('/admin')
    },
    contentClass() {
      if (this.isAdminRoute) return 'app-content-admin'
      return this.isLoggedIn ? 'app-content' : 'app-content-guest'
    }
  }
}
</script>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: Arial, sans-serif; background: #f3f3f3; color: #111; }
.navbar-guest {
  background: #131921;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 20px;
  position: fixed; top: 0; left: 0; right: 0; z-index: 1000;
  height: 52px;
}
.navbar-guest .logo-guest {
  display: flex; align-items: center; gap: 8px; text-decoration: none;
}
.logo-guest-text { display: flex; flex-direction: column; line-height: 1.15; }
.logo-guest-main { font-size: 16px; font-weight: 800; color: #ff9900; }
.logo-guest-sub  { font-size: 10px; color: #aab7c4; letter-spacing: 1px; }
.app-content { margin-top: 90px; }
.app-content-guest { margin-top: 52px; }
.app-content-admin { margin-top: 0; }
</style>
