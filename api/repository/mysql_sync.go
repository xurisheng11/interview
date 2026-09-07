package repository

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"interview-sim/config"
	"interview-sim/model"
)

// nullableTime 将零值时间转为 NULL，避免 DATETIME 越界写入失败
func nullableTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}

// SyncStat 单个业务域的同步统计
type SyncStat struct {
	Synced int `json:"synced"`
	Failed int `json:"failed"`
}

// scanAllKeys 用 SCAN 遍历匹配 pattern 的全部 key
func scanAllKeys(pattern string) ([]string, error) {
	var cursor uint64
	var all []string
	for {
		keys, next, err := Scan(pattern, 200, &cursor)
		if err != nil {
			return all, err
		}
		all = append(all, keys...)
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return all, nil
}

// hasAnyPrefix 判断 key 是否命中任一前缀（用于 SCAN 结果过滤）
func hasAnyPrefix(key string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}

// SyncAllToMySQL 全量同步 Redis 核心业务数据到 MySQL（只 upsert 不删除，MySQL 兼作归档）
// 单条失败记日志继续，最后返回各表统计
func SyncAllToMySQL() map[string]*SyncStat {
	stats := map[string]*SyncStat{}
	if !MySQLAvailable() {
		log.Println("SyncAllToMySQL: MySQL 不可用，跳过本次同步")
		return stats
	}
	start := time.Now()

	stats["users"] = syncUsers()
	stats["interview_sessions"] = syncSessions()
	stats["reports"] = syncReports()
	stats["resumes"] = syncResumes()
	stats["articles"] = syncArticles()
	stats["comments"] = syncComments()
	stats["contributed_questions"] = syncContributedQuestions()
	stats["user_credits"] = syncUserCredits()
	stats["sprint_plans"] = syncSprintPlans()
	// 关系类 Set 同步：插入参数顺序固定为（key 后缀, 集合成员），列语义由各自 SQL 保证
	stats["user_question_collects"] = syncSetRelations("user:collect:question:*", "user:collect:question:",
		"INSERT IGNORE INTO user_question_collects(user_id, question_id) VALUES(?, ?)")
	stats["article_likes"] = syncSetRelations("article:likes:*", "article:likes:",
		"INSERT IGNORE INTO article_relations(article_id, user_id, rel_type) VALUES(?, ?, 'like')")
	stats["article_collects"] = syncSetRelations("article:collects:*", "article:collects:",
		"INSERT IGNORE INTO article_relations(article_id, user_id, rel_type) VALUES(?, ?, 'collect')")
	stats["contribution_votes"] = syncSetRelations("contributed:votes:*", "contributed:votes:",
		"INSERT IGNORE INTO contribution_relations(question_id, user_id, rel_type) VALUES(?, ?, 'vote')")
	stats["contribution_reports"] = syncSetRelations("contributed:reports:*", "contributed:reports:",
		"INSERT IGNORE INTO contribution_relations(question_id, user_id, rel_type) VALUES(?, ?, 'report')")

	// 汇总日志
	var parts []string
	for name, s := range stats {
		parts = append(parts, name+"="+strconv.Itoa(s.Synced)+"/失败"+strconv.Itoa(s.Failed))
	}
	log.Printf("SyncAllToMySQL 完成，耗时 %s：%s", time.Since(start).Round(time.Millisecond), strings.Join(parts, "，"))
	return stats
}

// syncUsers 同步 user:{userId} Hash -> users 表
func syncUsers() *SyncStat {
	s := &SyncStat{}
	keys, err := scanAllKeys("user:*")
	if err != nil {
		log.Printf("syncUsers 扫描失败: %v", err)
		return s
	}
	for _, key := range keys {
		// 排除各类 user: 前缀的索引/子业务 key
		if hasAnyPrefix(key, "user:account:", "user:interviews:", "user:resumes:", "user:credits:", "user:collect:") {
			continue
		}
		hash, err := HGetAll(key)
		if err != nil || len(hash) == 0 || hash["userId"] == "" {
			if err != nil {
				log.Printf("syncUsers 读取 %s 失败: %v", key, err)
				s.Failed++
			}
			continue
		}
		user := model.UserFromRedisHash(hash)
		data, err := json.Marshal(user)
		if err != nil {
			log.Printf("syncUsers 序列化 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		_, err = mysqlDB.Exec(`INSERT INTO users(user_id, username, phone, email, created_at, data, synced_at)
			VALUES(?, ?, ?, ?, ?, ?, NOW())
			ON DUPLICATE KEY UPDATE username=VALUES(username), phone=VALUES(phone), email=VALUES(email),
			created_at=VALUES(created_at), data=VALUES(data), synced_at=NOW()`,
			user.UserID, user.Username, user.Phone, user.Email, nullableTime(user.CreatedAt), string(data))
		if err != nil {
			log.Printf("syncUsers 写入 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		s.Synced++
	}
	return s
}

// syncSessions 同步 interview:session:{id} JSON -> interview_sessions 表
func syncSessions() *SyncStat {
	s := &SyncStat{}
	keys, err := scanAllKeys("interview:session:*")
	if err != nil {
		log.Printf("syncSessions 扫描失败: %v", err)
		return s
	}
	for _, key := range keys {
		raw, err := Get(key)
		if err != nil {
			log.Printf("syncSessions 读取 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		session, err := model.SessionFromJSON(raw)
		if err != nil || session.InterviewID == "" {
			log.Printf("syncSessions 解析 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		_, err = mysqlDB.Exec(`INSERT INTO interview_sessions(interview_id, user_id, status, mode, start_time, data, synced_at)
			VALUES(?, ?, ?, ?, ?, ?, NOW())
			ON DUPLICATE KEY UPDATE user_id=VALUES(user_id), status=VALUES(status), mode=VALUES(mode),
			start_time=VALUES(start_time), data=VALUES(data), synced_at=NOW()`,
			session.InterviewID, session.UserID, session.Status, session.Mode,
			nullableTime(session.StartTime), raw)
		if err != nil {
			log.Printf("syncSessions 写入 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		s.Synced++
	}
	return s
}

// syncReports 同步 report:{userId}:{interviewId} JSON -> reports 表（排除 report:share:*）
func syncReports() *SyncStat {
	s := &SyncStat{}
	keys, err := scanAllKeys("report:*")
	if err != nil {
		log.Printf("syncReports 扫描失败: %v", err)
		return s
	}
	for _, key := range keys {
		if strings.HasPrefix(key, "report:share:") {
			continue
		}
		// key 格式 report:{userId}:{interviewId}
		parts := strings.SplitN(strings.TrimPrefix(key, "report:"), ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			continue
		}
		raw, err := Get(key)
		if err != nil {
			log.Printf("syncReports 读取 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		var report model.InterviewReport
		if err := json.Unmarshal([]byte(raw), &report); err != nil {
			log.Printf("syncReports 解析 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		_, err = mysqlDB.Exec(`INSERT INTO reports(interview_id, user_id, total_score, created_at, data, synced_at)
			VALUES(?, ?, ?, ?, ?, NOW())
			ON DUPLICATE KEY UPDATE user_id=VALUES(user_id), total_score=VALUES(total_score),
			created_at=VALUES(created_at), data=VALUES(data), synced_at=NOW()`,
			parts[1], parts[0], report.TotalScore, nullableTime(report.CreatedAt), raw)
		if err != nil {
			log.Printf("syncReports 写入 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		s.Synced++
	}
	return s
}

// syncResumes 同步 resume:{id} JSON -> resumes 表
func syncResumes() *SyncStat {
	s := &SyncStat{}
	keys, err := scanAllKeys("resume:*")
	if err != nil {
		log.Printf("syncResumes 扫描失败: %v", err)
		return s
	}
	for _, key := range keys {
		raw, err := Get(key)
		if err != nil {
			log.Printf("syncResumes 读取 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		var r model.ResumeRecord
		if err := json.Unmarshal([]byte(raw), &r); err != nil || r.ID == "" {
			log.Printf("syncResumes 解析 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		_, err = mysqlDB.Exec(`INSERT INTO resumes(resume_id, user_id, uploaded_at, analysis_status, data, synced_at)
			VALUES(?, ?, ?, ?, ?, NOW())
			ON DUPLICATE KEY UPDATE user_id=VALUES(user_id), uploaded_at=VALUES(uploaded_at),
			analysis_status=VALUES(analysis_status), data=VALUES(data), synced_at=NOW()`,
			r.ID, r.UserID, nullableTime(r.UploadedAt), r.AnalysisStatus, raw)
		if err != nil {
			log.Printf("syncResumes 写入 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		s.Synced++
	}
	return s
}

// syncArticles 同步 article:{id} Hash -> articles 表（排除 likes/collects/comments 子键）
func syncArticles() *SyncStat {
	s := &SyncStat{}
	keys, err := scanAllKeys("article:*")
	if err != nil {
		log.Printf("syncArticles 扫描失败: %v", err)
		return s
	}
	for _, key := range keys {
		if hasAnyPrefix(key, "article:likes:", "article:collects:", "article:comments:") {
			continue
		}
		hash, err := HGetAll(key)
		if err != nil || len(hash) == 0 {
			if err != nil {
				log.Printf("syncArticles 读取 %s 失败: %v", key, err)
				s.Failed++
			}
			continue
		}
		var a model.Article
		if err := a.FromRedisHash(hash); err != nil || a.ArticleID == "" {
			log.Printf("syncArticles 解析 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		data, _ := json.Marshal(&a)
		_, err = mysqlDB.Exec(`INSERT INTO articles(article_id, job_category, author_id, created_at, data, synced_at)
			VALUES(?, ?, ?, ?, ?, NOW())
			ON DUPLICATE KEY UPDATE job_category=VALUES(job_category), author_id=VALUES(author_id),
			created_at=VALUES(created_at), data=VALUES(data), synced_at=NOW()`,
			a.ArticleID, a.JobCategory, a.AuthorID, a.CreatedAt, string(data))
		if err != nil {
			log.Printf("syncArticles 写入 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		s.Synced++
	}
	return s
}

// syncComments 同步 comment:{id} Hash -> comments 表
func syncComments() *SyncStat {
	s := &SyncStat{}
	keys, err := scanAllKeys("comment:*")
	if err != nil {
		log.Printf("syncComments 扫描失败: %v", err)
		return s
	}
	for _, key := range keys {
		hash, err := HGetAll(key)
		if err != nil || len(hash) == 0 {
			if err != nil {
				log.Printf("syncComments 读取 %s 失败: %v", key, err)
				s.Failed++
			}
			continue
		}
		var c model.Comment
		if err := c.FromRedisHash(hash); err != nil || c.CommentID == "" {
			log.Printf("syncComments 解析 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		data, _ := json.Marshal(&c)
		_, err = mysqlDB.Exec(`INSERT INTO comments(comment_id, article_id, created_at, data, synced_at)
			VALUES(?, ?, ?, ?, NOW())
			ON DUPLICATE KEY UPDATE article_id=VALUES(article_id), created_at=VALUES(created_at),
			data=VALUES(data), synced_at=NOW()`,
			c.CommentID, c.ArticleID, c.CreatedAt, string(data))
		if err != nil {
			log.Printf("syncComments 写入 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		s.Synced++
	}
	return s
}

// syncContributedQuestions 同步 contributed:Q:{id} Hash -> contributed_questions 表
func syncContributedQuestions() *SyncStat {
	s := &SyncStat{}
	keys, err := scanAllKeys("contributed:Q:*")
	if err != nil {
		log.Printf("syncContributedQuestions 扫描失败: %v", err)
		return s
	}
	for _, key := range keys {
		hash, err := HGetAll(key)
		if err != nil || len(hash) == 0 {
			if err != nil {
				log.Printf("syncContributedQuestions 读取 %s 失败: %v", key, err)
				s.Failed++
			}
			continue
		}
		var q model.ContributedQuestion
		if err := q.FromRedisHash(hash); err != nil || q.ID == "" {
			log.Printf("syncContributedQuestions 解析 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		data, _ := json.Marshal(&q)
		_, err = mysqlDB.Exec(`INSERT INTO contributed_questions(question_id, company, contributor_id, status, created_at, data, synced_at)
			VALUES(?, ?, ?, ?, ?, ?, NOW())
			ON DUPLICATE KEY UPDATE company=VALUES(company), contributor_id=VALUES(contributor_id),
			status=VALUES(status), created_at=VALUES(created_at), data=VALUES(data), synced_at=NOW()`,
			q.ID, q.Company, q.ContributorID, q.Status, q.CreatedAt, string(data))
		if err != nil {
			log.Printf("syncContributedQuestions 写入 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		s.Synced++
	}
	return s
}

// syncUserCredits 同步 user:credits:{userId} Hash(balance) -> user_credits 表
func syncUserCredits() *SyncStat {
	s := &SyncStat{}
	keys, err := scanAllKeys("user:credits:*")
	if err != nil {
		log.Printf("syncUserCredits 扫描失败: %v", err)
		return s
	}
	for _, key := range keys {
		userID := strings.TrimPrefix(key, "user:credits:")
		if userID == "" {
			continue
		}
		hash, err := HGetAll(key)
		if err != nil || len(hash) == 0 {
			if err != nil {
				log.Printf("syncUserCredits 读取 %s 失败: %v", key, err)
				s.Failed++
			}
			continue
		}
		balance, _ := strconv.ParseInt(hash["balance"], 10, 64)
		_, err = mysqlDB.Exec(`INSERT INTO user_credits(user_id, balance, synced_at) VALUES(?, ?, NOW())
			ON DUPLICATE KEY UPDATE balance=VALUES(balance), synced_at=NOW()`,
			userID, balance)
		if err != nil {
			log.Printf("syncUserCredits 写入 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		s.Synced++
	}
	return s
}

// syncSprintPlans 同步 sprint_plan:{userId} JSON -> sprint_plans 表
func syncSprintPlans() *SyncStat {
	s := &SyncStat{}
	keys, err := scanAllKeys("sprint_plan:*")
	if err != nil {
		log.Printf("syncSprintPlans 扫描失败: %v", err)
		return s
	}
	for _, key := range keys {
		userID := strings.TrimPrefix(key, "sprint_plan:")
		if userID == "" {
			continue
		}
		raw, err := Get(key)
		if err != nil {
			log.Printf("syncSprintPlans 读取 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		_, err = mysqlDB.Exec(`INSERT INTO sprint_plans(user_id, data, synced_at) VALUES(?, ?, NOW())
			ON DUPLICATE KEY UPDATE data=VALUES(data), synced_at=NOW()`,
			userID, raw)
		if err != nil {
			log.Printf("syncSprintPlans 写入 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		s.Synced++
	}
	return s
}

// syncSetRelations 通用 Set 关系同步：Redis Set -> 关系表（INSERT IGNORE）
// 插入参数顺序固定为（key 后缀, 集合成员），列语义由调用方 SQL 的列顺序保证，例如：
//   - article:likes:{articleId} 成员为 userId -> (article_id, user_id)
//   - user:collect:question:{userId} 成员为 questionId -> (user_id, question_id)
func syncSetRelations(pattern, prefix, insertSQL string) *SyncStat {
	s := &SyncStat{}
	keys, err := scanAllKeys(pattern)
	if err != nil {
		log.Printf("syncSetRelations(%s) 扫描失败: %v", pattern, err)
		return s
	}
	for _, key := range keys {
		id := strings.TrimPrefix(key, prefix)
		if id == "" {
			continue
		}
		members, err := SMembers(key)
		if err != nil {
			log.Printf("syncSetRelations 读取 %s 失败: %v", key, err)
			s.Failed++
			continue
		}
		for _, m := range members {
			if m == "" {
				continue
			}
			if _, err := mysqlDB.Exec(insertSQL, id, m); err != nil {
				log.Printf("syncSetRelations 写入 %s(%s) 失败: %v", key, m, err)
				s.Failed++
				continue
			}
			s.Synced++
		}
	}
	return s
}

// StartNightlySync 启动每晚定时全量同步 goroutine（默认凌晨 MYSQL_SYNC_HOUR 整点触发）
func StartNightlySync() {
	go func() {
		for {
			next := nextSyncTime(config.Cfg.MySQLSyncHour)
			log.Printf("下一次 MySQL 全量同步时间: %s", next.Format("2006-01-02 15:04:05"))
			timer := time.NewTimer(time.Until(next))
			<-timer.C
			runSyncSafely()
		}
	}()
}

// nextSyncTime 计算下一个 syncHour 整点
func nextSyncTime(syncHour int) time.Time {
	if syncHour < 0 || syncHour > 23 {
		syncHour = 2
	}
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), syncHour, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

// runSyncSafely 带 panic recover 的同步执行
func runSyncSafely() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("SyncAllToMySQL panic 已恢复: %v", r)
		}
	}()
	SyncAllToMySQL()
}
