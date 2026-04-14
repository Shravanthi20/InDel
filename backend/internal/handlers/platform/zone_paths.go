package platform

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const zonePathLimit = 15

// --- City geo lookup (from CSV, loaded once) ---
var (
	cityGeoOnce sync.Once
	cityGeoMap  map[string]struct {
		State string
		Lat   float64
		Lon   float64
	}
	cityGeoErr error
)

func loadCityGeo(csvPath string) (map[string]struct {
	State    string
	Lat, Lon float64
}, error) {
	m := make(map[string]struct {
		State    string
		Lat, Lon float64
	})
	f, err := os.Open(csvPath)
	if err != nil {
		return m, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return m, err
	}
	if len(records) < 1 {
		return m, fmt.Errorf("empty csv")
	}
	header := make(map[string]int)
	for i, col := range records[0] {
		header[strings.ToLower(col)] = i
	}
	for _, row := range records[1:] {
		city := strings.TrimSpace(strings.Split(row[header["location"]], " Latitude")[0])
		state := strings.TrimSpace(row[header["state"]])
		lat := 0.0
		lon := 0.0
		if idx, ok := header["latitude"]; ok {
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(row[idx]), 64); err == nil {
				lat = parsed
			}
		}
		if idx, ok := header["longitude"]; ok {
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(row[idx]), 64); err == nil {
				lon = parsed
			}
		}
		m[city] = struct {
			State    string
			Lat, Lon float64
		}{state, lat, lon}
	}
	return m, nil
}

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
		fileName = "zone_a.json"
	case "b":
		fileName = "zone_b.json"
	case "c":
		fileName = "zone_c.json"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_type"})
		return
	}

	f, err := os.Open(fileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "zone_file_not_found", "file": fileName})
		return
	}
	defer f.Close()

	// For all types, decode as []map[string]interface{} (generic JSON array of objects)
	var data []map[string]interface{}
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "zone_file_decode_failed", "file": fileName})
		return
	}
	limit := zonePathLimit
	if len(data) < limit {
		limit = len(data)
	}
	result := data[:limit]

	if typeParam == "a" {
		// Map 'name' to 'city' for each entry
		for _, entry := range result {
			if name, ok := entry["name"]; ok {
				entry["city"] = name
			}
		}
		c.JSON(http.StatusOK, gin.H{"cities": result})
	} else {
		c.JSON(http.StatusOK, gin.H{"zones": result})
	}
}
