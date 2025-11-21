package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hunterjsb/tokygo/internal"
)

func main() {
	// Configuration
	const resolution int = 7
	const radiusKm float64 = 15.0

	fmt.Println("Generating H3 cells for:")
	for _, city := range internal.Cities {
		fmt.Printf("  - %s (%.4f, %.4f)\n", city.Name, city.Lat, city.Lng)
	}

	// Generate GeoJSON
	geojson, err := internal.GenerateGeoJSON(resolution, radiusKm)
	if err != nil {
		panic(err)
	}

	// Output to file
	file, err := os.Create("japan.geojson")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(geojson); err != nil {
		panic(err)
	}

	fmt.Printf("\n✅ Generated GeoJSON with %d H3 cells\n", len(geojson.Features))
	fmt.Println("📍 Coverage: Tokyo, Kyoto, Osaka")
	fmt.Printf("📏 Resolution: %d\n", resolution)
	fmt.Println("💾 Output saved to japan.geojson")
	fmt.Println("\n🚀 Run the server: go run cmd/server/main.go")
}
