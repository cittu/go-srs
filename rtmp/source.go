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
    Logger core.Logger
    SrsId int
}

func NewRtmpSource(req *protocol.RtmpRequest, logger core.Logger) *RtmpSource {
    return &RtmpSource{
        Req: req,
        Logger: logger,
    }
}

func (source *RtmpSource) Initialize() (err error) {
    return
}

func (source *RtmpSource) OnPublish(logger core.Logger, srsId int) (err error) {
    source.Logger = logger

    // whatever, the publish thread is the source or edge source,
    // save its id to srouce id.
    source.SourceId(srsId)

    // TODO: FIXME: implements it.

    return
}

func (source *RtmpSource) OnUnPublish() {
    // TODO: FIXME: implements it.
}

func (source *RtmpSource) SourceId(srsId int) {
    if source.SrsId == srsId {
        return
    }

    source.SrsId = srsId

    // notice all consumer
    // TODO: FIXME: implements it.
}

func (source *RtmpSource) OnMessage(msg *protocol.RtmpMessage) (err error) {
    // for edge, directly proxy message to origin.
    // TODO: FIXME: implements it.

    // process audio packet
    if msg.Header.IsAudio() {
        if err = source.OnAudio(msg); err != nil {
            source.Logger.Error("source process audio message failed")
            return
        }
    }

    // process video packet
    if msg.Header.IsVideo() {
        if err = source.OnVideo(msg); err != nil {
            source.Logger.Error("source process video message failed")
            return
        }
    }

    // process aggregate packet
    if msg.Header.IsAggregate() {
        // TODO: FIXME: implements it.
    }

    // process onMetaData
    if msg.Header.IsAmf0Data() || msg.Header.IsAmf3Data() {
        // TODO: FIXME: implements it.
    }
    return
}

func (source *RtmpSource) OnAudio(msg *protocol.RtmpMessage) (err error) {
    return
}

func (source *RtmpSource) OnVideo(msg *protocol.RtmpMessage) (err error) {
    return
}

func (source *RtmpSource) GopCache(enabledCache bool) {
    // TODO: FIXME: implements it.
}

var sources = map[string]*RtmpSource{}
func FindSource(req *protocol.RtmpRequest, logger core.Logger) (source *RtmpSource, err error) {
    url := req.StreamUrl()

    if _,ok := sources[url]; !ok {
        source = NewRtmpSource(req, logger)
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
