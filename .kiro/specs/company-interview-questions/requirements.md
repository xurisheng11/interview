# 需求文档

## 简介

本功能在"公司情报"（Company Intel）页面中新增**常见面试题展示**模块。用户查询某家公司的面试情报后，可在结果页中看到该公司的常见面试题列表，默认展示 11 道题，并可点击查看每道题的答案详情；列表底部提供"查看更多"入口，点击后加载并展示更多面试题。

后端复用现有题库服务（`question_service.go`），通过带公司/岗位筛选的分页接口提供数据；前端在已有的 `Intel.vue` 页面中扩展面试题板块。

---

## 词汇表

- **CompanyIntelPage**：公司情报页面（`Intel.vue`），用户在此搜索公司并查看面试情报结果
- **InterviewQuestion**：面试题条目，包含题目内容、类型（技术/行为）、标签、题目 ID 等字段
- **QuestionDetail**：面试题详情，包括 AI 生成的参考答案、考点分析等
- **QuestionList**：公司面试题列表组件，渲染 InterviewQuestion 集合
- **QuestionDetailPanel**：面试题详情面板，以展开或弹窗形式显示 QuestionDetail
- **QuestionService**：后端题库服务（`question_service.go`），提供题目检索与 AI 生成功能
- **QuestionAPI**：前端题库请求模块（`question.js`），封装对 `GET /api/v1/questions` 和 `GET /api/v1/questions/:id` 的调用
- **DefaultPageSize**：默认每页展示题目数量，固定为 11
- **LoadMorePageSize**：点击"查看更多"每次追加的题目数量，固定为 10

---

## 需求列表

### 需求 1：默认展示公司面试题列表

**用户故事：** 作为求职者，我希望在查看某家公司的面试情报时，能够自动看到该公司的常见面试题列表，以便快速了解考察重点。

#### 验收标准

1. WHEN 用户在 CompanyIntelPage 成功获取公司情报结果，THE QuestionList SHALL 在情报结果区域内自动展示该公司的常见面试题，每道题目显示序号、题目内容、题目类型标签（技术 / 行为）和关键标签。
2. THE QuestionList SHALL 默认展示不超过 DefaultPageSize（11）道面试题。
3. WHEN 后端返回的可用题目数量少于 DefaultPageSize，THE QuestionList SHALL 展示所有可用题目，不显示空占位行。
4. THE QuestionService SHALL 根据公司名称和岗位方向作为筛选条件，从题库中检索匹配的面试题；若无缓存命中，则调用 AI 服务生成并缓存后返回。
5. WHEN 题目列表加载中，THE QuestionList SHALL 显示骨架屏或加载占位符，防止页面抖动。
6. IF 题目列表加载失败，THEN THE QuestionList SHALL 显示错误提示信息，并提供重试入口。

---

### 需求 2：面试题详情展示

**用户故事：** 作为求职者，我希望点击某道面试题后能查看该题的参考答案和考点分析，以便有针对性地备考。

#### 验收标准

1. WHEN 用户点击 QuestionList 中的某道题目，THE QuestionDetailPanel SHALL 展开并显示该题的参考答案和考点分析。
2. THE QuestionDetailPanel SHALL 以展开/折叠（accordion）方式呈现：同一时刻最多展开一道题的详情，再次点击同一题时 THE QuestionDetailPanel SHALL 收起。
3. WHEN QuestionDetailPanel 首次展开某道题，THE QuestionAPI SHALL 请求 `GET /api/v1/questions/:id` 获取题目详情；同一会话内再次展开同一题时，THE QuestionDetailPanel SHALL 直接使用已缓存的详情数据，不重复发起网络请求。
4. WHILE QuestionDetailPanel 正在加载题目详情，THE QuestionDetailPanel SHALL 显示加载状态（如 loading spinner）。
5. IF 题目详情请求失败，THEN THE QuestionDetailPanel SHALL 在展开区域内显示错误提示，并提供重试按钮。
6. THE QuestionDetailPanel SHALL 使用 Markdown 渲染参考答案内容，与项目现有 MarkdownRender 组件保持一致。

---

### 需求 3：查看更多面试题

**用户故事：** 作为求职者，我希望能够加载更多面试题，以便在浏览初始列表后按需获取额外题目进行备考。

#### 验收标准

1. WHEN 后端可用题目数量超过 DefaultPageSize，THE QuestionList SHALL 在列表底部显示"查看更多"按钮，并注明当前已展示题目数与总数（如"已展示 11 / 20 道"）。
2. WHEN 用户点击"查看更多"按钮，THE QuestionList SHALL 向后端请求下一页题目（每次追加 LoadMorePageSize 即 10 道），并将新题目追加至现有列表末尾。
3. WHILE 追加题目请求进行中，THE QuestionList SHALL 禁用"查看更多"按钮并显示加载状态，防止重复点击触发重复请求。
4. WHEN 所有题目已全部展示，THE QuestionList SHALL 隐藏"查看更多"按钮，改为显示"已加载全部题目"提示文字。
5. IF 追加题目请求失败，THEN THE QuestionList SHALL 恢复"查看更多"按钮的可点击状态，并显示错误提示，允许用户重试。
6. THE QuestionList SHALL 在追加新题目时，保留已展开题目的 QuestionDetailPanel 展开状态不变。

---

### 需求 4：题目类型筛选与题目状态一致性

**用户故事：** 作为求职者，我希望在面试题列表上按"技术题 / 行为题 / 全部"筛选，并且切换公司搜索时题目列表能正确重置，以便每次都获得准确的参考信息。

#### 验收标准

1. THE QuestionList SHALL 提供"全部 / 技术题 / 行为题"三个筛选选项，默认选中"全部"；WHEN 用户切换筛选选项，THE QuestionList SHALL 仅展示与所选类型匹配的题目，其余题目不可见。
2. WHEN 用户在 CompanyIntelPage 发起新的公司搜索，THE QuestionList SHALL 重置为初始状态：清空已加载题目、将筛选选项恢复为"全部"、收起所有已展开的 QuestionDetailPanel、将分页重置为第一页。
3. THE QuestionList SHALL 在筛选状态变化后，保持已展开 QuestionDetailPanel 的缓存数据有效，不清除会话内已获取的题目详情。
