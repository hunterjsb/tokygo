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
	cell, err := h3.LatLngToCell(latLng, res)
	if err != nil {
		panic(err)
	}
	fmt.Println("cell ", cell)

	const h h3.Cell = 0x8a283082a677fff
	latLng, err = h.LatLng()
	if err != nil {
		panic(err)
	}
	fmt.Println("lat-long ", latLng)
}
