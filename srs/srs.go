package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/winlinvip/srs.go/core"
)

func main() {
	fmt.Println("golang for http://github.com/winlinvip/simple-rtmp-server")
	fmt.Println("SRS(Simple RTMP Server)", fmt.Sprintf("%d.%d.%d", core.Major, core.Minor, core.Revision), "Copyright (c) 2013-2014")

	m := martini.Classic()
	m.Get("/api/v3/version", func() string {
		return "Hello World!"
	})
	m.Run()
	fmt.Println("Hello, SRS!")
}

