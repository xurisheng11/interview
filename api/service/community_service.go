package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"interview-sim/config"
	"interview-sim/model"
	"interview-sim/repository"
)

// ---------- 敏感词 ----------

var sensitiveWords = []string{"fuck", "shit", "傻逼", "操你", "草泥马"}

func containsSensitiveWord(content string) bool {
	lower := strings.ToLower(content)
	for _, w := range sensitiveWords {
		if strings.Contains(lower, strings.ToLower(w)) {
			return true
		}
	}
	return false
}

// ---------- Key 辅助函数 ----------

func articleKey(articleID string) string {
	return "article:" + articleID
}

func articleTimeKey(jobCategory string) string {
	return "articles:time:" + jobCategory
}

func articleHotKey(jobCategory string) string {
	return "articles:hot:" + jobCategory
}

func articleLikesKey(articleID string) string {
	return "article:likes:" + articleID
}

func articleCollectsKey(articleID string) string {
	return "article:collects:" + articleID
}

func userCollectArticleKey(userID string) string {
	return "user:collect:article:" + userID
}

func commentKey(commentID string) string {
	return "comment:" + commentID
}

func articleCommentsKey(articleID string) string {
	return "article:comments:" + articleID
}

func communityAILimitKey(userID string) string {
	return "community:ai:limit:" + userID
}

func articleSeedLockKey(jobCategory string) string {
	return "articles:seed:lock:" + jobCategory
}

func articleSeedCursorKey(jobCategory string) string {
	return "articles:seed:cursor:" + jobCategory
}

// ---------- 每个分类的预设话题列表（用于自动 seed）----------

