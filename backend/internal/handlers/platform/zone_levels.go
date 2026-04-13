package platform

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type zoneLevelOption struct {
	Level       string `json:"level"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// GetZoneLevels returns the configured zone level options.
func GetZoneLevels(c *gin.Context) {
	levels := loadZoneLevelOptions()
	if len(levels) == 0 {
		levels = defaultZoneLevelOptions()
	}
	c.JSON(http.StatusOK, gin.H{"levels": levels})
}

func loadZoneLevelOptions() []zoneLevelOption {
	paths := []string{
		"/root/zone_level.json",
		"/app/zone_level.json",
		"../zone_level.json",
		"zone_level.json",
	}

	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		var payload []zoneLevelOption
		if err := json.NewDecoder(file).Decode(&payload); err == nil {
			_ = file.Close()
			return normalizeZoneLevelOptions(payload)
		}
		_ = file.Close()
	}

	return nil
}

func normalizeZoneLevelOptions(levels []zoneLevelOption) []zoneLevelOption {
	result := make([]zoneLevelOption, 0, len(levels))
	for _, level := range levels {
		normalizedLevel := strings.ToUpper(strings.TrimSpace(level.Level))
		if normalizedLevel == "" {
			continue
		}
		label := strings.TrimSpace(level.Label)
		if label == "" {
			label = normalizedLevel
		}
		result = append(result, zoneLevelOption{
			Level:       normalizedLevel,
			Label:       label,
			Description: strings.TrimSpace(level.Description),
		})
	}
	return result
}

func defaultZoneLevelOptions() []zoneLevelOption {
	return []zoneLevelOption{
		{Level: "A", Label: "A", Description: "Same-city zones"},
		{Level: "B", Label: "B", Description: "Intra-state zones"},
		{Level: "C", Label: "C", Description: "Inter-state zones"},
	}
}
