package router

import (
	"github.com/gin-gonic/gin"
	"interview-sim/handler"
	"interview-sim/middleware"
)

// handlerPlaceholder 临时占位，后续由具体 handler 替换
func handlerPlaceholder(c *gin.Context) {
	c.JSON(200, gin.H{"message": "not implemented yet"})
}

func Setup() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.GET("/me", middleware.AuthRequired(), handler.GetMe)
	}

	// 题库公开路由（不需鉴权）
	api.GET("/questions", handler.ListQuestions)
	api.GET("/questions/:id", handler.GetQuestion)

	protected := api.Group("")
	protected.Use(middleware.AuthRequired())
	{
		interviews := protected.Group("/interviews")
		{
			interviews.POST("", handler.CreateInterview)
			interviews.GET("", handler.GetInterviewList)
			interviews.GET("/:id", handler.GetInterview)
			interviews.POST("/:id/answers", handler.SubmitAnswer)
			interviews.PUT("/:id/pause", handler.PauseInterview)
			interviews.PUT("/:id/complete", handler.CompleteInterview)
		}
		protected.GET("/reports/:interviewId", handler.GetReport)
		protected.POST("/reports/:interviewId/share", handler.CreateShare)

		// 题库需鉴权路由
		protected.POST("/questions/:id/practice", handler.PracticeQuestion)
		protected.POST("/questions/:id/collect", handler.CollectQuestion)
		protected.DELETE("/questions/:id/collect", handler.UncollectQuestion)

		// 题库管理员路由
		adminQuestions := protected.Group("/questions")
		adminQuestions.Use(middleware.AdminRequired())
		{
			adminQuestions.POST("", handler.CreateQuestion)
			adminQuestions.PUT("/:id", handler.UpdateQuestion)
			adminQuestions.DELETE("/:id", handler.DeleteQuestion)
		}

		community := protected.Group("/community")
		{
			community.GET("/articles", handler.ListArticles)
			community.POST("/articles/ai", handler.GenerateAIArticle)
			community.GET("/articles/:id", handler.GetArticle)
			community.POST("/articles/:id/like", handler.LikeArticle)
			community.POST("/articles/:id/collect", handler.CollectArticle)
			community.GET("/articles/:id/comments", handler.ListComments)
			community.POST("/articles/:id/comments", handler.AddComment)
		}
		profile := protected.Group("/profile")
		{
			profile.GET("", handler.GetProfile)
			profile.PUT("", handler.UpdateProfile)
			profile.PUT("/password", handler.ChangePassword)
			profile.GET("/stats", handler.GetStats)
			profile.GET("/trend", handler.GetScoreTrend)
			profile.GET("/collections", handler.GetCollections)
		}

		// 公司面试情报
		protected.GET("/company/intel", handler.GetCompanyIntel)
		protected.GET("/company/question-answer", handler.GetCompanyQuestionAnswer)

		// 管理后台
		admin := protected.Group("/admin")
		admin.Use(middleware.AdminRequired())
		{
			admin.GET("/users", handler.AdminListUsers)
			admin.GET("/users/:id", handler.AdminGetUser)
			admin.PUT("/users/:id/password", handler.AdminResetPassword)
			admin.PUT("/users/:id/role", handler.AdminSetRole)
			admin.DELETE("/users/:id", handler.AdminDeleteUser)
		}
	}
	api.GET("/reports/share/:token", handler.GetSharedReport)

	return r
}
