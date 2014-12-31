go-srs
======

golang for https://github.com/winlinvip/simple-rtmp-server

## Usage

To clone from github, build, install to $GOPATH and start srs:

```
go get github.com/winlinvip/go-srs && $GOPATH/bin/go-srs
```

Or, for windows:

```
go get github.com/winlinvip/go-srs && %GOPATH%\bin\go-srs.exe
```


About how to set $GOPATH, read [prepare go](http://blog.csdn.net/win_lin/article/details/40618671).

## IDE

Go: http://www.golangtc.com/download

JetBrains IntelliJ IDEA: http://www.jetbrains.com/idea/download

Idea Plugin: https://github.com/go-lang-plugin-org/go-lang-idea-plugin

## Performance

Performance benchmark history.

### Play benchmark

The play benchmark by [st-load](https://github.com/winlinvip/st-load):

<table>
    <tr>
        <th>Update</th>
        <th>GO-SRS</th>
        <th>Clients</th>
        <th>Type</th>
        <th>CPU</th>
        <th>Memory</th>
        <th>Commit</th>
    </tr>
    <tr>
        <td>2014-12-31</td>
        <td>3.0.1</td>
        <td>10k(10000)</td>
        <td>players</td>
        <td>547.7%</td>
        <td>1.3GB</td>
        <td><a href="https://github.com/winlinvip/simple-rtmp-server/commit/d28da3e0ee90c353afc5641c49f251f8cce02701">commit</a></td>
    </tr>
</table>

## Features

* VP6 codec stream.
* FMLE/FFMPEG/Flash publish.
* Flash/VLC/st-load play.

Winlin 2014.11
