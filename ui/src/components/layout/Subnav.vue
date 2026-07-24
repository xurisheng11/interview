<template>
  <div class="subnav">
    <router-link to="/dashboard" class="subnav-item" active-class="active">首页</router-link>
    <router-link to="/interview/config" class="subnav-item" active-class="active">开始面试</router-link>
    <router-link to="/interview/history" class="subnav-item" active-class="active">面试历史</router-link>
    <router-link to="/questions" class="subnav-item" active-class="active">题库</router-link>
    <router-link to="/community" class="subnav-item" active-class="active">知识社区</router-link>
    <router-link to="/company/intel" class="subnav-item" active-class="active">🏢 公司面试知识库</router-link>
    <router-link to="/profile" class="subnav-item" active-class="active">个人中心</router-link>
  </div>
</template>

<script>
export default {
  name: 'Subnav'
}
</script>

<style scoped>
.subnav {
  background: #232f3e;
  display: flex;
  align-items: center;
  padding: 0 20px;
  position: fixed;
  top: 52px; left: 0; right: 0;
  z-index: 999;
  height: 38px;
}
.subnav-item {
  color: #ccc;
  text-decoration: none;
  font-size: 13px;
  padding: 0 12px;
  height: 38px;
  line-height: 38px;
  display: inline-block;
  transition: color 0.2s, background 0.2s;
}
.subnav-item:hover,
.subnav-item.active {
  color: #ff9900;
  background: rgba(255, 153, 0, 0.08);
}
</style>
