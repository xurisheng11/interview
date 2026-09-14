# 面试模拟小程序

AI 驱动的面试练习平台，帮助用户更好地准备面试。

## 功能模块

### 1. 首页 (`pages/index`)
- 用户登录状态展示
- 快速开始面试入口
- 统计数据展示（面试次数、平均分等）
- 最近面试记录

### 2. 登录页 (`pages/login`)
- 微信一键登录
- 账号密码登录
- 用户协议和隐私政策
- 注册账号入口

### 3. 注册页 (`pages/register`)
- 用户名、手机号、邮箱注册
- 密码设置与确认
- 用户协议同意

### 4. 面试列表 (`pages/interview/list`)
- 面试记录列表
- 状态筛选（待开始/进行中/已完成）
- 查看面试详情

### 5. 面试配置 (`pages/interview/create`)
- 选择公司名称
- 选择岗位名称
- 面试类型（技术/HR/行为/模拟）
- 难度设置（简单/中等/困难）
- 题目数量设置（3-10题）
- 答题时间设置

### 6. 面试进行页 (`pages/interview/session`)
- 问题展示（分类、难度、提示）
- 计时器功能
- 答案输入框
- 录音功能（录制语音回答）
- 播放录音
- 上一题/下一题导航
- 跳过功能
- 自动提交

### 7. 面试报告 (`pages/interview/detail`)
- 面试基本信息
- 综合评分展示
- 分项得分（准确度、流畅度、逻辑性）
- 回答详情列表
- AI 点评
- 分享报告
- 重新面试
- 复制回答

### 8. 题库 (`pages/questions`)
- 题目分类浏览（技术/行为/HR/项目）
- 搜索功能
- 题目卡片展示

### 9. 题目详情 (`pages/questions/detail`)
- 题目信息展示
- 参考答案
- 解题思路
- 回答要点
- 常见误区
- 相关题目推荐
- 收藏功能
- 开始练习

### 10. 个人资料 (`pages/profile`)
- 头像设置
- 基本信息编辑
- 求职意向设置
- 保存同步

### 11. 设置 (`pages/settings`)
- 消息通知开关
- 声音/震动设置
- 隐私政策/用户协议
- 清除缓存
- 数据导出
- 版本信息
- 退出登录

### 12. 帮助与反馈 (`pages/help`)
- 常见问题FAQ
- 使用教程
- 意见反馈
- 联系客服

### 13. 我的 (`pages/mine`)
- 用户信息展示
- 统计数据
- 功能菜单
- 设置、关于、帮助

## 技术栈

- **框架**: 微信小程序
- **状态管理**: 自定义 Store
- **网络请求**: wx.request 封装
- **样式**: WXSS

## 页面结构

```
pages/
├── index/              # 首页
├── login/              # 登录页
├── register/           # 注册页
├── interview/
│   ├── list/          # 面试列表
│   ├── create/        # 面试配置
│   ├── session/       # 面试进行
│   └── detail/        # 面试详情
├── questions/
│   ├── index/         # 题库列表
│   └── detail/        # 题目详情
├── profile/           # 个人资料
├── settings/          # 设置
├── help/              # 帮助与反馈
└── mine/              # 我的
```

## API 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| /auth/wxlogin | POST | 微信登录 |
| /auth/login | POST | 账号密码登录 |
| /auth/register | POST | 注册 |
| /auth/me | GET | 获取当前用户 |
| /interviews | GET | 获取面试列表 |
| /interviews | POST | 创建面试 |
| /interviews/:id | GET | 获取面试详情 |
| /interviews/:id/answers | POST | 提交回答 |
| /interviews/:id/complete | PUT | 完成面试 |
| /questions | GET | 获取题库列表 |
| /reports/:id | GET | 获取面试报告 |
| /profile | GET/PUT | 用户资料 |
| /profile/stats | GET | 用户统计 |

## 运行项目

1. 安装微信开发者工具
2. 打开项目目录
3. 修改 `app.js` 中的 API 地址为后端服务地址
4. 编译运行

## 注意事项

- 需要后端 API 支持才能完整使用
- 部分功能使用模拟数据进行演示
- 录音功能需要用户授权
