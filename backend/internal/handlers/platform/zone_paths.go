package platform

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

const zonePathLimit = 10

// CityState is a struct for city/state pairs
var cityStateList = []struct {
	City  string
	State string
}{
	{"Bangalore", "Karnataka"},
	{"Mumbai", "Maharashtra"},
	{"Chennai", "Tamil Nadu"},
	{"Delhi", "Delhi"},
}

// GetZonePaths returns cities or city pairs for zone types a, b, c
func GetZonePaths(c *gin.Context) {
	typeParam := strings.ToLower(c.Query("type"))
	if hasDB() {
		type zoneRow struct {
			ID    uint   `gorm:"column:id"`
			Name  string `gorm:"column:name"`
			City  string `gorm:"column:city"`
			State string `gorm:"column:state"`
			Level string `gorm:"column:level"`
		}

		rows := make([]zoneRow, 0)
		query := platformDB.Table("zones").Select("id, name, city, state, level")
		switch typeParam {
		case "a":
			query = query.Where("LOWER(COALESCE(level, 'b')) = 'a'")
		case "b":
			query = query.Where("LOWER(COALESCE(level, 'b')) = 'b'")
		case "c":
			query = query.Where("LOWER(COALESCE(level, 'b')) = 'c'")
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_type"})
			return
		}

		if err := query.Order("city ASC, name ASC").Scan(&rows).Error; err == nil {
			if typeParam == "a" {
				cities := make([]string, 0, len(rows))
				seen := map[string]struct{}{}
				for _, row := range rows {
					city := strings.TrimSpace(row.City)
					if city == "" {
						city = strings.TrimSpace(row.Name)
					}
					if city == "" {
						continue
					}
					key := strings.ToLower(city)
					if _, ok := seen[key]; ok {
						continue
					}
					seen[key] = struct{}{}
					cities = append(cities, city)
				}
				if len(cities) > zonePathLimit {
					cities = cities[:zonePathLimit]
				}
				c.JSON(http.StatusOK, gin.H{"cities": cities})
				return
			}

			zones := make([]gin.H, 0, len(rows))
			seen := map[string]struct{}{}
			for _, row := range rows {
				zoneName := strings.TrimSpace(row.Name)
				if zoneName == "" {
					zoneName = strings.TrimSpace(row.City)
				}
				zoneState := strings.TrimSpace(row.State)
				if zoneName == "" || zoneState == "" {
					continue
				}
				key := strings.ToLower(zoneState + "|" + zoneName)
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				zones = append(zones, gin.H{
					"zone_id":    row.ID,
					"zone_name":  zoneName,
					"zone_state": zoneState,
					"city":       strings.TrimSpace(row.City),
					"level":      strings.ToUpper(strings.TrimSpace(row.Level)),
				})
			}

			if len(zones) > zonePathLimit {
				zones = zones[:zonePathLimit]
			}
			c.JSON(http.StatusOK, gin.H{"zones": zones})
			return
		}
	}
	var fileName string
	switch typeParam {
	case "a":
		fileName = "/root/zone_a.json"
	case "b":
		fileName = "/root/zone_b.json"
	case "c":
		fileName = "/root/zone_c.json"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_type"})
		return
	}

	f, err := os.Open(fileName)
	if err == nil {
		defer f.Close()
		var data interface{}
		if err := json.NewDecoder(f).Decode(&data); err == nil {
			if typeParam == "a" {
				if cities, ok := data.([]any); ok && len(cities) > zonePathLimit {
					data = cities[:zonePathLimit]
				}
				c.JSON(http.StatusOK, gin.H{"cities": data})
				return
			} else {
				if pairs, ok := data.([]any); ok && len(pairs) > zonePathLimit {
					data = pairs[:zonePathLimit]
				}
				c.JSON(http.StatusOK, gin.H{"zones": data})
				return
			}
		}
	}
	// fallback to old logic if file not found or decode fails
	if typeParam == "a" {
		cities := make([]string, 0, len(cityStateList))
		for _, cs := range cityStateList {
			cities = append(cities, cs.City)
		}
		sort.Strings(cities)
		if len(cities) > zonePathLimit {
			cities = cities[:zonePathLimit]
		}
		c.JSON(http.StatusOK, gin.H{"cities": cities})
		return
	}
	if typeParam == "b" {
		zones := []gin.H{}
		for _, c1 := range cityStateList {
			zones = append(zones, gin.H{"zone_name": c1.City, "zone_state": c1.State, "city": c1.City})
		}
		if len(zones) > zonePathLimit {
			zones = zones[:zonePathLimit]
		}
		c.JSON(http.StatusOK, gin.H{"zones": zones})
		return
	}
	if typeParam == "c" {
		zones := []gin.H{}
		for _, c1 := range cityStateList {
			zones = append(zones, gin.H{"zone_name": c1.City, "zone_state": c1.State, "city": c1.City})
		}
		if len(zones) > zonePathLimit {
			zones = zones[:zonePathLimit]
		}
		c.JSON(http.StatusOK, gin.H{"zones": zones})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_type"})
}
