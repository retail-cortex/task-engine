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
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
)

type AdminHandler struct {
	adminService     service.AdminService
	schedulerService service.SchedulerService
	taskService      service.TaskService
	shiftService     service.ShiftService
	ragService       service.RAGService
}

func NewAdminHandler(
	adminService service.AdminService,
	schedulerService service.SchedulerService,
	taskService service.TaskService,
	shiftService service.ShiftService,
	ragService service.RAGService,
) *AdminHandler {
	return &AdminHandler{
		adminService:     adminService,
		schedulerService: schedulerService,
		taskService:      taskService,
		shiftService:     shiftService,
		ragService:       ragService,
	}
}

func parseOffsetLimit(c *gin.Context) (int, int, bool) {
	offsetStr := c.Query("offset")
	limitStr := c.Query("limit")
	if offsetStr == "" && limitStr == "" {
		return 0, 0, false
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	return offset, limit, true
}

// --- User Handlers ---

func (h *AdminHandler) ListUsers(c *gin.Context) {
	if offset, limit, ok := parseOffsetLimit(c); ok {
		users, err := h.adminService.ListUsersRange(c.Request.Context(), offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var dtos []UserDTO
		for _, u := range users {
			dtos = append(dtos, toUserDTO(u))
		}
		c.JSON(http.StatusOK, dtos)
		return
	}
	users, err := h.adminService.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var dtos []UserDTO
	for _, u := range users {
		dtos = append(dtos, toUserDTO(u))
	}
	c.JSON(http.StatusOK, dtos)
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.adminService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.adminService.RegisterUser(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.adminService.UpdateUser(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.adminService.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// --- Role Handlers ---

func (h *AdminHandler) CreateRole(c *gin.Context) {
	var role model.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.adminService.CreateRole(c.Request.Context(), &role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (h *AdminHandler) GetRole(c *gin.Context) {
	id := c.Param("id")
	role, err := h.adminService.GetRoleByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (h *AdminHandler) UpdateRole(c *gin.Context) {
	var role model.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.adminService.UpdateRole(c.Request.Context(), &role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (h *AdminHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if err := h.adminService.DeleteRole(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *AdminHandler) ListRoles(c *gin.Context) {
	if offset, limit, ok := parseOffsetLimit(c); ok {
		roles, err := h.adminService.ListRolesRange(c.Request.Context(), offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, roles)
		return
	}
	roles, err := h.adminService.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (h *AdminHandler) AssignRole(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		RoleID string `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.adminService.AssignRole(c.Request.Context(), userID, req.RoleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// --- Organization Handlers ---

func (h *AdminHandler) CreateOrganization(c *gin.Context) {
	var org model.Organization
	if err := c.ShouldBindJSON(&org); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.adminService.RegisterOrganization(c.Request.Context(), &org); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, org)
}

func (h *AdminHandler) GetOrganization(c *gin.Context) {
	id := c.Param("id")
	org, err := h.adminService.GetOrganizationByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, org)
}

func (h *AdminHandler) UpdateOrganization(c *gin.Context) {
	var org model.Organization
	if err := c.ShouldBindJSON(&org); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.adminService.UpdateOrganization(c.Request.Context(), &org); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, org)
}

func (h *AdminHandler) DeleteOrganization(c *gin.Context) {
	id := c.Param("id")
	if err := h.adminService.DeleteOrganization(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *AdminHandler) ListOrganizations(c *gin.Context) {
	if offset, limit, ok := parseOffsetLimit(c); ok {
		orgs, err := h.adminService.ListOrganizationsRange(c.Request.Context(), offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, orgs)
		return
	}
	orgs, err := h.adminService.ListOrganizations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orgs)
}

func (h *AdminHandler) AssignUserToOrganization(c *gin.Context) {
	orgID := c.Param("id")
	userID := c.Param("userId")
	if orgID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id and userId parameters are required"})
		return
	}
	if err := h.adminService.AssignUserToOrganization(c.Request.Context(), orgID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// --- Site Handlers ---

func (h *AdminHandler) CreateSite(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}
	var site model.Site
	if err := c.ShouldBindJSON(&site); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	site.OrganizationID = orgID
	if err := h.adminService.RegisterSite(c.Request.Context(), &site); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, site)
}

func (h *AdminHandler) GetSite(c *gin.Context) {
	id := c.Param("id")
	site, err := h.adminService.GetSiteByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, site)
}

func (h *AdminHandler) UpdateSite(c *gin.Context) {
	var site model.Site
	if err := c.ShouldBindJSON(&site); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.adminService.UpdateSite(c.Request.Context(), &site); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, site)
}

func (h *AdminHandler) DeleteSite(c *gin.Context) {
	id := c.Param("id")
	if err := h.adminService.DeleteSite(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *AdminHandler) ListSites(c *gin.Context) {
	if offset, limit, ok := parseOffsetLimit(c); ok {
		sites, err := h.adminService.ListSitesRange(c.Request.Context(), offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, sites)
		return
	}
	sites, err := h.adminService.ListSites(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sites)
}

// --- Location Handlers ---

func (h *AdminHandler) CreateLocation(c *gin.Context) {
	siteID := c.Param("siteId")
	if siteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "siteId parameter is required"})
		return
	}
	var loc model.Location
	if err := c.ShouldBindJSON(&loc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	loc.SiteID = siteID
	if err := h.adminService.RegisterLocation(c.Request.Context(), &loc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, loc)
}

func (h *AdminHandler) GetLocation(c *gin.Context) {
	id := c.Param("id")
	loc, err := h.adminService.GetLocationByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loc)
}

func (h *AdminHandler) UpdateLocation(c *gin.Context) {
	var loc model.Location
	if err := c.ShouldBindJSON(&loc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.adminService.UpdateLocation(c.Request.Context(), &loc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loc)
}

func (h *AdminHandler) DeleteLocation(c *gin.Context) {
	id := c.Param("id")
	if err := h.adminService.DeleteLocation(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *AdminHandler) ListLocations(c *gin.Context) {
	if offset, limit, ok := parseOffsetLimit(c); ok {
		locs, err := h.adminService.ListLocationsRange(c.Request.Context(), offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, locs)
		return
	}
	locs, err := h.adminService.ListLocations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, locs)
}

// --- Asset Handlers ---

func (h *AdminHandler) CreateAsset(c *gin.Context) {
	locationID := c.Param("locationId")
	if locationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "locationId parameter is required"})
		return
	}
	var asset model.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	asset.LocationID = locationID
	if err := h.adminService.RegisterAsset(c.Request.Context(), &asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, asset)
}

func (h *AdminHandler) GetAsset(c *gin.Context) {
	id := c.Param("id")
	asset, err := h.adminService.GetAssetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, asset)
}

func (h *AdminHandler) UpdateAsset(c *gin.Context) {
	var asset model.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.adminService.UpdateAsset(c.Request.Context(), &asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, asset)
}

func (h *AdminHandler) DeleteAsset(c *gin.Context) {
	id := c.Param("id")
	if err := h.adminService.DeleteAsset(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *AdminHandler) ListAssets(c *gin.Context) {
	if offset, limit, ok := parseOffsetLimit(c); ok {
		assets, err := h.adminService.ListAssetsRange(c.Request.Context(), offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, assets)
		return
	}
	assets, err := h.adminService.ListAssets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assets)
}

// --- Task Handlers ---

func (h *AdminHandler) CreateTaskTemplate(c *gin.Context) {
	var task model.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.adminService.CreateTaskTemplate(c.Request.Context(), &task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *AdminHandler) GetTaskTemplate(c *gin.Context) {
	id := c.Param("id")
	task, err := h.adminService.GetTaskTemplateByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *AdminHandler) UpdateTaskTemplate(c *gin.Context) {
	var task model.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.adminService.UpdateTaskTemplate(c.Request.Context(), &task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *AdminHandler) DeleteTaskTemplate(c *gin.Context) {
	id := c.Param("id")
	if err := h.adminService.DeleteTaskTemplate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *AdminHandler) ListTaskTemplates(c *gin.Context) {
	if offset, limit, ok := parseOffsetLimit(c); ok {
		tasks, err := h.adminService.ListTaskTemplatesRange(c.Request.Context(), offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tasks)
		return
	}
	tasks, err := h.adminService.ListTaskTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// --- Scheduler & Operational Controls ---

func (h *AdminHandler) TriggerSchedulerSweep(c *gin.Context) {
	err := h.schedulerService.ForceTriggerBatchSweep(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Scheduled batch sweeps materialized successfully."})
}

func (h *AdminHandler) GetSchedulerStatus(c *gin.Context) {
	status := h.schedulerService.GetStatus()
	c.JSON(http.StatusOK, status)
}

// --- Task Execution Handlers ---

func (h *AdminHandler) ListTaskExecutions(c *gin.Context) {
	if offset, limit, ok := parseOffsetLimit(c); ok {
		execs, err := h.taskService.ListTaskExecutionsRange(c.Request.Context(), offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, execs)
		return
	}
	execs, err := h.taskService.ListTaskExecutions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, execs)
}

func (h *AdminHandler) GetTaskExecution(c *gin.Context) {
	id := c.Param("id")
	exec, err := h.taskService.GetTaskExecutionByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, exec)
}

func (h *AdminHandler) DeleteTaskExecution(c *gin.Context) {
	id := c.Param("id")
	if err := h.taskService.DeleteTaskExecution(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// --- Shift Agent Session Handlers ---

func (h *AdminHandler) ListShiftSessions(c *gin.Context) {
	if offset, limit, ok := parseOffsetLimit(c); ok {
		sess, err := h.shiftService.ListSessionsRange(c.Request.Context(), offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, sess)
		return
	}
	sess, err := h.shiftService.ListSessions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sess)
}

func (h *AdminHandler) GetShiftSession(c *gin.Context) {
	id := c.Param("id")
	sess, err := h.shiftService.GetSessionByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sess)
}

func (h *AdminHandler) DeleteShiftSession(c *gin.Context) {
	id := c.Param("id")
	if err := h.shiftService.DeleteSession(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// --- SOP RAG Handlers ---

func (h *AdminHandler) ListSOPs(c *gin.Context) {
	if offset, limit, ok := parseOffsetLimit(c); ok {
		sops, err := h.ragService.ListSOPsRange(c.Request.Context(), offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, sops)
		return
	}
	sops, err := h.ragService.ListSOPs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sops)
}

func (h *AdminHandler) GetSOP(c *gin.Context) {
	id := c.Param("id")
	sop, err := h.ragService.GetSOPByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sop)
}

func (h *AdminHandler) DeleteSOP(c *gin.Context) {
	id := c.Param("id")
	if err := h.ragService.DeleteSOP(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// --- SOP Process Handlers ---

func (h *AdminHandler) ListProcesses(c *gin.Context) {
	if offset, limit, ok := parseOffsetLimit(c); ok {
		procs, err := h.ragService.ListProcessesRange(c.Request.Context(), offset, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, procs)
		return
	}
	procs, err := h.ragService.ListProcesses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, procs)
}

func (h *AdminHandler) GetProcess(c *gin.Context) {
	id := c.Param("id")
	proc, err := h.ragService.GetProcessByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, proc)
}

func (h *AdminHandler) DeleteProcess(c *gin.Context) {
	id := c.Param("id")
	if err := h.ragService.DeleteProcess(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
