package internal

import (
	"fmt"
	"math"

	"github.com/uber/h3-go/v4"
)

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
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

// GenerateGeoJSON creates H3 hexagons for all cities
func GenerateGeoJSON(resolution int, radiusKm float64) (*GeoJSON, error) {
	features := []Feature{}
	cellSet := make(map[h3.Cell]bool)

	for _, city := range Cities {
		// Get the center cell
		centerLatLng := h3.LatLng{Lat: city.Lat, Lng: city.Lng}
		centerCell, err := h3.LatLngToCell(centerLatLng, resolution)
		if err != nil {
			return nil, fmt.Errorf("error creating center cell for %s: %w", city.Name, err)
		}

		// Get all cells within radius using grid disk
		// Convert radius from km to number of rings (approximate)
		rings := int(radiusKm / 0.5) // Rough approximation for resolution 7

		diskCells, err := centerCell.GridDisk(rings)
		if err != nil {
			return nil, fmt.Errorf("error creating grid disk for %s: %w", city.Name, err)
		}

		// Add each cell to our feature collection
		for _, cell := range diskCells {
			// Skip if we've already added this cell
			if cellSet[cell] {
				continue
			}
			cellSet[cell] = true

			// Get the boundary of the H3 cell
			boundary, err := cell.Boundary()
			if err != nil {
				continue
			}

			// Convert boundary to GeoJSON coordinates
			coordinates := make([][]float64, len(boundary)+1)
			for i, latLng := range boundary {
				coordinates[i] = []float64{latLng.Lng, latLng.Lat}
			}
			// Close the polygon by repeating the first point
			coordinates[len(boundary)] = []float64{boundary[0].Lng, boundary[0].Lat}

			// Get center of cell for properties
			cellCenter, _ := cell.LatLng()

			// Determine which city this cell is closest to
			closestCity := city.Name
			minDist := distance(cellCenter.Lat, cellCenter.Lng, city.Lat, city.Lng)
			for _, otherCity := range Cities {
				dist := distance(cellCenter.Lat, cellCenter.Lng, otherCity.Lat, otherCity.Lng)
				if dist < minDist {
					minDist = dist
					closestCity = otherCity.Name
				}
			}

			feature := Feature{
				Type: "Feature",
				Geometry: Geometry{
					Type:        "Polygon",
					Coordinates: [][][]float64{coordinates},
				},
				Properties: map[string]interface{}{
					"h3_index":   cell.String(),
					"resolution": resolution,
					"city":       closestCity,
				},
			}

			features = append(features, feature)
		}
	}

	return &GeoJSON{
		Type:     "FeatureCollection",
		Features: features,
	}, nil
}

// distance calculates the haversine distance between two points in km
func distance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0 // Earth radius in km

	lat1Rad := lat1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0
	deltaLat := (lat2 - lat1) * math.Pi / 180.0
	deltaLng := (lng2 - lng1) * math.Pi / 180.0

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
