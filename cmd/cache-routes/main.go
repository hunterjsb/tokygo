package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/hunterjsb/tokygo/internal"
)

type MapboxResponse struct {
	Routes []struct {
		Geometry struct {
			Type        string      `json:"type"`
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
	} `json:"routes"`
}

func main() {
	token := os.Getenv("MAPBOX_TOKEN")
	if token == "" {
		fmt.Println("Error: MAPBOX_TOKEN not set")
		os.Exit(1)
	}

	fmt.Println("🗺️  Fetching real routes from Mapbox...")

	cachedRoutes := make([]internal.CachedRoute, 0)

	for _, route := range internal.TripRoutes {
		fmt.Printf("Fetching: %s (%s)\n", route.Name, route.Type)

		// Build waypoints list
		waypoints := []internal.Location{route.Origin}
		waypoints = append(waypoints, route.Waypoints...)
		waypoints = append(waypoints, route.Destination)

		// Build coordinate string
		coordString := ""
		for i, wp := range waypoints {
			if i > 0 {
				coordString += ";"
			}
			coordString += fmt.Sprintf("%f,%f", wp.Lng, wp.Lat)
		}

		// Determine Mapbox profile
		profile := "mapbox/driving"
		switch route.Type {
		case "train":
			profile = "mapbox/driving" // Use driving as approximation for rail
		case "walk":
			profile = "mapbox/walking"
		case "car":
			profile = "mapbox/driving"
		}

		// Fetch from Mapbox
		url := fmt.Sprintf("https://api.mapbox.com/directions/v5/%s/%s?access_token=%s&geometries=geojson&overview=full",
			profile, coordString, token)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("  ❌ Error reading response: %v\n", err)
			continue
		}

		var mapboxResp MapboxResponse
		if err := json.Unmarshal(body, &mapboxResp); err != nil {
			fmt.Printf("  ❌ Error parsing response: %v\n", err)
			continue
		}

		if len(mapboxResp.Routes) == 0 {
			fmt.Printf("  ❌ No routes found\n")
			continue
		}

		mbRoute := mapboxResp.Routes[0]
		fmt.Printf("  ✅ Got route: %.1f km, %.0f min\n",
			mbRoute.Distance/1000, mbRoute.Duration/60)

		// Create cached route
		cached := internal.CachedRoute{
			Name:        route.Name,
			Type:        route.Type,
			Origin:      route.Origin,
			Destination: route.Destination,
			Geometry:    mbRoute.Geometry.Coordinates,
			Distance:    mbRoute.Distance,
			Duration:    mbRoute.Duration,
		}

		cachedRoutes = append(cachedRoutes, cached)
	}

	// Save to JSON file
	output, err := json.MarshalIndent(cachedRoutes, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("internal/cached_routes.json", output, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Cached %d routes to internal/cached_routes.json\n", len(cachedRoutes))
}
