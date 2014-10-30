/**
go build ./tcp.server.go

1. ./tcp.server 1990 4096 >/dev/null
----total-cpu-usage---- -dsk/total- ---net/lo-- ---paging-- ---system--
usr sys idl wai hiq siq| read  writ| recv  send|  in   out | int   csw 
 24  41  27   0   0   8|   0    12k| 632M  632M|   0     0 |  10k   27k
 22  44  27   0   0   7|   0     0 | 645M  645M|   0     0 |9953    26k
*/
package main
import (
    "fmt"
    "net"
    "os"
    "strconv"
)

func main() {
    var (
        listen_port, packet_bytes int
        err error
    )
    
    fmt.Println("tcp server to send random data to clients.")
    if len(os.Args) <= 2 {
        fmt.Println("Usage:", os.Args[0], "<port> <packet_bytes>")
        fmt.Println("   port: the listen port.")
        fmt.Println("   packet_bytes: the bytes for packet to send.")
        fmt.Println("For example:")
        fmt.Println("   ", os.Args[0], 1990, 4096)
        return
    }
    
    if listen_port, err = strconv.Atoi(os.Args[1]); err != nil {
        fmt.Println("invalid option port", os.Args[1], "and err is", err)
        return
    }
    fmt.Println("listen_port is", listen_port)
    
    if packet_bytes, err = strconv.Atoi(os.Args[2]); err != nil {
        fmt.Println("invalid packet_bytes port", os.Args[2], "and err is", err)
        return
    }
    fmt.Println("packet_bytes is", packet_bytes)
    
    listenEP := fmt.Sprintf(":%d", listen_port)
    ln, err := net.Listen("tcp", listenEP)
    if err != nil {
        fmt.Println("listen err", err)
        return
    }
    defer ln.Close()
    fmt.Println("listen ok at", listenEP)
    
    for {
        conn, err := ln.Accept()
        if err != nil {
            fmt.Println("accept err", err)
            continue
        }
        fmt.Println("got a client", conn)
        go handleConnection(conn, packet_bytes)
    }
}

func handleConnection(conn net.Conn, packet_bytes int) {
    defer conn.Close()
    fmt.Println("handle connection", conn)
    
    b := make([]byte, packet_bytes)
    fmt.Println("write", len(b), "bytes to conn")
    
    for {
        n, err := conn.Write(b)
        if err != nil {
            fmt.Println("write data error, n is", n, "and err is", err)
            break
        }
        fmt.Println("write data ok, n is", n)
    }
}
