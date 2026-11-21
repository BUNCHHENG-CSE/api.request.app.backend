package handlers

import (
	"net/http"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service application.UserService
}

func NewUserHandler(service application.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) RegisterUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "User created successfully", user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, ok := pathUUID(c, "id", "user")
	if !ok {
		return
	}
	authenticatedID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	if authenticatedID != id {
		response.Error(c, http.StatusForbidden, "you can only access your own user record")
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), id)
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User retrieved successfully", user)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	token, user, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Login successful", gin.H{"token": token, "user": user})
}

func (h *UserHandler) Update(c *gin.Context) {
	id, ok := pathUUID(c, "id", "user")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	if ownerID != id {
		response.Error(c, http.StatusForbidden, "you can only update your own user record")
		return
	}
	var req struct {
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.service.UpdateUser(c.Request.Context(), id, req.Username)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "User updated successfully", user)
}

func (h *UserHandler) Refresh(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	token, err := h.service.RefreshToken(c.Request.Context(), userID)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Token refreshed successfully", gin.H{"token": token})
}

func (h *UserHandler) Logout(c *gin.Context) {
	response.Success(c, http.StatusOK, "Logout successful; discard the bearer token on the client", gin.H{})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	err := h.service.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword, req.ConfirmPassword)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Password changed successfully", gin.H{})
}
