package platform

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

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
				c.JSON(http.StatusOK, gin.H{"cities": data})
				return
			} else {
				c.JSON(http.StatusOK, gin.H{"city_pairs": data})
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
		c.JSON(http.StatusOK, gin.H{"cities": cities})
		return
	}
	if typeParam == "b" {
		pairs := []gin.H{}
		for i, c1 := range cityStateList {
			for j, c2 := range cityStateList {
				if i != j && c1.State == c2.State {
					pairs = append(pairs, gin.H{"from": c1.City, "to": c2.City, "state": c1.State})
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"city_pairs": pairs})
		return
	}
	if typeParam == "c" {
		pairs := []gin.H{}
		for i, c1 := range cityStateList {
			for j, c2 := range cityStateList {
				if i != j && c1.State != c2.State {
					pairs = append(pairs, gin.H{"from": c1.City, "to": c2.City, "from_state": c1.State, "to_state": c2.State})
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"city_pairs": pairs})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_type"})
}
