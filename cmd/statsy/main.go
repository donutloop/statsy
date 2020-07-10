package main

import (
	"github.com/donutloop/statsy/internal/api"
)

func main() {
	a := api.NewAPI(false)
	a.Bootstrap()
	a.Start()
}
