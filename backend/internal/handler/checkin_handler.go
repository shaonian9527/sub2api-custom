package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// CheckinHandler handles daily check-in requests.
type CheckinHandler struct {
	redeemService *service.RedeemService
}

// NewCheckinHandler creates a new CheckinHandler.
func NewCheckinHandler(redeemService *service.RedeemService) *CheckinHandler {
	return &CheckinHandler{redeemService: redeemService}
}

// GetStatus returns today's check-in status.
// GET /api/v1/user/checkin/status
func (h *CheckinHandler) GetStatus(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	status, err := h.redeemService.GetCheckinStatus(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, status)
}

// Checkin performs today's daily check-in.
// POST /api/v1/user/checkin
func (h *CheckinHandler) Checkin(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	result, err := h.redeemService.Checkin(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, result)
}
