package main

import (
	"fmt"
	"github.com/go-martini/martini"
)

func main() {
	m := martini.Classic()
	m.Get("/api", func() string {
		return "Hello World!"
	})
	m.Run()
	fmt.Println("Hello, SRS!")
}

