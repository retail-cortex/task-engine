package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
)

type AdminHandler struct {
	adminService service.AdminService
}

func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.adminService.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

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

func (h *AdminHandler) AssignUserToOrganization(c *gin.Context) {
	orgID := c.Param("orgId")
	userID := c.Param("userId")
	if orgID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orgId and userId parameters are required"})
		return
	}
	if err := h.adminService.AssignUserToOrganization(c.Request.Context(), orgID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *AdminHandler) ListOrganizations(c *gin.Context) {
	orgs, err := h.adminService.ListOrganizations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orgs)
}

func (h *AdminHandler) CreateSite(c *gin.Context) {
	orgID := c.Param("orgId")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orgId parameter is required"})
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
