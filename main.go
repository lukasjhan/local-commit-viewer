package main

import (
	"flag"
)

func main() {
	var folder string
	flag.StringVar(&folder, "add", "", "add a new folder to scan for Git repositories")
	flag.Parse()

	if folder != "" {
		scan(folder)
		return
	}

	stats()
}
