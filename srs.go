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
	"github.com/winlinvip/go-srs/server"
)

func main() {
	fmt.Println("The golang for", core.SrsUrl)
	fmt.Println(core.SrsSignature, fmt.Sprintf("%d.%d.%d",
		core.Major, core.Minor, core.Revision), core.Copyright)

	// TODO: FIXME: read and parse the config file.

	// the factory to create objects.
	factory := server.NewFactory()
	logger := factory.CreateLogger("srs")
	logger.Trace("Use %d cpus for multiple processes", core.Cpus)
	runtime.GOMAXPROCS(core.Cpus)

	logger.Trace("Rtmp listen at %v", core.ListenRtmp)
	go func(){
		if err := rtmp.ListenAndServe(fmt.Sprintf(":%d", core.ListenRtmp), factory); err != nil {
			logger.Error("Serve RTMP failed, err is %v", err)
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
			logger.Error("marshal json failed, err is %v", err)
			return
		}

		io.WriteString(w, string(data))
	})

	url := fmt.Sprintf("http://127.0.0.1:%d/api/v3/version", core.ListenApi)
	logger.Trace("Api listen at %v, url is %v", core.ListenApi, url)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", core.ListenApi), nil); err != nil {
		logger.Error("Serve HTTP failed, err is %v", err)
		return
	}
}