// seedTopics 返回指定分类下预设的话题列表，每条是 [topic, jobCategory]
var categoryTopics = map[string][][2]string{
	"backend": {
		{"Redis持久化机制RDB与AOF", "backend"},
		{"MySQL索引原理与优化", "backend"},
		{"Java虚拟机GC垃圾回收机制", "backend"},
		{"Spring Boot自动装配原理", "backend"},
		{"分布式锁的实现方案", "backend"},
		{"Kafka消息队列核心原理", "backend"},
		{"TCP三次握手四次挥手", "backend"},
		{"HTTP与HTTPS区别及TLS握手", "backend"},
		{"微服务架构与单体架构对比", "backend"},
		{"数据库事务ACID与隔离级别", "backend"},
		{"线程池核心参数与调优", "backend"},
		{"布隆过滤器原理与应用", "backend"},
		{"缓存穿透缓存击穿缓存雪崩", "backend"},
		{"B+树与B树的区别", "backend"},
		{"常见限流算法令牌桶漏桶", "backend"},
		{"Elasticsearch基本原理与使用", "backend"},
		{"Docker容器化核心概念", "backend"},
		{"Kubernetes核心组件与调度", "backend"},
		{"分布式事务解决方案", "backend"},
		{"消息队列选型RabbitMQ vs Kafka", "backend"},
		{"Go语言协程与channel机制", "backend"},
		{"Redis集群与哨兵模式", "backend"},
		{"MySQL主从复制原理", "backend"},
		{"JWT认证机制原理与安全", "backend"},
		{"设计模式：观察者模式与事件驱动", "backend"},
		{"RESTful API设计最佳实践", "backend"},
		{"SQL慢查询分析与EXPLAIN使用", "backend"},
		{"HashMap底层原理与并发安全", "backend"},
		{"Spring事务管理原理", "backend"},
		{"服务网格Service Mesh与Istio", "backend"},
	},
	"frontend": {
		{"Vue3响应式原理Proxy与Reflect", "frontend"},
		{"React Hooks原理与最佳实践", "frontend"},
		{"浏览器事件循环机制", "frontend"},
		{"JavaScript原型链与继承", "frontend"},
		{"CSS布局Flex与Grid对比", "frontend"},
		{"Webpack打包优化策略", "frontend"},
		{"前端性能优化完整指南", "frontend"},
		{"TypeScript类型系统核心概念", "frontend"},
		{"浏览器渲染流程与重排重绘", "frontend"},
		{"跨域问题解决方案CORS与代理", "frontend"},
		{"前端安全XSS与CSRF防御", "frontend"},
		{"Vue虚拟DOM与Diff算法", "frontend"},
		{"微前端架构方案对比", "frontend"},
		{"前端状态管理Vuex与Pinia", "frontend"},
		{"PWA渐进式Web应用核心技术", "frontend"},
		{"前端单元测试Jest最佳实践", "frontend"},
		{"ES6核心特性Promise与async", "frontend"},
		{"浏览器缓存策略强缓存与协商缓存", "frontend"},
		{"前端监控与错误追踪方案", "frontend"},
		{"CSS动画性能优化", "frontend"},
		{"Node.js事件循环与异步IO", "frontend"},
		{"前端工程化CI/CD流水线", "frontend"},
		{"WebSocket实时通信应用", "frontend"},
		{"前端路由原理Hash与History", "frontend"},
		{"图片懒加载与虚拟滚动优化", "frontend"},
		{"JavaScript内存泄漏排查", "frontend"},
		{"HTTP/2与HTTP/3新特性", "frontend"},
		{"前端模块化CommonJS与ESM", "frontend"},
		{"React性能优化memo与useMemo", "frontend"},
		{"移动端适配方案rem与vw", "frontend"},
	},
	"bigdata": {
		{"Hadoop HDFS架构原理", "bigdata"},
		{"Spark RDD与DataFrame核心概念", "bigdata"},
		{"Flink流批一体架构详解", "bigdata"},
		{"Hive数据仓库SQL优化", "bigdata"},
		{"Kafka在大数据中的应用", "bigdata"},
		{"数据湖与数据仓库的区别", "bigdata"},
		{"Spark调优内存与并行度", "bigdata"},
		{"ClickHouse列式存储原理", "bigdata"},
		{"MapReduce编程模型原理", "bigdata"},
		{"大数据ETL流程设计", "bigdata"},
		{"Flink窗口与水位线机制", "bigdata"},
		{"数仓分层ODS/DWD/DWS/ADS", "bigdata"},
		{"HBase行列式存储原理", "bigdata"},
		{"Spark Streaming与结构化流", "bigdata"},
		{"数据倾斜问题排查与解决", "bigdata"},
		{"Yarn资源调度原理", "bigdata"},
		{"Iceberg数据湖表格式", "bigdata"},
		{"大数据实时数仓架构设计", "bigdata"},
		{"Zookeeper分布式协调原理", "bigdata"},
		{"Presto分布式查询引擎", "bigdata"},
	},
	"ai": {
		{"Transformer架构详解Attention机制", "ai"},
		{"大语言模型微调LoRA与SFT", "ai"},
		{"机器学习模型评估指标AUC/F1", "ai"},
		{"卷积神经网络CNN核心原理", "ai"},
		{"推荐系统协同过滤算法", "ai"},
		{"梯度下降与优化器Adam", "ai"},
		{"过拟合与正则化L1/L2/Dropout", "ai"},
		{"特征工程核心方法与实践", "ai"},
		{"强化学习基本概念与应用", "ai"},
		{"目标检测YOLO系列算法", "ai"},
		{"BERT预训练模型原理", "ai"},
		{"知识图谱构建与应用", "ai"},
		{"RAG检索增强生成技术", "ai"},
		{"向量数据库与语义搜索", "ai"},
		{"模型量化与推理加速", "ai"},
		{"多模态大模型技术进展", "ai"},
		{"机器学习特征选择方法", "ai"},
		{"XGBoost梯度提升树原理", "ai"},
		{"自然语言处理NLP核心任务", "ai"},
		{"深度学习框架PyTorch核心用法", "ai"},
	},
	"accounting": {
		{"资产负债表核心科目解读", "accounting"},
		{"利润表与现金流量表分析", "accounting"},
		{"财务比率分析盈利能力指标", "accounting"},
		{"增值税核心原理与申报", "accounting"},
		{"企业所得税汇算清缴要点", "accounting"},
		{"成本核算方法与管理会计", "accounting"},
		{"内部控制与风险管理", "accounting"},
		{"财务报表合并原则", "accounting"},
		{"固定资产折旧方法对比", "accounting"},
		{"应收账款管理与坏账计提", "accounting"},
		{"Excel财务数据处理技巧", "accounting"},
		{"预算管理与差异分析", "accounting"},
		{"税务筹划基本思路", "accounting"},
		{"会计准则与国际财务报告准则", "accounting"},
		{"资金管理与现金流预测", "accounting"},
	},
	"general": {
		{"职场沟通与表达技巧", "general"},
		{"时间管理GTD方法论", "general"},
		{"项目管理敏捷与Scrum", "general"},
		{"职业规划与个人发展路径", "general"},
		{"团队协作与冲突处理", "general"},
		{"数据分析思维框架", "general"},
		{"结构化思维金字塔原理", "general"},
		{"商务写作邮件与报告", "general"},
		{"面试技巧STAR法则", "general"},
		{"薪资谈判策略与技巧", "general"},
		{"远程办公效率提升方法", "general"},
		{"领导力与影响力培养", "general"},
		{"批判性思维与问题解决", "general"},
		{"职场情绪管理", "general"},
		{"OKR目标管理方法", "general"},
	},
}

