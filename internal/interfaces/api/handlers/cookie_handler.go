package handlers

import (
	"net/http"
	"time"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/domain"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CookieHandler struct {
	service application.CookieService
}

func NewCookieHandler(service application.CookieService) *CookieHandler {
	return &CookieHandler{service: service}
}

type AddCookieRequest struct {
	Domain    string     `json:"domain" binding:"required"`
	Path      string     `json:"path"`
	Name      string     `json:"name" binding:"required"`
	Value     string     `json:"value" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at"`
	Secure    bool       `json:"secure"`
	HttpOnly  bool       `json:"http_only"`
}

func (h *CookieHandler) AddCookie(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var req AddCookieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	cookie, err := h.service.AddCookie(c.Request.Context(), ownerID, workspaceID, domain.WorkspaceCookie{
		Domain:    req.Domain,
		Path:      req.Path,
		Name:      req.Name,
		Value:     req.Value,
		ExpiresAt: req.ExpiresAt,
		Secure:    req.Secure,
		HttpOnly:  req.HttpOnly,
	})
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Cookie stored successfully", cookie)
}

func (h *CookieHandler) ListCookies(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	cookies, err := h.service.ListCookies(c.Request.Context(), ownerID, workspaceID)
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Cookies retrieved successfully", cookies)
}

func (h *CookieHandler) DeleteCookie(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	cookieIDStr := c.Param("cookieId")
	cookieID, err := uuid.Parse(cookieIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid cookieId UUID")
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	if err := h.service.DeleteCookie(c.Request.Context(), ownerID, workspaceID, cookieID); err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Cookie deleted successfully", gin.H{})
}

func (h *CookieHandler) ClearCookies(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	if err := h.service.ClearCookies(c.Request.Context(), ownerID, workspaceID); err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "All workspace cookies cleared successfully", gin.H{})
}
