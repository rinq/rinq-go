package main

import "os"

func main() {
	if os.Getenv("OVERPASS_SERVER") != "" {
		runServer()
	} else {
		runClient()
	}
}
