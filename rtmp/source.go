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

package rtmp

import (
    "github.com/winlinvip/go-srs/protocol"
    "github.com/winlinvip/go-srs/core"
)

type RtmpSource struct {
    Req *protocol.RtmpRequest
    SrsId int
}

func NewRtmpSource(req *protocol.RtmpRequest) *RtmpSource {
    return &RtmpSource{
        Req: req,
    }
}

func (source *RtmpSource) Initialize() (err error) {
    return
}

func (source *RtmpSource) GopCache(enabledCache bool) {
    // TODO: FIXME: implements it.
}

var sources = map[string]*RtmpSource{}
func FindSource(req *protocol.RtmpRequest, logger core.Logger) (source *RtmpSource, err error) {
    url := req.StreamUrl()

    if _,ok := sources[url]; !ok {
        source = NewRtmpSource(req)
        if err = source.Initialize(); err != nil {
            return
        }
        sources[url] = source
        logger.Info("create new source for url=%s, vhost=%s", url, req.Vhost)
    }

    // we always update the request of resource,
    // for origin auth is on, the token in request maybe invalid,
    // and we only need to update the token of request, it's simple.
    source = sources[url]
    source.Req.UpdateAuth(req)

    return
}
