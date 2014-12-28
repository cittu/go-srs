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

package server

import (
    "os"
    "fmt"
    "log"
    "io"
    "github.com/winlinvip/go-srs/core"
)

var goroutineIdSeed int = 99
func goroutineId() int {
    goroutineIdSeed += 1
    return goroutineIdSeed
}

type Factory struct {
}

func NewLog(out io.Writer, prefix string, flag int) core.Logger {
    v := &Logger{}
    v.Flag = flag
    v.Logger = log.New(out, prefix, flag)
    return v
}

func (f *Factory) CreateContext(name string) core.Context {
    v := &Context{}
    v.Id = goroutineId()
    // TODO: FIXME: apply config file.
    v.Logger = NewLog(os.Stdout, fmt.Sprintf("[%s][%d][%d] ", name, os.Getpid(), v.Id),
                        log.Ldate | log.Ltime | core.Ltrace | core.Lwarn | core.Lerror)
    return v
}