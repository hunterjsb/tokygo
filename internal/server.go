package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

// Config holds server configuration
type Config struct {
	Resolution int
	RadiusKm   float64
}

// Server handles HTTP requests
type Server struct {
	RootDir      string
	Config       Config
	geojsonCache *GeoJSON
	cacheMutex   sync.RWMutex
	cacheTime    time.Time
}

// NewServer creates a new server instance
func NewServer(rootDir string) *Server {
	return &Server{
		RootDir: rootDir,
		Config: Config{
			Resolution: 7,
			RadiusKm:   15.0,
		},
	}
}

// RegisterHandlers sets up all HTTP routes
func (s *Server) RegisterHandlers() {
	// API endpoints
	http.HandleFunc("/api/cities", s.handleCities)
	http.HandleFunc("/api/geojson", s.handleGeoJSON)
	http.HandleFunc("/api/config", s.handleConfig)

	// Custom handler for static files that doesn't catch /api routes
	frontendDir := filepath.Join(s.RootDir, "frontend")
	fs := http.FileServer(http.Dir(frontendDir))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}

// handleCities returns the cities data as JSON
func (s *Server) handleCities(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"cities": Cities,
		"colors": CityColors,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGeoJSON generates and serves GeoJSON with caching
func (s *Server) handleGeoJSON(w http.ResponseWriter, r *http.Request) {
	// Check if we have cached data (cache for 5 minutes)
	s.cacheMutex.RLock()
	if s.geojsonCache != nil && time.Since(s.cacheTime) < 5*time.Minute {
		cached := s.geojsonCache
		s.cacheMutex.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(cached)
		return
	}
	s.cacheMutex.RUnlock()

	// Generate new GeoJSON
	geojson, err := GenerateGeoJSON(s.Config.Resolution, s.Config.RadiusKm)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating GeoJSON: %v", err), http.StatusInternalServerError)
		return
	}

	// Update cache
	s.cacheMutex.Lock()
	s.geojsonCache = geojson
	s.cacheTime = time.Now()
	s.cacheMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(geojson)
}

// handleConfig returns the server configuration
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"resolution": s.Config.Resolution,
		"radiusKm":   s.Config.RadiusKm,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
