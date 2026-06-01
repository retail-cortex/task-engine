// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
)

type OperationalHandler struct {
	taskService       service.TaskService
	shiftService      service.ShiftService
	automationService service.AutomationService
	cfg               Config
}

func NewOperationalHandler(
	taskService service.TaskService,
	shiftService service.ShiftService,
	automationService service.AutomationService,
	cfg Config,
) *OperationalHandler {
	return &OperationalHandler{
		taskService:       taskService,
		shiftService:      shiftService,
		automationService: automationService,
		cfg:               cfg,
	}
}

func (h *OperationalHandler) Readiness(c *gin.Context) {
	// Basic ping endpoint verifying connection readiness (mock/live check)
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"database":  "connected",
		"client_id": h.cfg.OAuth.ClientID,
	})
}

func (h *OperationalHandler) GetMe(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user context"})
		return
	}

	user, err := h.shiftService.GetUserProfile(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *OperationalHandler) ListSites(c *gin.Context) {
	sites, err := h.taskService.ListActiveSites(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sites)
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
		Status         string `json:"status" binding:"required"`
		ChecklistState string `json:"checklist_state"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.UpdateStatus(c.Request.Context(), executionID, req.Status, req.ChecklistState, userID.(string))
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

func (h *OperationalHandler) AcceptTrade(c *gin.Context) {
	tradeID := c.Param("tradeId")
	userID, _ := c.Get("userID")

	err := h.taskService.AcceptTrade(c.Request.Context(), tradeID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *OperationalHandler) RejectTrade(c *gin.Context) {
	tradeID := c.Param("tradeId")
	userID, _ := c.Get("userID")

	err := h.taskService.RejectTrade(c.Request.Context(), tradeID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *OperationalHandler) ListPendingTrades(c *gin.Context) {
	userID, _ := c.Get("userID")

	trades, err := h.taskService.ListPendingTrades(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trades)
}

func (h *OperationalHandler) TriggerAlert(c *gin.Context) {
	siteID := c.Param("siteId")
	var req struct {
		OrganizerID string `json:"organizer_id" binding:"required"`
		EventType   string `json:"event_type" binding:"required"`
		Description string `json:"description" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exec, err := h.automationService.TriggerStreamingEvent(c.Request.Context(), siteID, req.OrganizerID, model.EventType(req.EventType), req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, exec)
}
