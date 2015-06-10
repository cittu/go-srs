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

package protocol

import (
    "math/rand"
    "strings"
    "github.com/cittu/go-srs/core"
    "strconv"
    "net/url"
)

func RandomGenerate (r *rand.Rand, b []byte) {
    for i,_ := range b {
        // the common value in [0x0f, 0xf0]
        b[i] = 0x0f + byte(r.Int31() % (256 - 0x0f - 0x0f))
    }
}

func DiscoveryTcUrl(tcUrl string, logger core.Logger) (schema, host, vhost, app string, port int, param string, err error) {
    rawurl := tcUrl

    // parse the ...vhost... to standard query &vhost=
    for strings.Index(rawurl, "...") >= 0 {
        rawurl = strings.Replace(rawurl, "...", "&", 1)
        rawurl = strings.Replace(rawurl, "...", "=", 1)
    }

    // use url module to parse.
    var uri *url.URL
    if uri,err = url.Parse(tcUrl); err != nil {
        logger.Error("parse tcUrl=%v failed", tcUrl)
        return
    }

    schema = uri.Scheme
    host = uri.Host
    app = strings.Trim(uri.Path, "/")
    param = uri.RawQuery

    port = core.SRS_CONSTS_RTMP_DEFAULT_PORT
    if pos := strings.Index(host, ":"); pos >= 0 {
        if port,err = strconv.Atoi(host[pos + 1:]); err != nil {
            logger.Error("parse port from host=%v failed", host)
            return
        }
        host = host[0:pos]
        logger.Info("discovery host=%v, port=%v", host, port)
    }

    vhost = host
    query := uri.Query()
    if query.Get("vhost") != "" {
        vhost = query.Get("vhost")
    }

    logger.Info("tcUrl parsed to schema=%v, host=%v, port=%v, vhost=%v, app=%v, param=%v",
        schema, host, port, vhost, app, param)

    return
}
