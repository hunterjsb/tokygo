package internal

// api_types.go
//
// Centralized, reusable types for API responses and H3 cell payloads.
// This eliminates ad-hoc map[string]interface{} usage and repeated inline struct definitions.
// Prefer these across server handlers to keep response shapes consistent and type-safe.

// CitiesResponse is returned by /api/cities.
type CitiesResponse struct {
	Cities []City            `json:"cities"`
	Colors map[string]string `json:"colors"`
}

// ConfigResponse is returned by /api/config.
type ConfigResponse struct {
	Resolution int     `json:"resolution"`
	RadiusKm   float64 `json:"radiusKm"`
}

// RoutesResponse is returned by /api/routes.
type RoutesResponse struct {
	Routes []Route `json:"routes"`
}

// H3Boundary represents a closed-loop boundary for an H3 cell.
// Coordinates are [lng, lat] pairs, and the last vertex should repeat the first.
type H3Boundary [][]float64

// H3CellInfo is a canonical representation of an H3 cell with its geometry,
// center point, and neighbor references. Used in grid endpoints.
type H3CellInfo struct {
	Boundary  H3Boundary `json:"boundary"`
	Center    []float64  `json:"center"`    // [lng, lat]
	Neighbors []string   `json:"neighbors"` // neighbor H3 indexes
}

// H3CellResponse is returned by /api/h3/cell.
type H3CellResponse struct {
	H3Index  string     `json:"h3_index"`
	Boundary H3Boundary `json:"boundary"`
}

// H3RingCell is a single ring neighbor cell within /api/h3/ring response.
type H3RingCell struct {
	H3Index  string     `json:"h3_index"`
	Boundary H3Boundary `json:"boundary"`
}

// H3RingResponse is returned by /api/h3/ring.
type H3RingResponse struct {
	Ring []H3RingCell `json:"ring"`
}

// H3GridResponse is returned by /api/h3/grid.
type H3GridResponse struct {
	Cells      map[string]H3CellInfo `json:"cells"`
	Resolution int                   `json:"resolution"`
}

// BBox represents a geographic bounding box.
type BBox struct {
	MinLat float64 `json:"minLat"`
	MinLng float64 `json:"minLng"`
	MaxLat float64 `json:"maxLat"`
	MaxLng float64 `json:"maxLng"`
}

// H3GridWindowResponse is returned by /api/h3/grid_window.
type H3GridWindowResponse struct {
	Cells      map[string]H3CellInfo `json:"cells"`
	Resolution int                   `json:"resolution"`
	BBox       BBox                  `json:"bbox"`
}
