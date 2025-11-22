package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/uber/h3-go/v4"
)

// json helpers for consistent JSON responses across handlers (generic)
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func okJSON(w http.ResponseWriter, v any) {
	writeJSON(w, http.StatusOK, v)
}

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
	// API endpoints with CORS support
	http.HandleFunc("/api/cities", corsMiddleware(s.handleCities))
	http.HandleFunc("/api/config", corsMiddleware(s.handleConfig))
	http.HandleFunc("/api/routes", corsMiddleware(s.handleRoutes))

	http.HandleFunc("/api/routes/lines", corsMiddleware(s.handleRoutesLines))
	http.HandleFunc("/api/locations", corsMiddleware(s.handleLocations))
	http.HandleFunc("/api/h3/cell", corsMiddleware(s.handleH3Cell))
	http.HandleFunc("/api/h3/ring", corsMiddleware(s.handleH3Ring))
	http.HandleFunc("/api/h3/grid", corsMiddleware(s.handleH3Grid))
	http.HandleFunc("/api/h3/grid_window", corsMiddleware(s.handleH3GridWindow))
	http.HandleFunc("/api/mapbox/directions", corsMiddleware(s.handleMapboxDirections))
	http.HandleFunc("/api/mapbox/geocoding", corsMiddleware(s.handleMapboxGeocoding))

	// Custom handler for static files that doesn't catch /api routes
	frontendDir := filepath.Join(s.RootDir, "frontend", "dist")
	fs := http.FileServer(http.Dir(frontendDir))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}

// handleCities returns the cities data as JSON
func (s *Server) handleCities(w http.ResponseWriter, r *http.Request) {
	response := CitiesResponse{
		Cities: Cities,
		Colors: CityColors,
	}

	okJSON(w, response)
}

// handleConfig returns the server configuration
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	response := ConfigResponse{
		Resolution: s.Config.Resolution,
		RadiusKm:   s.Config.RadiusKm,
	}

	okJSON(w, response)
}

// handleRoutes returns the list of routes
func (s *Server) handleRoutes(w http.ResponseWriter, r *http.Request) {
	response := RoutesResponse{
		Routes: TripRoutes,
	}

	okJSON(w, response)
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
			Properties: map[string]any{
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

	okJSON(w, geojson)
}

// handleLocations returns trip locations as GeoJSON points
func (s *Server) handleLocations(w http.ResponseWriter, r *http.Request) {
	geojson, err := GetLocationsGeoJSON(s.Config.Resolution)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating locations GeoJSON: %v", err), http.StatusInternalServerError)
		return
	}

	okJSON(w, geojson)
}

// handleH3Cell returns H3 cell and boundary for a given lat/lng
func (s *Server) handleH3Cell(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	resStr := r.URL.Query().Get("resolution")

	if latStr == "" || lngStr == "" {
		http.Error(w, "lat and lng parameters required", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "invalid lat", http.StatusBadRequest)
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		http.Error(w, "invalid lng", http.StatusBadRequest)
		return
	}

	resolution := 9
	if resStr != "" {
		res, err := strconv.Atoi(resStr)
		if err == nil {
			resolution = res
		}
	}

	latLng := h3.LatLng{Lat: lat, Lng: lng}
	cell, err := h3.LatLngToCell(latLng, resolution)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting H3 cell: %v", err), http.StatusInternalServerError)
		return
	}

	boundary, err := cell.Boundary()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting boundary: %v", err), http.StatusInternalServerError)
		return
	}

	coords := make([][]float64, len(boundary)+1)
	for i, latLng := range boundary {
		coords[i] = []float64{latLng.Lng, latLng.Lat}
	}
	coords[len(boundary)] = []float64{boundary[0].Lng, boundary[0].Lat}

	response := H3CellResponse{
		H3Index:  cell.String(),
		Boundary: coords,
	}

	okJSON(w, response)
}

// handleH3Ring returns the ring of 6 neighboring cells
func (s *Server) handleH3Ring(w http.ResponseWriter, r *http.Request) {
	h3IndexStr := r.URL.Query().Get("h3_index")
	if h3IndexStr == "" {
		http.Error(w, "h3_index parameter required", http.StatusBadRequest)
		return
	}

	cell := h3.Cell(0)
	if err := cell.UnmarshalText([]byte(h3IndexStr)); err != nil {
		http.Error(w, "invalid h3_index", http.StatusBadRequest)
		return
	}

	// Get the ring of neighbors (k=1 means immediate neighbors)
	ring, err := cell.GridDisk(1)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting ring: %v", err), http.StatusInternalServerError)
		return
	}

	ringData := []H3RingCell{}
	for _, ringCell := range ring {
		if ringCell == cell {
			continue // Skip the center cell
		}

		boundary, err := ringCell.Boundary()
		if err != nil {
			continue
		}

		coords := make([][]float64, len(boundary)+1)
		for i, latLng := range boundary {
			coords[i] = []float64{latLng.Lng, latLng.Lat}
		}
		coords[len(boundary)] = []float64{boundary[0].Lng, boundary[0].Lat}

		ringData = append(ringData, H3RingCell{
			H3Index:  ringCell.String(),
			Boundary: coords,
		})
	}

	response := H3RingResponse{
		Ring: ringData,
	}

	okJSON(w, response)
}

