package main

import (
	"flag"
	"log"
)

func main() {
	log.Println("seafile-browse")

	directory := flag.String("d", "", "The seafile-data directory.")

	if *directory == "" {
		log.Fatalln("The directory (d) flag is required.")
	}
}
