package internal

import (
	_ "embed"
	"encoding/json"
)

//go:embed cached_routes.json
var cachedRoutesJSON []byte

var CachedRoutes []CachedRoute

// Route represents a travel route between two points
type Route struct {
	Name        string     `json:"name"`
	Type        string     `json:"type"` // "train", "walk", "car", "flight"
	Origin      Location   `json:"origin"`
	Destination Location   `json:"destination"`
	Waypoints   []Location `json:"waypoints,omitempty"`
	Distance    float64    `json:"distance"` // in km
	Duration    int        `json:"duration"` // in minutes
}

// CachedRoute represents a route with actual geometry from Mapbox
type CachedRoute struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Origin      Location    `json:"origin"`
	Destination Location    `json:"destination"`
	Geometry    [][]float64 `json:"geometry"` // [lng, lat] pairs
	Distance    float64     `json:"distance"` // in meters
	Duration    float64     `json:"duration"` // in seconds
}

// Location represents a point on the map
type Location struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

// TripRoutes contains all routes for the Japan trip
var TripRoutes = []Route{
	{
		Name: "Tokyo to Kyoto Shinkansen",
		Type: "train",
		Origin: Location{
			Name: "Tokyo Station",
			Lat:  35.6812,
			Lng:  139.7671,
		},
		Destination: Location{
			Name: "Kyoto Station",
			Lat:  34.9851,
			Lng:  135.7584,
		},
		Waypoints: []Location{
			{Name: "Shinagawa", Lat: 35.6284, Lng: 139.7387},
			{Name: "Shin-Yokohama", Lat: 35.5067, Lng: 139.6174},
			{Name: "Nagoya", Lat: 35.1707, Lng: 136.8816},
		},
		Distance: 476.0,
		Duration: 140,
	},
	{
		Name: "Kyoto to Osaka Transfer",
		Type: "car",
		Origin: Location{
			Name: "The Hotel Seiryu Kyoto Kiyomizu",
			Lat:  34.9960,
			Lng:  135.7813,
		},
		Destination: Location{
			Name: "The Osaka Station Hotel",
			Lat:  34.7024,
			Lng:  135.4959,
		},
		Distance: 55.0,
		Duration: 60,
	},
	{
		Name: "Haneda Airport to Hotel",
		Type: "car",
		Origin: Location{
			Name: "Haneda Airport (HND)",
			Lat:  35.5494,
			Lng:  139.7798,
		},
		Destination: Location{
			Name: "HOTEL GROOVE SHINJUKU",
			Lat:  35.6938,
			Lng:  139.7036,
		},
		Distance: 22.0,
		Duration: 35,
	},
	{
		Name: "Osaka Hotel to Itami Airport",
		Type: "car",
		Origin: Location{
			Name: "The Osaka Station Hotel",
			Lat:  34.7024,
			Lng:  135.4959,
		},
		Destination: Location{
			Name: "Osaka Itami Airport (ITM)",
			Lat:  34.7855,
			Lng:  135.4381,
		},
		Distance: 15.0,
		Duration: 25,
	},
}

func init() {
	// Load cached routes on startup
	if err := json.Unmarshal(cachedRoutesJSON, &CachedRoutes); err != nil {
		panic("Failed to load cached routes: " + err.Error())
	}
}
