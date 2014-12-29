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
    "log"
    "os"
    "fmt"
    "github.com/winlinvip/go-srs/core"
)

type Logger struct {
    GoroutineId int
    Flag int
    Logger *log.Logger
}

func (l *Logger) Info(format string, v ...interface{}) {
    if l.Flag&core.Linfo != 0 {
        l.Logger.Output(2, "[info] "+fmt.Sprintf(format, v...)+"\n")
    }
}

func (l *Logger) Trace(format string, v ...interface{}) {
    if l.Flag&(core.Ltrace) != 0 {
        l.Logger.Output(2, "[trace] "+fmt.Sprintf(format, v...)+"\n")
    }
}

func (l *Logger) Warn(format string, v ...interface{}) {
    if l.Flag&core.Ltrace != 0 {
        l.Logger.Output(2, "[warn] "+fmt.Sprintf(format, v...)+"\n")
    }
}

func (l *Logger) Error(format string, v ...interface{}) {
    if l.Flag&core.Ltrace != 0 {
        l.Logger.Output(2, "[error] "+fmt.Sprintf(format, v...)+"\n")
    }
}

func (l *Logger) Print(v ...interface{}) {
    if l.Flag&core.Ltrace != 0 {
        l.Logger.Output(2, "[trace] "+fmt.Sprint(v...))
    }
}

func (l *Logger) Printf(format string, v ...interface{}) {
    if l.Flag&core.Ltrace != 0 {
        l.Logger.Output(2, "[trace] "+fmt.Sprintf(format, v...))
    }
}

func (l *Logger) Println(v ...interface{}) {
    if l.Flag&core.Ltrace != 0 {
        l.Logger.Output(2, "[trace] "+fmt.Sprintln(v...))
    }
}

func (l *Logger) Fatal(v ...interface{}) {
    if l.Flag&core.Lerror != 0 {
        l.Logger.Output(2, "[error] " + fmt.Sprint(v...))
    }
    os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
    if l.Flag&core.Lerror != 0 {
        l.Logger.Output(2, "[error] "+fmt.Sprintf(format, v...))
    }
    os.Exit(1)
}

func (l *Logger) Fatalln(v ...interface{}) {
    if l.Flag&core.Lerror != 0 {
        l.Logger.Output(2, "[error] "+fmt.Sprintln(v...))
    }
    os.Exit(1)
}

func (l *Logger) Panic(v ...interface{}) {
    s := "[error] " + fmt.Sprint(v...)
    if l.Flag&core.Lerror != 0 {
        l.Logger.Output(2, s)
    }
    panic(s)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
    s := "[error] "+fmt.Sprintf(format, v...)
    if l.Flag&core.Lerror != 0 {
        l.Logger.Output(2, s)
    }
    panic(s)
}

func (l *Logger) Panicln(v ...interface{}) {
    s := "[error] "+fmt.Sprintln(v...)
    if l.Flag&core.Lerror != 0 {
        l.Logger.Output(2, s)
    }
    panic(s)
}
