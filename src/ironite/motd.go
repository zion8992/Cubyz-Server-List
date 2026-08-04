package main

import "math/rand/v2"

var motds = []string{
	"All systems nominal.",
	"Have you tried turning it off and on again?",
	"Uptime is a lifestyle.",
	"I have a duck in my router, it always goes NAT NAT NAT.",
}

func randomMOTD() string {
	return motds[rand.IntN(len(motds))]
}
