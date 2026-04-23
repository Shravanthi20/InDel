package insurer

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetLedger returns high-integrity transaction data
func (h *InsurerHandler) GetLedger(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	data, total, err := h.Service.GetLedger(offset, limit)
	if err != nil {
		SendError(c, 500, "ledger_error", err.Error(), "")
		return
	}

	SendPaginated(c, data, page, limit, int(total))
}
