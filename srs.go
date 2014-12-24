/*
The MIT License (MIT)

Copyright (c) 2013-2014 winlin

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"io"
	"fmt"
	"net/http"
	"encoding/json"
	"runtime"
	"github.com/winlinvip/go-srs/core"
	"github.com/winlinvip/go-srs/rtmp"
)

func main() {
	fmt.Println("The golang for", core.SrsUrl)
	fmt.Println(core.SrsSignature, fmt.Sprintf("%d.%d.%d",
		core.Major, core.Minor, core.Revision), core.Copyright)

	fmt.Println("Use", core.Cpus, "cpus for multiple processes")
	runtime.GOMAXPROCS(core.Cpus)

	fmt.Println("Rtmp listen at", core.ListenRtmp)
	go func(){
		if err := rtmp.ListenAndServe(fmt.Sprintf(":%d", core.ListenRtmp)); err != nil {
			fmt.Println("Serve RTMP failed, err is", err)
			return
		}
	}()

	http.HandleFunc("/api/v3/version", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Server", fmt.Sprintf("SRS/%d.%d.%d",
			core.Major, core.Minor, core.Revision))
		w.Header().Set("Content-Type", "application/json")

		data, err := json.Marshal(map[string]interface {}{
			"code": 0,
			"major": core.Major,
			"minor": core.Minor,
			"revision": core.Revision,
		})
		if err != nil {
			fmt.Println("marshal json failed, err is", err)
			return
		}

		io.WriteString(w, string(data))
	})

	url := fmt.Sprintf("http://127.0.0.1:%d/api/v3/version", core.ListenApi)
	fmt.Println("Api listen at ", core.ListenApi, "and url is", url)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", core.ListenApi), nil); err != nil {
		fmt.Println("Serve HTTP failed, err is", err)
		return
	}
}