// getSeedTopicsForCategory 获取指定分类的种子话题列表（all 分类合并所有）
func getSeedTopicsForCategory(jobCategory string) [][2]string {
	if jobCategory == "all" {
		var all [][2]string
		for _, topics := range categoryTopics {
			all = append(all, topics...)
		}
		return all
	}
	return categoryTopics[jobCategory]
}

// ---------- in-memory 锁防止同进程并发 seed ----------
var seedMu sync.Mutex
var seedingCategories = map[string]bool{}

// ---------- 文章写入辅助 ----------

// saveArticleToRedis 将文章写入 Redis Hash
func saveArticleToRedis(a *model.Article) error {
	hash := a.ToRedisHash()
	return repository.HSetMap(articleKey(a.ArticleID), hash)
}

// ---------- 自动 Seed ----------

// seedTarget 每个分类期望保持的最少文章数
const seedTarget = 100

// seedMinTrigger 文章数低于此值时触发自动补全
const seedMinTrigger = 20

// TriggerSeedIfNeeded 检查分类文章数，不足则异步后台补全
// 传入 "all" 时检查全局；传入具体分类时只检查该分类
func TriggerSeedIfNeeded(jobCategory string) {
	// 检查当前数量
	zsetKey := articleTimeKey(jobCategory)
	ids, err := repository.ZRevRange(zsetKey, 0, -1)
	if err != nil {
		return
	}
	current := len(ids)
	if current >= seedMinTrigger {
		return
	}

	// 用内存锁防止同一进程并发触发
	seedMu.Lock()
	if seedingCategories[jobCategory] {
		seedMu.Unlock()
		return
	}
	seedingCategories[jobCategory] = true
	seedMu.Unlock()

	go func() {
		defer func() {
			seedMu.Lock()
			delete(seedingCategories, jobCategory)
			seedMu.Unlock()
		}()
		doSeed(jobCategory, current)
	}()
}

