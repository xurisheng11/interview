package service

import (
	"errors"
	"fmt"
	"strings"
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

// ---------- 文章写入辅助 ----------

// saveArticleToRedis 将文章写入 Redis Hash
func saveArticleToRedis(a *model.Article) error {
	hash := a.ToRedisHash()
	return repository.HSetMap(articleKey(a.ArticleID), hash)
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
