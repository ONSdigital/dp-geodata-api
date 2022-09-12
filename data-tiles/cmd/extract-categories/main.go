package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ONSdigital/dp-geodata-api/data-tiles/content"
)

func main() {
	c, err := content.Load(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	for _, cat := range c.Categories() {
		fmt.Println(cat)
	}
}