// doSeed 后台异步补全文章到 seedTarget 篇
func doSeed(jobCategory string, alreadyHave int) {
	topics := getSeedTopicsForCategory(jobCategory)
	if len(topics) == 0 {
		return
	}

	// 读取当前已生成到第几个话题（cursor）
	cursorKey := articleSeedCursorKey(jobCategory)
	cursorStr, _ := repository.Get(cursorKey)
	cursor := 0
	if cursorStr != "" {
		fmt.Sscanf(cursorStr, "%d", &cursor)
	}

	need := seedTarget - alreadyHave
	if need <= 0 {
		return
	}

	log.Printf("[seed] 开始为分类 %s 补全文章，当前 %d 篇，目标 %d 篇", jobCategory, alreadyHave, seedTarget)

	generated := 0
	for generated < need {
		idx := (cursor + generated) % len(topics)
		pair := topics[idx]
		topic, cat := pair[0], pair[1]

		article, err := GenerateArticle(topic, cat)
		if err != nil {
			log.Printf("[seed] 生成文章失败 topic=%s: %v", topic, err)
			time.Sleep(2 * time.Second)
			continue
		}

		article.ArticleID = uuid.New().String()
		article.CreatedAt = time.Now().Unix()

		if err := saveArticleToRedis(article); err != nil {
			log.Printf("[seed] 写入 Redis 失败: %v", err)
			continue
		}

		score := float64(article.CreatedAt)
		_ = repository.ZAdd(articleTimeKey(cat), score, article.ArticleID)
		_ = repository.ZAdd(articleTimeKey("all"), score, article.ArticleID)
		_ = repository.ZAdd(articleHotKey(cat), 0, article.ArticleID)
		_ = repository.ZAdd(articleHotKey("all"), 0, article.ArticleID)

		generated++
		// 更新 cursor
		newCursor := (cursor + generated) % len(topics)
		_ = repository.Set(cursorKey, fmt.Sprintf("%d", newCursor), 0)

		log.Printf("[seed] 分类 %s 已生成第 %d/%d 篇: %s", jobCategory, generated, need, article.Title)
		// 避免频繁调用 AI，间隔 500ms
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("[seed] 分类 %s 补全完成，共生成 %d 篇", jobCategory, generated)
}

// ---------- 1. GenerateAIArticle ----------

func GenerateAIArticle(userID, topic, jobCategory string) (*model.Article, error) {
	limitKey := communityAILimitKey(userID)

	// 先检查当前计数再决定是否继续
	count, err := repository.Incr(limitKey)
	if err != nil {
		return nil, fmt.Errorf("检查 AI 限额失败: %w", err)
	}
	// 首次写入时设置到明天凌晨过期
	if count == 1 {
		tomorrow := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)
		_ = repository.ExpireAt(limitKey, tomorrow)
	}

	limit := config.Cfg.AIDailyLimit
	if limit <= 0 {
		limit = 10
	}
	if count > int64(limit) {
		return nil, errors.New("今日AI生成次数已用完")
	}

	// 调用 AI 生成
	article, err := GenerateArticle(topic, jobCategory)
	if err != nil {
		return nil, fmt.Errorf("AI 生成文章失败: %w", err)
	}

	// 设置 articleId（保留 authorId = "ai"，AI 生成的文章作者标记为 ai）
	article.ArticleID = uuid.New().String()
	// authorId 保持 "ai"（GenerateArticle 已设置），但 userId 作为触发者记录可在上层处理
	_ = userID // userID 用于限额，实际 authorId 保持 "ai"

	now := time.Now().Unix()
	article.CreatedAt = now

	// 写 Hash（永久）
	if err := saveArticleToRedis(article); err != nil {
		return nil, fmt.Errorf("写入文章失败: %w", err)
	}

	score := float64(article.CreatedAt)

	// 写时间 ZSet：分类 + all 聚合（永久不过期）
	timeKeyCategory := articleTimeKey(article.JobCategory)
	timeKeyAll := articleTimeKey("all")
	_ = repository.ZAdd(timeKeyCategory, score, article.ArticleID)
	_ = repository.ZAdd(timeKeyAll, score, article.ArticleID)

	// 写热度 ZSet：分类 + all 聚合（永久不过期）
	hotKeyCategory := articleHotKey(article.JobCategory)
	hotKeyAll := articleHotKey("all")
	_ = repository.ZAdd(hotKeyCategory, 0, article.ArticleID)
	_ = repository.ZAdd(hotKeyAll, 0, article.ArticleID)

	return article, nil
}

// ---------- 2. ListArticles ----------

