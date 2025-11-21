package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

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
	// Health check endpoint (no CORS needed for health checks)
	http.HandleFunc("/health", s.handleHealth)

	// API endpoints with CORS support
	http.HandleFunc("/api/cities", corsMiddleware(s.handleCities))
	http.HandleFunc("/api/config", corsMiddleware(s.handleConfig))
	http.HandleFunc("/api/routes", corsMiddleware(s.handleRoutes))

	http.HandleFunc("/api/routes/lines", corsMiddleware(s.handleRoutesLines))
	http.HandleFunc("/api/mapbox/directions", corsMiddleware(s.handleMapboxDirections))
	http.HandleFunc("/api/mapbox/geocoding", corsMiddleware(s.handleMapboxGeocoding))

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

// handleConfig returns the server configuration
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"resolution": s.Config.Resolution,
		"radiusKm":   s.Config.RadiusKm,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleRoutes returns the list of routes
func (s *Server) handleRoutes(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"routes": TripRoutes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRoutesLines returns route lines as GeoJSON LineStrings
func (s *Server) handleRoutesLines(w http.ResponseWriter, r *http.Request) {
	features := []Feature{}

	for _, route := range CachedRoutes {
		feature := Feature{
			Type: "Feature",
			Geometry: Geometry{
				Type:        "LineString",
				Coordinates: route.Geometry,
			},
			Properties: map[string]interface{}{
				"route_name": route.Name,
				"route_type": route.Type,
				"distance":   route.Distance,
				"duration":   route.Duration,
			},
		}
		features = append(features, feature)
	}

	geojson := &GeoJSON{
		Type:     "FeatureCollection",
		Features: features,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(geojson)
}

// handleMapboxDirections proxies requests to Mapbox Directions API
func (s *Server) handleMapboxDirections(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv("MAPBOX_TOKEN")
	if token == "" {
		http.Error(w, "Mapbox token not configured", http.StatusInternalServerError)
		return
	}

	// Get query parameters (URL decoded automatically by Query().Get())
	coordinates := r.URL.Query().Get("coordinates")
	if coordinates == "" {
		http.Error(w, "coordinates parameter required", http.StatusBadRequest)
		return
	}

	// Build Mapbox API URL (coordinates already decoded by Query().Get())
	mapboxURL := fmt.Sprintf("https://api.mapbox.com/directions/v5/mapbox/driving/%s?access_token=%s&geometries=geojson&overview=full",
		coordinates, token)

	// Make request to Mapbox
	resp, err := http.Get(mapboxURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error calling Mapbox API: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// handleMapboxGeocoding proxies requests to Mapbox Geocoding API
func (s *Server) handleMapboxGeocoding(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv("MAPBOX_TOKEN")
	if token == "" {
		http.Error(w, "Mapbox token not configured", http.StatusInternalServerError)
		return
	}

	// Get query parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "q parameter required", http.StatusBadRequest)
		return
	}

	// Build Mapbox API URL
	mapboxURL := fmt.Sprintf("https://api.mapbox.com/geocoding/v5/mapbox.places/%s.json?access_token=%s&country=JP",
		url.QueryEscape(query), token)

	// Make request to Mapbox
	resp, err := http.Get(mapboxURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error calling Mapbox API: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
