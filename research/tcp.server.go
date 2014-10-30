/**
1. set packet to 4096
go build ./tcp.server.go && ./tcp.server 1990 4096 >/dev/null

----total-cpu-usage---- -dsk/total- ---net/lo-- ---paging-- ---system--
usr sys idl wai hiq siq| read  writ| recv  send|  in   out | int   csw 
 22  30  34   0   0  14|   0     0 | 995M  995M|   0     0 |4403    11k
 23  31  35   0   0  11|   0    24k|1005M 1005M|   0     0 |4629    11k
 
  PID USER      PR  NI  VIRT  RES  SHR S %CPU %MEM    TIME+  COMMAND                                                                                                       
 9651 winlin    20   0  115m 8244 1284 R 125.2  0.4   3:35.25 ./tcp.client 1990 4096                                                                                        
 9530 winlin    20   0  115m 8364 1392 S 102.6  0.4   2:57.66 ./tcp.server 1990 4096  
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
    addr, err := net.ResolveTCPAddr("tcp4", listenEP)
    if err != nil {
        fmt.Println("resolve addr err", err)
        return
    }
    ln, err := net.ListenTCP("tcp", addr)
    if err != nil {
        fmt.Println("listen err", err)
        return
    }
    defer ln.Close()
    fmt.Println("listen ok at", listenEP)
    
    for {
        conn, err := ln.AcceptTCP()
        if err != nil {
            fmt.Println("accept err", err)
            continue
        }
        fmt.Println("got a client", conn)
        
        go handleConnection(conn, packet_bytes)
    }
}

func handleConnection(conn *net.TCPConn, packet_bytes int) {
    defer conn.Close()
    fmt.Println("handle connection", conn)
    
    if err := conn.SetNoDelay(false); err != nil {
        fmt.Println("set no delay to false failed.")
        return
    }
    fmt.Println("set no delay to false ok.")
    
    SO_SNDBUF := 16384
    if err := conn.SetWriteBuffer(SO_SNDBUF); err != nil {
        fmt.Println("set send SO_SNDBUF failed.")
        return
    }
    fmt.Println("set send SO_SNDBUF to", SO_SNDBUF, "ok.")
    
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
