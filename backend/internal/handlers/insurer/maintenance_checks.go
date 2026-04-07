package insurer

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *InsurerHandler) GetMaintenanceChecks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	data, total, err := h.Service.GetMaintenanceChecks(offset, limit)
	if err != nil {
		SendError(c, 500, "INTERNAL_ERROR", "failed to load maintenance checks", "")
		return
	}

	SendPaginated(c, data, page, limit, int(total))
}

func (h *InsurerHandler) RespondToMaintenanceCheck(c *gin.Context) {
	checkID := c.Param("id")
	body := map[string]string{}
	if err := c.ShouldBindJSON(&body); err != nil {
		SendError(c, 400, "VALIDATION_ERROR", err.Error(), "")
		return
	}

	findings := body["findings"]
	if findings == "" {
		findings = body["response"]
	}
	if findings == "" {
		SendError(c, 400, "VALIDATION_ERROR", "findings_required", "findings")
		return
	}

	if err := h.Service.RespondToMaintenanceCheck(checkID, findings); err != nil {
		SendError(c, 500, "INTERNAL_ERROR", "failed to update maintenance check", "")
		return
	}

	SendSuccess(c, gin.H{"status": "updated", "maintenance_check_id": checkID})
}
