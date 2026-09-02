# Requirements Document

## Introduction

本功能模块在现有求职面试平台（Vue 2 + Element UI 前端 / Go + MongoDB 后端）中新增"简历"功能板块。该模块包含两大核心能力：

1. **简历驱动面试**：用户上传简历后，系统解析简历内容（项目经历、技能栈、工作经历等），AI 自动生成针对该简历的个性化面试题目，用户可发起"简历面试"模式进行针对性练习。

2. **简历评分与分析**：上传简历后，AI 对简历进行多维度综合评分（满分100分），并给出详细改进建议；用户可查看历史简历分析记录。

该模块需集成到现有侧边栏导航体系，复用已有的 DeepSeek AI 服务（`Chat()` 函数）和 JWT 鉴权中间件。

---

## Glossary

- **Resume_Parser**：负责解析上传简历文件内容、提取结构化信息的后端服务组件。
- **Resume_Analyzer**：负责调用 AI 对简历内容进行评分和改进建议生成的后端服务组件。
- **Resume_Interview_Generator**：负责基于简历结构化内容生成面试题目的后端服务组件。
- **Resume_Store**：负责简历元数据、分析结果及面试会话持久化存储的后端存储层（MongoDB）。
- **Resume_UI**：简历功能模块的前端页面及组件集合。
- **ResumeRecord**：单条简历记录，包含文件元数据、解析内容、评分结果等。
- **ResumeAnalysisResult**：AI 对单份简历的评分与分析结果，包含总分、各维度分项、改进建议。
- **ResumeInterviewSession**：基于简历生成的面试会话，复用现有 `InterviewSession` 结构，并通过 `source` 字段标记来源为简历。
- **Authenticated_User**：已通过 JWT 鉴权的登录用户。
- **DeepSeek_Service**：现有封装好的 AI 调用服务，通过 `Chat(prompt string) (string, error)` 函数对外提供能力。

---

## Requirements

### Requirement 1: 简历上传与解析

**User Story:** As an Authenticated_User, I want to upload my resume in PDF or Word format, so that the system can extract my experience and generate personalized interview questions.

#### Acceptance Criteria

1. THE Resume_UI SHALL provide a file upload entry point accessible from the main sidebar navigation under a "简历" tab item.
2. WHEN an Authenticated_User selects a file whose extension is not `.pdf`, `.doc`, or `.docx`, THE Resume_UI SHALL reject the file before upload and display an error message stating only PDF and Word files are accepted.
3. WHEN an Authenticated_User uploads a file exceeding 10 MB, THE Resume_UI SHALL display an error message indicating the 10 MB file size limit and reject the upload without sending the file to the server.
4. WHEN an Authenticated_User submits a file that passes both the extension check and the size check, THE Resume_Parser SHALL extract structured information including work experience entries, project experience entries, and skill keywords from the file content.
5. WHILE the Resume_Parser is processing the uploaded file, THE Resume_UI SHALL display a loading indicator on the upload button and prevent re-submission until processing completes or fails.
6. WHEN the Resume_Parser successfully extracts structured information, THE Resume_Store SHALL persist a ResumeRecord containing the original filename, file size in bytes, upload timestamp, MIME type, extracted structured content, and the Authenticated_User's user ID.
7. IF the Resume_Parser fails to extract readable text from the uploaded file, THEN THE Resume_Parser SHALL return an error response with a descriptive message, and THE Resume_UI SHALL display that error message to the user and remain on the upload page.
8. WHEN a ResumeRecord is successfully created, THE Resume_UI SHALL display a success notification and navigate to the resume detail page for that record.

---

### Requirement 2: 简历评分与分析

**User Story:** As an Authenticated_User, I want to receive an AI-generated score and analysis for my uploaded resume, so that I can understand its quality and know how to improve it.

#### Acceptance Criteria

