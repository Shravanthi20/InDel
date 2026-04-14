package platform

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
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
				// Always return a list of objects for zone A
				cities := make([]gin.H, 0, len(rows))
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
					// Try to get geo info if available
					geo := struct {
						State    string
						Lat, Lon float64
					}{row.State, 0, 0}
					if cityGeoMap != nil {
						if g, ok := cityGeoMap[city]; ok {
							geo = g
						}
					}
					cities = append(cities, gin.H{
						"city":  city,
						"state": geo.State,
						"lat":   geo.Lat,
						"lon":   geo.Lon,
					})
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
		fileName = "zone_a.json"
	case "b":
		fileName = "zone_b.json"
	case "c":
		fileName = "zone_c.json"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_type"})
		return
	}

	if typeParam == "a" {
		// Load city geo lookup once
		cityGeoOnce.Do(func() {
			cityGeoMap, cityGeoErr = loadCityGeo("Indian Cities Geo Data.csv")
		})
		f, err := os.Open(fileName)
		if err == nil && cityGeoErr == nil {
			defer f.Close()
			var data []string
			if err := json.NewDecoder(f).Decode(&data); err == nil {
				limit := zonePathLimit
				if len(data) < limit {
					limit = len(data)
				}
				cities := data[:limit]
				result := make([]gin.H, 0, len(cities))
				for _, city := range cities {
					geo, ok := cityGeoMap[city]
					if !ok {
						geo = struct {
							State    string
							Lat, Lon float64
						}{"Unknown", 0, 0}
					}
					result = append(result, gin.H{
						"city":  city,
						"state": geo.State,
						"lat":   geo.Lat,
						"lon":   geo.Lon,
					})
				}
				c.JSON(http.StatusOK, gin.H{"cities": result})
				return
			}
		}
		// fallback to old logic if file not found or decode fails
		cities := make([]gin.H, 0, len(cityStateList))
		for _, cs := range cityStateList {
			cities = append(cities, gin.H{"city": cs.City, "state": cs.State, "lat": 0, "lon": 0})
		}
		sort.Slice(cities, func(i, j int) bool { return fmt.Sprint(cities[i]["city"]) < fmt.Sprint(cities[j]["city"]) })
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
