package internal

import (
	"github.com/uber/h3-go/v4"
)

// LocationType represents the type of location
type LocationType string

const (
	LocationTypeHotel   LocationType = "hotel"
	LocationTypeAirport LocationType = "airport"
	LocationTypeStation LocationType = "station"
)

// TripLocation represents a point of interest in the trip
type TripLocation struct {
	Name string       `json:"name"`
	Type LocationType `json:"type"`
	City string       `json:"city"`
	Lat  float64      `json:"lat"`
	Lng  float64      `json:"lng"`
}

// TripLocations contains all locations for the Japan trip
var TripLocations = []TripLocation{
	// Hotels
	{
		Name: "HOTEL GROOVE SHINJUKU",
		Type: LocationTypeHotel,
		City: "Tokyo",
		Lat:  35.6938,
		Lng:  139.7036,
	},
	{
		Name: "The Hotel Seiryu Kyoto Kiyomizu",
		Type: LocationTypeHotel,
		City: "Kyoto",
		Lat:  34.9960,
		Lng:  135.7813,
	},
	{
		Name: "The Osaka Station Hotel",
		Type: LocationTypeHotel,
		City: "Osaka",
		Lat:  34.7024,
		Lng:  135.4959,
	},
	// Airports
	{
		Name: "Haneda Airport",
		Type: LocationTypeAirport,
		City: "Tokyo",
		Lat:  35.5494,
		Lng:  139.7798,
	},
	{
		Name: "Osaka Itami Airport",
		Type: LocationTypeAirport,
		City: "Osaka",
		Lat:  34.7855,
		Lng:  135.4381,
	},
	// Stations
	{
		Name: "Tokyo Station",
		Type: LocationTypeStation,
		City: "Tokyo",
		Lat:  35.6812,
		Lng:  139.7671,
	},
	{
		Name: "Kyoto Station",
		Type: LocationTypeStation,
		City: "Kyoto",
		Lat:  34.9851,
		Lng:  135.7584,
	},
}

// GetLocationsGeoJSON returns locations as GeoJSON points
func GetLocationsGeoJSON(resolution int) (*GeoJSON, error) {
	features := []Feature{}

	for _, loc := range TripLocations {
		// Convert to H3 cell
		latLng := h3.LatLng{Lat: loc.Lat, Lng: loc.Lng}
		cell, err := h3.LatLngToCell(latLng, resolution)
		if err != nil {
			continue
		}

		// Get center of cell
		center, err := cell.LatLng()
		if err != nil {
			continue
		}

		feature := Feature{
			Type: "Feature",
			Geometry: Geometry{
				Type:        "Point",
				Coordinates: []float64{center.Lng, center.Lat},
			},
			Properties: map[string]any{
				"name":       loc.Name,
				"type":       string(loc.Type),
				"city":       loc.City,
				"h3_index":   cell.String(),
				"resolution": resolution,
			},
		}

		features = append(features, feature)
	}

	return &GeoJSON{
		Type:     "FeatureCollection",
		Features: features,
	}, nil
}
