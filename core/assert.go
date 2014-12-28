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

import (
    "fmt"
)

type assertError struct {
    a interface {}
    b interface {}
    op string
}

func newAssertError(a interface {}, b interface {}, op string) *assertError {
    return &assertError{
        a: a,
        b: b,
        op: op,
    }
}

func (ae *assertError) Error() string {
    return fmt.Sprintf("assert (%v %s %v) failed", ae.a, ae.op, ae.b)
}

func AssertNil(a interface {}) {
    AssertEquals(a, nil)
}

func AssertNotNil(a interface {}) {
    AssertNotEquals(a, nil)
}

func AssertEquals(a interface {}, b interface {}) {
    if a == b {
        return
    }
    panic(newAssertError(a, b, "=="))
}

func AssertNotEquals(a interface {}, b interface {}) {
    if a != b {
        return
    }
    panic(newAssertError(a, b, "!="))
}

func AssertGreaterThan(a int64, b int64) {
    if a > b {
        return
    }
    panic(newAssertError(a, b, ">"))
}

func AssertSmallerThan(a int64, b int64) {
    if a < b {
        return
    }
    panic(newAssertError(a, b, "<"))
}

func AssertSmallerOrEquals(a int64, b int64) {
    if a <= b {
        return
    }
    panic(newAssertError(a, b, "<="))
}

func AssertGreaterOrEquals(a int64, b int64) {
    if a >= b {
        return
    }
    panic(newAssertError(a, b, ">="))
}

