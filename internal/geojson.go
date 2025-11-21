package internal

// GeoJSON represents a GeoJSON FeatureCollection
type GeoJSON struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

// Feature represents a GeoJSON Feature
type Feature struct {
	Type       string                 `json:"type"`
	Geometry   Geometry               `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// Geometry represents a GeoJSON Geometry
type Geometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"` // Can be [][]float64 (LineString) or [][][]float64 (Polygon)
}
