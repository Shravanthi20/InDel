package insurer

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *InsurerHandler) GetPlanUsers(c *gin.Context) {
	users, err := h.Service.ListUserPlanStatuses()
	if err != nil {
		SendError(c, http.StatusInternalServerError, "plan_users_fetch_failed", err.Error(), "")
		return
	}

	SendSuccess(c, gin.H{"users": users})
}

func (h *InsurerHandler) StartUserPlan(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		SendError(c, http.StatusBadRequest, "invalid_user_id", "user id must be a number", "id")
		return
	}

	user, svcErr := h.Service.StartUserPlan(uint(userID))
	if svcErr != nil {
		statusCode := http.StatusInternalServerError
		code := "plan_start_failed"
		if svcErr.Error() == "user not found" {
			statusCode = http.StatusNotFound
			code = "user_not_found"
		}
		SendError(c, statusCode, code, svcErr.Error(), "id")
		return
	}

	SendSuccess(c, gin.H{"success": true, "user": user})
}

func (h *InsurerHandler) EndUserPlan(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		SendError(c, http.StatusBadRequest, "invalid_user_id", "user id must be a number", "id")
		return
	}

	user, svcErr := h.Service.EndUserPlan(uint(userID))
	if svcErr != nil {
		statusCode := http.StatusInternalServerError
		code := "plan_end_failed"
		if svcErr.Error() == "user not found" {
			statusCode = http.StatusNotFound
			code = "user_not_found"
		}
		SendError(c, statusCode, code, svcErr.Error(), "id")
		return
	}

	SendSuccess(c, gin.H{"success": true, "user": user})
}
