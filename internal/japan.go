package internal

// City represents a Japanese city with its coordinates
type City struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

// Cities contains the main cities for the Japan trip
var Cities = []City{
	{Name: "Tokyo", Lat: 35.6762, Lng: 139.6503},
	{Name: "Kyoto", Lat: 35.0116, Lng: 135.7681},
	{Name: "Osaka", Lat: 34.6937, Lng: 135.5023},
}

// CityColors maps city names to their display colors
var CityColors = map[string]string{
	"Tokyo": "#e74c3c",
	"Kyoto": "#3498db",
	"Osaka": "#2ecc71",
}