// handleH3Grid returns a precomputed grid of H3 cells for the Japan region
func (s *Server) handleH3Grid(w http.ResponseWriter, r *http.Request) {
	resStr := r.URL.Query().Get("resolution")
	resolution := 7
	if resStr != "" {
		res, err := strconv.Atoi(resStr)
		if err == nil {
			resolution = res
		}
	}

	// Define bounding box for Japan region
	minLat, maxLat := 30.0, 46.0
	minLng, maxLng := 128.0, 146.0

	cells := make(map[string]H3CellInfo)

	// Generate grid
	for lat := minLat; lat <= maxLat; lat += 0.3 {
		for lng := minLng; lng <= maxLng; lng += 0.3 {
			latLng := h3.LatLng{Lat: lat, Lng: lng}
			cell, err := h3.LatLngToCell(latLng, resolution)
			if err != nil {
				continue
			}

			h3Index := cell.String()
			if _, exists := cells[h3Index]; exists {
				continue
			}

			boundary, err := cell.Boundary()
			if err != nil {
				continue
			}

			coords := make([][]float64, len(boundary)+1)
			for i, ll := range boundary {
				coords[i] = []float64{ll.Lng, ll.Lat}
			}
			coords[len(boundary)] = []float64{boundary[0].Lng, boundary[0].Lat}

			center, _ := cell.LatLng()

			// Get neighbors
			neighbors, _ := cell.GridDisk(1)
			neighborIndices := []string{}
			for _, n := range neighbors {
				if n != cell {
					neighborIndices = append(neighborIndices, n.String())
				}
			}

			cells[h3Index] = H3CellInfo{
				Boundary:  coords,
				Center:    []float64{center.Lng, center.Lat},
				Neighbors: neighborIndices,
			}
		}
	}

	response := H3GridResponse{
		Cells:      cells,
		Resolution: resolution,
	}

	okJSON(w, response)
}

// handleH3GridWindow returns H3 cells within a provided bounding box at a given resolution
// Query params:
// - minLat, minLng, maxLat, maxLng: bounding box (required)
// - resolution: H3 resolution (optional, default 7)
func (s *Server) handleH3GridWindow(w http.ResponseWriter, r *http.Request) {
	minLatStr := r.URL.Query().Get("minLat")
	minLngStr := r.URL.Query().Get("minLng")
	maxLatStr := r.URL.Query().Get("maxLat")
	maxLngStr := r.URL.Query().Get("maxLng")
	resStr := r.URL.Query().Get("resolution")

	if minLatStr == "" || minLngStr == "" || maxLatStr == "" || maxLngStr == "" {
		http.Error(w, "minLat, minLng, maxLat, and maxLng parameters are required", http.StatusBadRequest)
		return
	}

	minLat, err := strconv.ParseFloat(minLatStr, 64)
	if err != nil {
		http.Error(w, "invalid minLat", http.StatusBadRequest)
		return
	}
	minLng, err := strconv.ParseFloat(minLngStr, 64)
	if err != nil {
		http.Error(w, "invalid minLng", http.StatusBadRequest)
		return
	}
	maxLat, err := strconv.ParseFloat(maxLatStr, 64)
	if err != nil {
		http.Error(w, "invalid maxLat", http.StatusBadRequest)
		return
	}
	maxLng, err := strconv.ParseFloat(maxLngStr, 64)
	if err != nil {
		http.Error(w, "invalid maxLng", http.StatusBadRequest)
		return
	}

	if minLat > maxLat || minLng > maxLng {
		http.Error(w, "min values must be <= max values", http.StatusBadRequest)
		return
	}

	resolution := 7
	if resStr != "" {
		if res, err := strconv.Atoi(resStr); err == nil {
			resolution = res
		}
	}

	cells := make(map[string]H3CellInfo)

	// Build a polygon for the current viewport bbox and densely cover it with H3 cells
	loop := h3.GeoLoop{
		{Lat: minLat, Lng: minLng},
		{Lat: minLat, Lng: maxLng},
		{Lat: maxLat, Lng: maxLng},
		{Lat: maxLat, Lng: minLng},
	}
	polygon := h3.GeoPolygon{GeoLoop: loop}

	polyCells, err := h3.PolygonToCells(polygon, resolution)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating H3 cells for bbox: %v", err), http.StatusInternalServerError)
		return
	}

	for _, cell := range polyCells {
		h3Index := cell.String()
		if _, exists := cells[h3Index]; exists {
			continue
		}

		boundary, err := cell.Boundary()
		if err != nil {
			continue
		}

		coords := make([][]float64, len(boundary)+1)
		for i, ll := range boundary {
			coords[i] = []float64{ll.Lng, ll.Lat}
		}
		coords[len(boundary)] = []float64{boundary[0].Lng, boundary[0].Lat}

		center, _ := cell.LatLng()

		// Neighbors (optional; helpful for UI effects)
		neighbors, _ := cell.GridDisk(1)
		neighborIndices := []string{}
		for _, n := range neighbors {
			if n != cell {
				neighborIndices = append(neighborIndices, n.String())
			}
		}

		cells[h3Index] = H3CellInfo{
			Boundary:  coords,
			Center:    []float64{center.Lng, center.Lat},
			Neighbors: neighborIndices,
		}
	}

	response := H3GridWindowResponse{
		Cells:      cells,
		Resolution: resolution,
		BBox: BBox{
			MinLat: minLat,
			MinLng: minLng,
			MaxLat: maxLat,
			MaxLng: maxLng,
		},
	}

	okJSON(w, response)
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
