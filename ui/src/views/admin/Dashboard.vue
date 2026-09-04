<template>
  <div class="admin-layout">
    <!-- 顶栏 -->
    <header class="admin-header">
      <div class="admin-header-left">
        <svg width="28" height="28" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
          <rect x="2" y="3" width="22" height="16" rx="4" fill="#ff9900"/>
          <path d="M8 19 L6 25 L14 19Z" fill="#ff9900"/>
          <polyline points="7,11 11,15 19,8" stroke="#131921" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
          <rect x="18" y="17" width="12" height="9" rx="3" fill="#febd69"/>
          <line x1="21" y1="20" x2="27" y2="20" stroke="#131921" stroke-width="1.5" stroke-linecap="round"/>
          <line x1="21" y1="23" x2="25" y2="23" stroke="#131921" stroke-width="1.5" stroke-linecap="round"/>
        </svg>
        <span class="admin-header-title">MockInterview 后台管理</span>
      </div>
      <div class="admin-header-right">
        <span class="admin-user-info">
          <i class="el-icon-s-custom"></i>
          {{ adminUser.nickname || adminUser.username }}
        </span>
        <el-button type="text" class="admin-logout-btn" @click="handleLogout">退出登录</el-button>
      </div>
    </header>

    <!-- 内容区 -->
    <main class="admin-main">
      <!-- 统计卡片 -->
      <div class="admin-stats">
        <div class="stat-card">
          <div class="stat-icon" style="background:#e8f4fd"><i class="el-icon-user" style="color:#409eff;font-size:24px"></i></div>
          <div class="stat-info">
            <div class="stat-num">{{ userList.length }}</div>
            <div class="stat-label">注册用户总数</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon" style="background:#fef0e6"><i class="el-icon-s-check" style="color:#ff9900;font-size:24px"></i></div>
          <div class="stat-info">
            <div class="stat-num">{{ adminCount }}</div>
            <div class="stat-label">管理员账号数</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon" style="background:#e8f9f0"><i class="el-icon-time" style="color:#67c23a;font-size:24px"></i></div>
          <div class="stat-info">
            <div class="stat-num">{{ todayLoginCount }}</div>
            <div class="stat-label">今日登录用户</div>
          </div>
        </div>
      </div>

      <!-- 用户管理表格 -->
      <div class="admin-table-card">
        <div class="table-header">
          <h3 class="table-title">用户管理</h3>
          <div style="display:flex;gap:10px;align-items:center">
            <el-button
              size="small"
              icon="el-icon-refresh"
              :loading="migrateLoading"
              @click="handleMigrate"
            >导入历史用户</el-button>
            <el-input
              v-model="searchKeyword"
              placeholder="搜索用户名/昵称/手机/邮箱"
              prefix-icon="el-icon-search"
              size="small"
              clearable
              style="width: 280px"
            />
          </div>
        </div>

        <el-table
          :data="filteredUsers"
          v-loading="tableLoading"
          stripe
          style="width: 100%"
          :header-cell-style="{ background: '#f5f7fa', color: '#606266', fontWeight: '600' }"
        >
          <el-table-column label="用户名" prop="username" min-width="120" show-overflow-tooltip />
          <el-table-column label="昵称" prop="nickname" min-width="120" show-overflow-tooltip />
          <el-table-column label="手机号" prop="phone" min-width="130">
            <template v-slot="{ row }">
              <span>{{ row.phone || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="邮箱" prop="email" min-width="180" show-overflow-tooltip>
            <template v-slot="{ row }">
              <span>{{ row.email || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="角色" prop="role" width="90" align="center">
            <template v-slot="{ row }">
              <el-tag :type="row.role === 'admin' ? 'warning' : 'info'" size="small">
                {{ row.role === 'admin' ? '管理员' : '普通用户' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="注册时间" min-width="160">
            <template v-slot="{ row }">
              <span>{{ formatTime(row.createdAt) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="最后登录" min-width="160">
            <template v-slot="{ row }">
              <span>{{ row.lastLoginAt ? formatTime(row.lastLoginAt) : '从未登录' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="240" align="center" fixed="right">
            <template v-slot="{ row }">
              <el-button
                type="text"
                size="small"
                icon="el-icon-key"
                @click="openResetDialog(row)"
              >重置密码</el-button>
              <el-button
                v-if="row.userId !== adminUser.userId"
                type="text"
                size="small"
                :icon="row.role === 'admin' ? 'el-icon-remove-outline' : 'el-icon-circle-plus-outline'"
                :style="{ color: row.role === 'admin' ? '#f56c6c' : '#409eff' }"
                @click="toggleRole(row)"
              >{{ row.role === 'admin' ? '撤销管理员' : '设为管理员' }}</el-button>
              <el-button
                v-if="row.userId !== adminUser.userId"
                type="text"
                size="small"
                icon="el-icon-delete"
                style="color: #f56c6c"
                @click="handleDelete(row)"
              >删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </main>

    <!-- 重置密码弹窗 -->
    <el-dialog
      title="重置用户密码"
      :visible.sync="resetDialog.visible"
      width="400px"
      @close="resetDialogClose"
    >
      <div class="reset-user-info">
        <i class="el-icon-user-solid"></i>
        <span>{{ resetDialog.user && resetDialog.user.username }}</span>
        <el-tag size="small" type="info" style="margin-left:8px">
          {{ resetDialog.user && (resetDialog.user.nickname || '') }}
        </el-tag>
      </div>
      <el-form ref="resetForm" :model="resetDialog.form" :rules="resetRules" label-width="90px" style="margin-top:16px">
        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="resetDialog.form.newPassword"
            type="password"
            placeholder="至少 8 位"
            show-password
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="resetDialog.form.confirmPassword"
            type="password"
            placeholder="再次输入新密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <div slot="footer">
        <el-button @click="resetDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="resetDialog.loading" @click="confirmReset">确认重置</el-button>
      </div>
    </el-dialog>

    <!-- 贡献审核区域 -->
    <div class="admin-table-card contribution-review">
      <div class="table-header">
        <h3 class="table-title">📝 贡献题库审核</h3>
        <div style="display:flex;gap:10px;align-items:center">
          <el-tabs v-model="contribTabs" size="small" style="margin-right:12px">
            <el-tab-pane label="待审核" name="pending">
              <span slot="label">
                待审核
                <el-badge :value="pendingCount" :hidden="pendingCount === 0" type="warning" style="margin-left:4px"/>
              </span>
            </el-tab-pane>
            <el-tab-pane label="全部" name="all" />
          </el-tabs>
          <el-button size="small" icon="el-icon-refresh" :loading="contribLoading" @click="loadContributions">刷新</el-button>
        </div>
      </div>

      <el-table
        :data="filteredContributions"
        v-loading="contribLoading"
        stripe
        style="width: 100%"
        :header-cell-style="{ background: '#f5f7fa', color: '#606266', fontWeight: '600' }"
      >
        <el-table-column label="贡献者" min-width="100">
          <template v-slot="{ row }">
            <span>{{ row.contributor || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="公司" prop="company" width="120" show-overflow-tooltip />
        <el-table-column label="岗位" prop="jobTitle" width="120" show-overflow-tooltip />
        <el-table-column label="题目内容" min-width="200">
          <template v-slot="{ row }">
            <div style="font-size:13px;line-height:1.5;max-height:60px;overflow:hidden">{{ row.content }}</div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="80">
          <template v-slot="{ row }">
            <el-tag size="mini" type="info">{{ questionTypeText(row.questionType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="难度" width="70">
          <template v-slot="{ row }">
            <el-tag size="mini" :type="difficultyType(row.difficulty)">{{ difficultyText(row.difficulty) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template v-slot="{ row }">
            <el-tag :type="statusType(row.status)" size="mini">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="提交时间" width="140">
          <template v-slot="{ row }">
            <span style="font-size:12px">{{ formatTime(row.createdAt) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" align="center" fixed="right">
          <template v-slot="{ row }">
            <template v-if="row.status === 'pending'">
              <el-button type="text" size="small" style="color:#67c23a" @click="handleApprove(row)">通过</el-button>
              <el-button type="text" size="small" style="color:#f56c6c" @click="handleReject(row)">驳回</el-button>
            </template>
            <span v-else style="color:#999;font-size:12px">{{ row.status === 'approved' ? '已通过' : row.status === 'rejected' ? '已驳回' : '-' }}</span>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="contributions.length === 0 && !contribLoading" style="text-align:center;padding:40px 0;color:#999">
        暂无{{ contribTabs === 'pending' ? '待审核' : '' }}贡献记录
      </div>

      <!-- 分页 -->
      <el-pagination
        v-if="contributions.length > 0"
        style="margin-top:16px;text-align:right"
        :current-page="contribPage"
        :page-size="20"
        :total="contribTotal"
        layout="total, prev, pager, next"
        @current-change="handleContribPageChange"
      />
    </div>
  </div>
</template>

<script>
import { listUsers, resetUserPassword, setUserRole, deleteUser, migrateUsers, listContributions, approveContribution, rejectContribution } from '@/api/admin'

export default {
  name: 'AdminDashboard',
  data() {
    const validateConfirm = (rule, value, callback) => {
      if (value !== this.resetDialog.form.newPassword) {
        callback(new Error('两次密码不一致'))
      } else {
        callback()
      }
    }
    return {
      adminUser: JSON.parse(localStorage.getItem('admin_user') || '{}'),
      tableLoading: false,
      migrateLoading: false,
      userList: [],
      searchKeyword: '',
      resetDialog: {
        visible: false,
        user: null,
        loading: false,
        form: { newPassword: '', confirmPassword: '' }
      },
      resetRules: {
        newPassword: [
          { required: true, message: '请输入新密码', trigger: 'blur' },
          { min: 8, message: '密码不能少于8位', trigger: 'blur' }
        ],
        confirmPassword: [
          { required: true, message: '请再次输入密码', trigger: 'blur' },
          { validator: validateConfirm, trigger: 'blur' }
        ]
      },
      // 贡献审核相关
      contribTabs: 'pending',
      contribLoading: false,
      contributions: [],
      contribPage: 1,
      contribTotal: 0
    }
  },
  computed: {
    filteredUsers() {
      if (!this.searchKeyword) return this.userList
      const kw = this.searchKeyword.toLowerCase()
      return this.userList.filter(u =>
        (u.username || '').toLowerCase().includes(kw) ||
        (u.nickname || '').toLowerCase().includes(kw) ||
        (u.phone || '').includes(kw) ||
        (u.email || '').toLowerCase().includes(kw)
      )
    },
    adminCount() {
      return this.userList.filter(u => u.role === 'admin').length
    },
    todayLoginCount() {
      const today = new Date().toISOString().slice(0, 10)
      return this.userList.filter(u => u.lastLoginAt && u.lastLoginAt.startsWith(today)).length
    },
    filteredContributions() {
      if (this.contribTabs === 'all') return this.contributions
      return this.contributions.filter(c => c.status === 'pending')
    },
    pendingCount() {
      return this.contributions.filter(c => c.status === 'pending').length
    }
  },
  created() {
    this.loadUsers()
    this.loadContributions()
  },
  methods: {
    async loadUsers() {
      this.tableLoading = true
      try {
        const res = await listUsers()
        this.userList = res.data.list || []
      } catch (e) {
        // 统一由拦截器处理
      } finally {
        this.tableLoading = false
      }
    },
    formatTime(iso) {
      if (!iso) return '-'
      const d = new Date(typeof iso === 'number' && iso < 10000000000 ? iso * 1000 : iso)
      const pad = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
    },
    openResetDialog(user) {
      this.resetDialog.user = user
      this.resetDialog.form = { newPassword: '', confirmPassword: '' }
      this.resetDialog.visible = true
    },
    resetDialogClose() {
      this.$refs.resetForm && this.$refs.resetForm.clearValidate()
    },
    confirmReset() {
      this.$refs.resetForm.validate(async valid => {
        if (!valid) return
        this.resetDialog.loading = true
        try {
          await resetUserPassword(this.resetDialog.user.userId, this.resetDialog.form.newPassword)
          this.$message.success('密码重置成功')
          this.resetDialog.visible = false
        } catch (e) {
          // 拦截器处理
        } finally {
          this.resetDialog.loading = false
        }
      })
    },
    async toggleRole(user) {
      const newRole = user.role === 'admin' ? 'user' : 'admin'
      const action = newRole === 'admin' ? '设为管理员' : '撤销管理员权限'
      try {
        await this.$confirm(`确认要将用户 "${user.username}" ${action}？`, '确认操作', {
          type: 'warning',
          confirmButtonText: '确认',
          cancelButtonText: '取消'
        })
      } catch {
        return
      }
      try {
        await setUserRole(user.userId, newRole)
        this.$message.success(`操作成功`)
        this.loadUsers()
      } catch (e) {
        // 拦截器处理
      }
    },
    async handleMigrate() {
      this.migrateLoading = true
      try {
        const res = await migrateUsers()
        this.$message.success(`迁移完成，共导入 ${res.data.migratedCount} 个用户`)
        this.loadUsers()
      } catch (e) {
        // 拦截器处理
      } finally {
        this.migrateLoading = false
      }
    },
    handleLogout() {
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_user')
      this.$router.push('/admin/login')
    },
    async handleDelete(user) {
      try {
        await this.$confirm(
          `确认删除用户 "${user.username}"？删除后不可恢复，该用户的登录账号将立即失效。`,
          '危险操作确认',
          {
            type: 'warning',
            confirmButtonText: '确认删除',
            cancelButtonText: '取消',
            confirmButtonClass: 'el-button--danger'
          }
        )
      } catch {
        return
      }
      try {
        await deleteUser(user.userId)
        this.$message.success(`用户 "${user.username}" 已删除`)
        this.loadUsers()
      } catch (e) {
        // 拦截器处理
      }
    },

    // ===== 贡献审核 =====
    async loadContributions() {
      this.contribLoading = true
      try {
        const res = await listContributions(this.contribTabs === 'all' ? '' : 'pending')
        this.contributions = res.data.list || []
        this.contribTotal = res.data.total || 0
      } catch (e) {
        // 拦截器处理
      } finally {
        this.contribLoading = false
      }
    },
    handleContribPageChange(page) {
      this.contribPage = page
      this.loadContributions()
    },
    async handleApprove(row) {
      try {
        await this.$confirm(`确认通过该题目？通过后题目将进入题库。`, '审核通过', {
          type: 'success',
          confirmButtonText: '确认通过'
        })
      } catch { return }
      try {
        await approveContribution(row.id)
        this.$message.success('已通过')
        row.status = 'approved'
      } catch (e) {}
    },
    async handleReject(row) {
      try {
        await this.$confirm(`确认驳回该题目？驳回后题目将不会进入题库。`, '审核驳回', {
          type: 'warning',
          confirmButtonText: '确认驳回',
          cancelButtonText: '取消'
        })
      } catch { return }
      try {
        await rejectContribution(row.id)
        this.$message.success('已驳回')
        row.status = 'rejected'
      } catch (e) {}
    },
    questionTypeText(t) {
      return { technical: '技术', behavioral: '行为', algorithm: '算法', 'system-design': '系统设计' }[t] || t || '-'
    },
    difficultyText(d) {
      return { easy: '简单', medium: '中等', hard: '困难' }[d] || d || '-'
    },
    difficultyType(d) {
      return { easy: 'success', medium: 'warning', hard: 'danger' }[d] || 'info'
    },
    statusType(s) {
      return { pending: 'warning', approved: 'success', verified: 'primary', rejected: 'danger' }[s] || 'info'
    },
    statusText(s) {
      return { pending: '待审核', approved: '已通过', verified: '已验证', rejected: '已驳回' }[s] || s || '-'
    }
  }
}
</script>

<style scoped>
.admin-layout {
  min-height: 100vh;
  background: #f0f2f5;
  font-family: Arial, sans-serif;
}
.admin-header {
  background: #131921;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  position: fixed;
  top: 0; left: 0; right: 0;
  z-index: 100;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2);
}
.admin-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}
.admin-header-title {
  color: #ff9900;
  font-size: 17px;
  font-weight: 700;
  letter-spacing: 1px;
}
.admin-header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
.admin-user-info {
  color: #ccc;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 4px;
}
.admin-logout-btn {
  color: #ff9900 !important;
  font-size: 13px;
}
.admin-main {
  padding: 80px 24px 24px;
  max-width: 1400px;
  margin: 0 auto;
}
.admin-stats {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
}
.stat-card {
  background: #fff;
  border-radius: 8px;
  padding: 20px 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}
.stat-icon {
  width: 52px;
  height: 52px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.stat-num {
  font-size: 28px;
  font-weight: 700;
  color: #131921;
  line-height: 1.2;
}
.stat-label {
  font-size: 13px;
  color: #888;
  margin-top: 2px;
}
.admin-table-card {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}
.table-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.table-title {
  font-size: 16px;
  font-weight: 600;
  color: #131921;
}
.reset-user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  color: #333;
  padding: 8px 0;
}
.reset-user-info i {
  font-size: 18px;
  color: #409eff;
}

/* 贡献审核 */
.contribution-review {
  margin-top: 20px;
}
</style>
