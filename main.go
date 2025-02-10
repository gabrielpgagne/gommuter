package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"googlemaps.github.io/maps"
)

func main() {
	from := flag.String("from", "", "Starting point")
	to := flag.String("to", "", "Destination")
	key := flag.String("key", "", "Google Maps API Key (Optional)")

	flag.Parse()

	if *from == "" || *to == "" {
		fmt.Println("Missing itinerary parameters")
		flag.Usage()
		return
	}

	if *key == "" {
		value, exists := os.LookupEnv("GOOGLE_MAPS_API_KEY")
		key = &value
		if !exists {
			// Variable is set
			fmt.Println("Using ")
			fmt.Println("Missing Google Maps API Key (env. var GOOGLE_MAPS_API_KEY)")
			flag.Usage()
		}
	}

	c, err := maps.NewClient(maps.WithAPIKey(*key))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	r := &maps.DistanceMatrixRequest{
		Origins:       []string{*from},
		Destinations:  []string{*to},
		DepartureTime: "now",
		// DepartureTime: fmt.Sprint(specificDate.Unix()),
	}
	routes, err := c.DistanceMatrix(context.Background(), r)
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
	fmt.Println(routes.Rows[0].Elements[0].DurationInTraffic.Minutes())
}
