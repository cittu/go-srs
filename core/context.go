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

package core

import "fmt"

// server info.
const (
    RTMP_SIG_SRS_KEY = "GOSRS"
    RTMP_SIG_SRS_ROLE = "origin/edge server"
    RTMP_SIG_SRS_NAME = RTMP_SIG_SRS_KEY + "(Simple RTMP Server)"
    RTMP_SIG_SRS_URL_SHORT = "github.com/winlinvip/go-srs"
    RTMP_SIG_SRS_URL = "https://" + RTMP_SIG_SRS_URL_SHORT
    RTMP_SIG_SRS_WEB = "http://blog.csdn.net/win_lin"
    RTMP_SIG_SRS_EMAIL = "winlin@vip.126.com"
    RTMP_SIG_SRS_LICENSE = "The MIT License (MIT)"
    RTMP_SIG_SRS_COPYRIGHT = "Copyright (c) 2013-2014 winlin"
    RTMP_SIG_SRS_PRIMARY = "winlin"
)

var RTMP_SIG_SRS_VERSION = fmt.Sprintf("%v.%v.%v", Major, Minor, Revision)
var RTMP_SIG_SRS_HANDSHAKE = RTMP_SIG_SRS_KEY + "(" + RTMP_SIG_SRS_VERSION + ")"
func RTMP_SIG_SRS_ISSUES(id int) string {
    return fmt.Sprint("%v/issues/%v", RTMP_SIG_SRS_URL, id)
}
