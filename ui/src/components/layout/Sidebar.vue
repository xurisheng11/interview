<template>
  <div class="sidebar">
    <div v-for="(group, gi) in items" :key="gi">
      <div v-if="group.title" class="sidebar-group-title">{{ group.title }}</div>
      <div
        v-for="(item, ii) in group.children"
        :key="gi + '-' + ii"
        class="sidebar-item"
        :class="{ active: isActive(item.path) }"
        @click="navigate(item.path)"
      >
        <span class="sidebar-icon">{{ item.icon }}</span>
        {{ item.label }}
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Sidebar',
  props: {
    // items: [{ title: '分组名', children: [{ icon, label, path }] }]
    items: {
      type: Array,
      default: () => []
    }
  },
  methods: {
    isActive(path) {
      return this.$route.path === path || this.$route.path.startsWith(path + '/')
    },
    navigate(path) {
      if (this.$route.path !== path) {
        this.$router.push(path)
      }
    }
  }
}
</script>

<style scoped>
.sidebar {
  width: 200px;
  background: #fff;
  border-right: 1px solid #ddd;
  flex-shrink: 0;
  padding: 10px 0;
  min-height: calc(100vh - 90px);
}
.sidebar-group-title {
  padding: 8px 20px 4px;
  font-size: 11px;
  color: #999;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.sidebar-item {
  padding: 10px 20px;
  font-size: 14px;
  cursor: pointer;
  color: #333;
  transition: background 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}
.sidebar-item:hover { background: #f5f5f5; }
.sidebar-item.active {
  background: #fff8ee;
  color: #ff9900;
  font-weight: bold;
  border-left: 3px solid #ff9900;
  padding-left: 17px;
}
.sidebar-icon { font-size: 16px; }
</style>
