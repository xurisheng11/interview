package handler

import (
	"errors"

	"interview-sim/model"
	"interview-sim/pkg/response"
	"interview-sim/service"

	"github.com/gin-gonic/gin"
)

// GenerateSprintPlan POST /api/v1/learn/plan
func GenerateSprintPlan(c *gin.Context) {
	userId := c.GetString("userId")

	var cfg model.SprintPlanConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	plan, err := service.GenerateSprintPlan(userId, &cfg)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, plan)
}

// GetSprintPlan GET /api/v1/learn/plan
func GetSprintPlan(c *gin.Context) {
	userId := c.GetString("userId")
	plan, err := service.GetSprintPlan(userId)
	if err != nil {
		if errors.Is(err, service.ErrPlanNotFound) {
			response.Success(c, nil) // 未定制计划返回空，前端展示定制表单
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, plan)
}

// UpdateSprintTask PUT /api/v1/learn/plan/task
func UpdateSprintTask(c *gin.Context) {
	userId := c.GetString("userId")

	var body struct {
		PhaseIndex *int `json:"phaseIndex" binding:"required"`
		TaskIndex  *int `json:"taskIndex" binding:"required"`
		Done       bool `json:"done"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	plan, err := service.UpdateSprintTask(userId, *body.PhaseIndex, *body.TaskIndex, body.Done)
	if err != nil {
		if errors.Is(err, service.ErrPlanNotFound) {
			response.NotFound(c, "冲刺计划不存在")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, plan)
}

// DeleteSprintPlan DELETE /api/v1/learn/plan
func DeleteSprintPlan(c *gin.Context) {
	userId := c.GetString("userId")
	if err := service.DeleteSprintPlan(userId); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "计划已删除"})
}