func ListArticles(jobCategory, sortBy string, page, pageSize int) ([]model.Article, int, error) {
	// 空分类统一用 all
	if jobCategory == "" {
		jobCategory = "all"
	}

	// 异步检查并补全文章（不阻塞本次请求）
	go TriggerSeedIfNeeded(jobCategory)

	var zsetKey string
	if sortBy == "hot" {
		zsetKey = articleHotKey(jobCategory)
	} else {
		zsetKey = articleTimeKey(jobCategory)
	}

	// 取全部 ID（倒序：高分/新文章在前）
	ids, err := repository.ZRevRange(zsetKey, 0, -1)
	if err != nil || len(ids) == 0 {
		return []model.Article{}, 0, nil
	}

	// 批量读取 Hash
	var all []model.Article
	for _, id := range ids {
		m, err := repository.HGetAll(articleKey(id))
		if err != nil || len(m) == 0 {
			continue
		}
		a := &model.Article{}
		if err := a.FromRedisHash(m); err != nil {
			continue
		}
		all = append(all, *a)
	}

	total := len(all)

	// 分页
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	start := (page - 1) * pageSize
	if start >= total {
		return []model.Article{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

// ---------- 3. GetArticle ----------

func GetArticle(articleID string) (*model.Article, error) {
	key := articleKey(articleID)
	m, err := repository.HGetAll(key)
	if err != nil || len(m) == 0 {
		return nil, errors.New("文章不存在")
	}

	a := &model.Article{}
	if err := a.FromRedisHash(m); err != nil {
		return nil, err
	}

	return a, nil
}

// ---------- 4. LikeArticle ----------

func LikeArticle(userID, articleID string) error {
	likesKey := articleLikesKey(articleID)

	// 检查是否已点赞（切换行为）
	isLiked, err := repository.SIsMember(likesKey, userID)
	if err != nil {
		return fmt.Errorf("检查点赞状态失败: %w", err)
	}

	var delta int64
	if isLiked {
		// 取消点赞
		if err := repository.SRem(likesKey, userID); err != nil {
			return fmt.Errorf("取消点赞失败: %w", err)
		}
		delta = -1
	} else {
		// 点赞
		if err := repository.SAdd(likesKey, userID); err != nil {
			return fmt.Errorf("点赞失败: %w", err)
		}
		delta = 1
	}

	// 更新 likeCount
	_, _ = repository.HIncrBy(articleKey(articleID), "likeCount", delta)

	// 更新热度 ZSet
	jobCategory := getArticleJobCategory(articleID)
	if jobCategory != "" {
		_ = repository.ZIncrBy(articleHotKey(jobCategory), float64(delta), articleID)
	}

	return nil
}

// ---------- 5. CollectArticle ----------

func CollectArticle(userID, articleID string) error {
	// 收藏文章（不切换，只添加）
	if err := repository.SAdd(articleCollectsKey(articleID), userID); err != nil {
		return fmt.Errorf("收藏失败: %w", err)
	}

	// 添加到个人收藏列表
	if err := repository.SAdd(userCollectArticleKey(userID), articleID); err != nil {
		return fmt.Errorf("更新个人收藏失败: %w", err)
	}

	// 更新 collectCount
	_, _ = repository.HIncrBy(articleKey(articleID), "collectCount", 1)

	// 更新热度 ZSet
	jobCategory := getArticleJobCategory(articleID)
	if jobCategory != "" {
		_ = repository.ZIncrBy(articleHotKey(jobCategory), 1, articleID)
	}

	return nil
}

// ---------- 6. AddComment ----------

func AddComment(userID, articleID, content string) (*model.Comment, error) {
	// 敏感词过滤
	if containsSensitiveWord(content) {
		return nil, errors.New("评论内容包含敏感词，请修改后重试")
	}

	comment := &model.Comment{
		CommentID: uuid.New().String(),
		ArticleID: articleID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now().Unix(),
	}

	// 写 Hash（永久）
	hash := comment.ToRedisHash()
	if err := repository.HSetMap(commentKey(comment.CommentID), hash); err != nil {
		return nil, fmt.Errorf("写入评论失败: %w", err)
	}

	// LPush 到评论列表（最新在前）
	if err := repository.LPush(articleCommentsKey(articleID), comment.CommentID); err != nil {
		return nil, fmt.Errorf("更新评论列表失败: %w", err)
	}

	// 更新 commentCount
	_, _ = repository.HIncrBy(articleKey(articleID), "commentCount", 1)

	// 更新热度 ZSet
	jobCategory := getArticleJobCategory(articleID)
	if jobCategory != "" {
		_ = repository.ZIncrBy(articleHotKey(jobCategory), 1, articleID)
	}

	return comment, nil
}

// ---------- 7. ListComments ----------

func ListComments(articleID string) ([]model.Comment, error) {
	ids, err := repository.LRange(articleCommentsKey(articleID), 0, -1)
	if err != nil || len(ids) == 0 {
		return []model.Comment{}, nil
	}

	var comments []model.Comment
	for _, id := range ids {
		m, err := repository.HGetAll(commentKey(id))
		if err != nil || len(m) == 0 {
			continue
		}
		c := &model.Comment{}
		if err := c.FromRedisHash(m); err != nil {
			continue
		}
		comments = append(comments, *c)
	}

	if comments == nil {
		comments = []model.Comment{}
	}
	return comments, nil
}

// ---------- 8. GetCollectedArticles ----------

func GetCollectedArticles(userID string) ([]model.Article, error) {
	ids, err := repository.SMembers(userCollectArticleKey(userID))
	if err != nil {
		return nil, err
	}

	var articles []model.Article
	for _, id := range ids {
		m, err := repository.HGetAll(articleKey(id))
		if err != nil || len(m) == 0 {
			continue
		}
		a := &model.Article{}
		if err := a.FromRedisHash(m); err != nil {
			continue
		}
		articles = append(articles, *a)
	}

	if articles == nil {
		articles = []model.Article{}
	}
	return articles, nil
}

// ---------- 辅助：从 Redis 读文章的 jobCategory ----------

func getArticleJobCategory(articleID string) string {
	v, err := repository.HGet(articleKey(articleID), "jobCategory")
	if err != nil {
		return ""
	}
	return v
}
