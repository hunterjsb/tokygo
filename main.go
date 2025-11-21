package main

import (
	"fmt"

	"github.com/uber/h3-go/v4"
)

func main() {
	const lat float64 = 37.7955
	const long float64 = -122.3937
	latLng := h3.LatLng{
		Lat: lat,
		Lng: long,
	}

	const res int = 10

	cell, _ := h3.LatLngToCell(latLng, res)
	fmt.Println(cell)
}
