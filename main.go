package main

import (
	"flag"
	"log"
)

func main() {
	log.Println("seafile-browse")

	directory := flag.String("d", "", "The seafile-data directory.")

	flag.Parse()

	if *directory == "" {
		log.Fatalln("The directory (d) flag is required.")
	}
}
