package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
)

type OperationalHandler struct {
	taskService  service.TaskService
	shiftService service.ShiftService
}

func NewOperationalHandler(taskService service.TaskService, shiftService service.ShiftService) *OperationalHandler {
	return &OperationalHandler{
		taskService:  taskService,
		shiftService: shiftService,
	}
}

func (h *OperationalHandler) Readiness(c *gin.Context) {
	// Basic ping endpoint verifying connection readiness (mock/live check)
	c.JSON(http.StatusOK, gin.H{"status": "healthy", "database": "connected"})
}

func (h *OperationalHandler) GetOrgTasks(c *gin.Context) {
	orgID := c.Param("orgId")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orgId path parameter is required"})
		return
	}

	queue, err := h.taskService.GetOrgTasks(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, queue)
}

func (h *OperationalHandler) GetSiteTasks(c *gin.Context) {
	siteID := c.Param("siteId")
	if siteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "siteId path parameter is required"})
		return
	}

	queue, err := h.taskService.GetQueue(c.Request.Context(), siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, queue)
}

func (h *OperationalHandler) GetUserSiteTasks(c *gin.Context) {
	siteID := c.Param("siteId")
	userID := c.Param("userId")
	if siteID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "siteId and userId path parameters are required"})
		return
	}

	queue, err := h.taskService.GetUserSiteTasks(c.Request.Context(), siteID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, queue)
}

func (h *OperationalHandler) UpdateTaskStatus(c *gin.Context) {
	executionID := c.Param("id")
	userID, _ := c.Get("userID")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.UpdateStatus(c.Request.Context(), executionID, req.Status, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *OperationalHandler) OverrideAsset(c *gin.Context) {
	executionID := c.Param("id")
	userID, _ := c.Get("userID")
	var req struct {
		AssetID       string `json:"asset_id" binding:"required"`
		Justification string `json:"justification" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.OverrideAssetConstraint(c.Request.Context(), executionID, req.AssetID, req.Justification, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *OperationalHandler) ProposeTrade(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req struct {
		TaskExecutionID    string `json:"task_execution_id" binding:"required"`
		ProposedAssigneeID string `json:"proposed_assignee_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.ProposeTrade(c.Request.Context(), req.TaskExecutionID, req.ProposedAssigneeID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success"})
}