1. WHEN a ResumeRecord is created, THE Resume_Analyzer SHALL automatically initiate an analysis task for that record without requiring additional user action.
2. WHILE a resume analysis task is in progress, THE Resume_UI SHALL display a loading state on the analysis result panel and prevent the "开始简历面试" button from being clicked.
3. WHEN the Resume_Analyzer completes analysis, THE Resume_Analyzer SHALL produce a ResumeAnalysisResult containing: a total score between 0 and 100, individual dimension scores (each between 0 and 100) for project experience quality, skill description completeness, format compliance, and highlights extraction, plus a list of at least three improvement suggestions.
4. WHEN a ResumeAnalysisResult is available, THE Resume_UI SHALL display the total score, all four dimension scores, and the improvement suggestions on the resume detail page.
5. IF the DeepSeek_Service exhausts its internal retries and returns an error, THEN THE Resume_Analyzer SHALL set the ResumeRecord's analysis status to `"failed"` and persist that status without performing additional retries at the analyzer layer.
6. THE Resume_UI SHALL render the four dimension scores using the existing `RadarChart` component with input format `[{name, score}]`, and SHALL render the total score using the existing `ScoreBar` component.
7. IF a ResumeRecord's analysis status is `"failed"`, THEN THE Resume_UI SHALL display an error message on the analysis panel and provide a retry button that re-initiates the analysis task for that record.

---

### Requirement 3: 简历历史记录管理

**User Story:** As an Authenticated_User, I want to view a list of all resumes I have uploaded with their analysis results, so that I can track improvements over time.

#### Acceptance Criteria

1. THE Resume_UI SHALL provide a resume list page at `/resume` that displays all ResumeRecords belonging to the Authenticated_User, ordered by upload time descending, showing each record's original filename, upload date, analysis status, and total score (when available).
2. WHEN the resume list page loads, THE Resume_Store SHALL return only ResumeRecords whose user ID matches the Authenticated_User's user ID, with a maximum of 50 records per response.
3. WHEN the Authenticated_User clicks a ResumeRecord in the list, THE Resume_UI SHALL navigate to `/resume/:id` and display the ResumeAnalysisResult for that record.
4. WHEN the Authenticated_User confirms deletion of a ResumeRecord, THE Resume_Store SHALL remove the record and its associated ResumeAnalysisResult, and THE Resume_UI SHALL remove that item from the displayed list without a full page reload.
5. IF the Resume_Store returns an error during deletion, THEN THE Resume_UI SHALL display an error notification and leave the record in the list.
6. IF the Authenticated_User has no uploaded resumes, THEN THE Resume_UI SHALL display an empty-state prompt with a button that navigates to the resume upload entry point.

---

### Requirement 4: 简历驱动面试题目生成

**User Story:** As an Authenticated_User, I want the system to generate interview questions tailored to my resume content, so that I can practice answering questions specific to my background.

#### Acceptance Criteria

1. WHEN a ResumeRecord's ResumeAnalysisResult has a non-failed status, THE Resume_UI SHALL display an enabled "开始简历面试" action button on the resume detail page.
2. WHEN the Authenticated_User clicks "开始简历面试", THE Resume_Interview_Generator SHALL generate interview questions based on the structured content of the ResumeRecord, including at least one project-deep-dive question per project experience entry and at least one skill-assessment question per skill keyword.
3. WHILE the Resume_Interview_Generator is generating questions, THE Resume_UI SHALL display a loading state on the "开始简历面试" button and prevent duplicate submissions.
4. WHEN the Resume_Interview_Generator successfully produces a question list of 5 to 15 Question objects each with a non-empty `content` field, THE Resume_Store SHALL create a ResumeInterviewSession with `source` set to `"resume"`, `resumeId` referencing the originating ResumeRecord, and `config` fields populated with `jobTitle` from the ResumeRecord's extracted content, `difficulty` set to `"middle"`, `round` set to `"round1"`, and `mode` set to `"text"`.
5. WHEN the ResumeInterviewSession is created, THE Resume_UI SHALL navigate to `/interview/:id/doing`, reusing the existing interview-in-progress page without modification.
6. IF the Resume_Interview_Generator fails to parse a valid question list on the first attempt, THE Resume_Interview_Generator SHALL retry once; IF the second attempt also fails, THE Resume_Interview_Generator SHALL return an error response and THE Resume_UI SHALL display an error notification to the user.

---

### Requirement 5: 简历面试会话兼容性

**User Story:** As an Authenticated_User, I want the resume-driven interview session to work seamlessly within the existing interview flow, so that I get a consistent and familiar experience.

#### Acceptance Criteria

