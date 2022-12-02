package main

import (
	"flag"
	"fmt"
)

var pattern string
var cfg string

func init() {
	flag.StringVar(&pattern, "p", "default", "the selected pattern for simulation")
	flag.StringVar(&cfg, "c", "", "the config filename")
}

func main() {
	fmt.Println("Simulation Start!")

	// gateway := broker.MockGateway{}
	// coordinator := broker.MockCoordinator{}

	// simulation.Simulate(simulation.System{
	// 	Gateway:     &gateway,
	// 	Coordinator: coordinator,
	// }, simulation.InitConf{})

	fmt.Println("Simluation End!")
}
