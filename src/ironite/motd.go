package main

import "math/rand/v2"

var motds = []string{
	"All systems nominal.",
	"Have you tried turning it off and on again?",
	"Uptime is a lifestyle.",
	"It's always DNS.",
	"Backups you haven't restored are just hopes.",
}

func randomMOTD() string {
	return motds[rand.IntN(len(motds))]
}