1. THE Resume_Store SHALL store ResumeInterviewSession data in Redis using the same `InterviewSession` schema fields (`interviewId`, `userId`, `config`, `questions`, `currentIndex`, `answers`, `status`, `mode`, `startTime`) with two additional fields: `source` set to `"resume"` and `resumeId` referencing the originating ResumeRecord; the same TTL rules apply: 7-day TTL while status is `"ongoing"` or `"paused"`, permanent storage when status transitions to `"completed"`.
2. WHEN an Authenticated_User completes a ResumeInterviewSession, THE Resume_UI SHALL navigate to `/report/:interviewId`, which SHALL display the total score, per-question scores, and answer feedback without requiring any modification to the existing report page.
3. WHEN an Authenticated_User views the interview history list at `/interview/history`, THE Resume_Store SHALL include ResumeInterviewSession entries in the list returned by the existing interview list endpoint, and THE Resume_UI SHALL render a "简历面试" tag in the "轮次" column for each entry whose `source` field equals `"resume"`.
4. THE Resume_Interview_Generator SHALL format each generated question using the `Question` model fields (`index`, `content`, `tags`, `difficulty`, `estimatedMinutes`, `type`), where `difficulty` is one of `"junior"`, `"middle"`, or `"senior"`, `estimatedMinutes` is an integer between 1 and 10 inclusive, and `type` is one of `"basic"`, `"algorithm"`, `"design"`, or `"hr"`.

---

### Requirement 6: 后端 API 接口规范

**User Story:** As a developer, I want well-defined RESTful API endpoints for resume operations, so that the frontend and backend remain decoupled and maintainable.

#### Acceptance Criteria

1. THE Resume_Store SHALL expose the following REST endpoints under `/api/v1/resumes`, all requiring JWT authentication:
   - `POST /api/v1/resumes` — upload and parse a resume file (multipart/form-data)
   - `GET /api/v1/resumes` — list the Authenticated_User's ResumeRecords
   - `GET /api/v1/resumes/:id` — get a single ResumeRecord with its ResumeAnalysisResult
   - `DELETE /api/v1/resumes/:id` — delete a ResumeRecord
   - `POST /api/v1/resumes/:id/interview` — create a ResumeInterviewSession from the ResumeRecord
2. WHEN a request is made to any `/api/v1/resumes` endpoint without a valid JWT token, THE Resume_Store SHALL return HTTP 401.
3. WHEN a request is made to any `/api/v1/resumes/:id` endpoint (GET, DELETE, or POST /:id/interview) with an ID that does not belong to the Authenticated_User, THE Resume_Store SHALL return HTTP 403.
4. WHEN the file upload request body contains a MIME type other than `application/pdf`, `application/msword`, or `application/vnd.openxmlformats-officedocument.wordprocessingml.document`, or when the file exceeds 10 MB, THE Resume_Store SHALL return HTTP 400 with a descriptive error message identifying the rejection reason.
5. WHEN a request is made to any `/api/v1/resumes/:id` endpoint with an ID that does not exist in the Resume_Store, THE Resume_Store SHALL return HTTP 404.
6. WHEN `POST /api/v1/resumes` successfully creates a ResumeRecord, THE Resume_Store SHALL return HTTP 201 with a response body containing the new record's ID and an `analysisStatus` field set to `"pending"`.

---

### Requirement 7: 前端导航集成

**User Story:** As an Authenticated_User, I want to access the resume module from the main sidebar navigation, so that I can find it consistently alongside other features.

#### Acceptance Criteria

1. THE Resume_UI SHALL add a "📄 简历" navigation item to the main layout's sidebar `items` configuration array in the same group as "面试模拟" and "题库练习", so that it appears in every page that renders the shared sidebar.
2. WHEN the Authenticated_User clicks the "📄 简历" navigation item, THE Resume_UI SHALL call `this.$router.push('/resume')`, which SHALL render the resume list page component.
3. THE Resume_UI SHALL register the following lazy-loaded routes in `src/router/index.js`: `{ path: '/resume', name: 'ResumeList', component: () => import('@/views/resume/List.vue'), meta: { requiresAuth: true, title: '我的简历' } }` and `{ path: '/resume/:id', name: 'ResumeDetail', component: () => import('@/views/resume/Detail.vue'), meta: { requiresAuth: true, title: '简历详情' } }`.
4. WHEN the current route path is `/resume` or starts with `/resume/`, THE Sidebar component's existing `isActive(path)` method SHALL return `true` for the "📄 简历" item's path `/resume`, applying the same orange highlight and left-border styling as other active navigation items.
