package core

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/apiutil"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type syntheticRequest struct {
	Seed      int    `json:"seed"`
	Scenario  string `json:"scenario"`
	OutputDir string `json:"output_dir"`
}

func QueueClaimPayout(c *gin.Context) {
	if !hasDB() {
		apiutil.SendError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "db unavailable", "")
		return
	}

	claimID, err := parseUintParam(c.Param("claim_id"), "clm_")
	if err != nil {
		apiutil.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid claim id", "claim_id")
		return
	}

	result, err := coreOps.QueueClaimPayout(uint(claimID))
	if err != nil {
		status := http.StatusInternalServerError
		code := "INTERNAL_ERROR"
		field := ""
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
			code = "NOT_FOUND"
			field = "claim_id"
		}
		apiutil.SendError(c, status, code, err.Error(), field)
		return
	}

	apiutil.SendSuccess(c, http.StatusAccepted, result)
}

func RunWeeklyCycle(c *gin.Context) {
	if !hasDB() {
		apiutil.SendError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "db unavailable", "")
		return
	}

	result, err := coreOps.RunWeeklyCycle(time.Now().UTC())
	if err != nil {
		apiutil.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), "")
		return
	}
	apiutil.SendSuccess(c, http.StatusOK, result)
}

func GenerateClaimsForDisruption(c *gin.Context) {
	if !hasDB() {
		apiutil.SendError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "db unavailable", "")
		return
	}

	disruptionID, err := parseUintParam(c.Param("disruption_id"), "dis_")
	if err != nil {
		apiutil.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid disruption id", "disruption_id")
		return
	}

	result, err := coreOps.GenerateClaimsForDisruption(uint(disruptionID), time.Now().UTC())
	if err != nil {
		status := http.StatusInternalServerError
		code := "INTERNAL_ERROR"
		field := ""
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
			code = "NOT_FOUND"
			field = "disruption_id"
		}
		apiutil.SendError(c, status, code, err.Error(), field)
		return
	}

	apiutil.SendSuccess(c, http.StatusOK, result)
}

func ProcessPayouts(c *gin.Context) {
	if !hasDB() {
		apiutil.SendError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "db unavailable", "")
		return
	}

	result, err := coreOps.ProcessQueuedPayouts(time.Now().UTC())
	if err != nil {
		apiutil.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), "")
		return
	}
	apiutil.SendSuccess(c, http.StatusOK, result)
}

func GetPayoutReconciliation(c *gin.Context) {
	if !hasDB() {
		apiutil.SendError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "db unavailable", "")
		return
	}

	from, err := parseDateQuery(c.DefaultQuery("from", time.Now().UTC().AddDate(0, 0, -7).Format("2006-01-02")))
	if err != nil {
		apiutil.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid from date", "from")
		return
	}
	to, err := parseDateQuery(c.DefaultQuery("to", time.Now().UTC().Format("2006-01-02")))
	if err != nil {
		apiutil.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid to date", "to")
		return
	}

	result, err := coreOps.GetPayoutReconciliation(from, to.Add(23*time.Hour+59*time.Minute+59*time.Second))
	if err != nil {
		apiutil.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), "")
		return
	}
	apiutil.SendSuccess(c, http.StatusOK, result)
}

func GenerateSyntheticData(c *gin.Context) {
	if !hasDB() {
		apiutil.SendError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "db unavailable", "")
		return
	}

	var req syntheticRequest
	if err := c.ShouldBindJSON(&req); err != nil && !strings.Contains(err.Error(), "EOF") {
		apiutil.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "body")
		return
	}

	result, err := coreOps.GenerateSyntheticData(services.SyntheticGenerateRequest{Seed: req.Seed, Scenario: req.Scenario, OutputDir: req.OutputDir}, time.Now().UTC())
	if err != nil {
		apiutil.SendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), "")
		return
	}
	apiutil.SendSuccess(c, http.StatusOK, result)
}

func parseUintParam(raw string, prefix string) (uint64, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, prefix)
	trimmed = strings.TrimPrefix(trimmed, strings.ReplaceAll(prefix, "_", "-"))
	return strconv.ParseUint(trimmed, 10, 64)
}

func parseDateQuery(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(raw))
}
