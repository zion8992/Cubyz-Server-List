package main

import (
	"fmt"
	"github.com/zion8992/netgliss"
	"os"
)

func main() {
	list := netgliss.New()

	err := list.LoadServers("servers-*.json")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	var results []netgliss.Server
	results, err = list.SearchServers("*pvp*", "gamemodes")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	if len(results) == 0 {
		fmt.Printf("No results")
		os.Exit(0)
	}
	for _, v := range results {
		fmt.Printf("%s\n", v.Name)
		fmt.Printf("%s\n", v.Gamemodes)
		fmt.Printf("%s\n", v.Version)
		fmt.Printf("---\n")
	}
}
