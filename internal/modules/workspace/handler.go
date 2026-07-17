package workspace

import (
	"backend/internal/core/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler holds dependencies for the HTTP layer
type Handler struct {
	service Service
}

// RegisterRoutes sets up the Gin routes for this module
func RegisterRoutes(r *gin.Engine, s Service) {
	h := &Handler{service: s}

	// Create a route group for workspaces
	routes := r.Group("/api/v1/workspaces")
	{
		routes.POST("/", h.Create)
		routes.GET("/:id", h.Get)
	}
}

// Create handles the POST request to create a workspace
func (h *Handler) Create(c *gin.Context) {
	var req CreateWorkspaceRequest

	// Bind JSON body to struct and validate tags
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Retrieve user ID injected by the Auth middleware
	ownerID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Call the service
	workspace, err := h.service.CreateWorkspace(req, ownerID.(string))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create workspace")
		return
	}

	response.Success(c, http.StatusCreated, workspace, "Workspace created successfully")
}

// Get handles the GET request to retrieve a workspace by ID
func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")

	workspace, err := h.service.GetWorkspace(id)
	if err != nil {
		if err.Error() == "workspace not found" {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve workspace")
		return
	}

	response.Success(c, http.StatusOK, workspace, "Get workspace successfully")
}
