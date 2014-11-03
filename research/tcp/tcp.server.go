/**
================================================================================================
1. VirtualBox, Thinkpad, T430, 2CPU, 4096B/packet, S:GO, C:GO
go build ./tcp.server.go && ./tcp.server 1990 4096 >/dev/null
go build ./tcp.client.go && ./tcp.client 1990 4096 >/dev/null

----total-cpu-usage---- -dsk/total- ---net/lo-- ---paging-- ---system--
usr sys idl wai hiq siq| read  writ| recv  send|  in   out | int   csw 
 17  34  30   0   0  19|   0     0 | 657M  657M|   0     0 |6183    22k
 16  32  31   0   0  20|   0     0 | 655M  655M|   0     0 |6205    22k
 
  PID USER      PR  NI  VIRT  RES  SHR S %CPU %MEM    TIME+  COMMAND                                                                                                     
 5467 winlin    20   0  115m 2156 1276 S 129.1  0.1   0:20.80 ./tcp.client 1990 4096                                                                                        
 5415 winlin    20   0  180m 2356 1404 R 100.8  0.1   1:36.31 ./tcp.server 1990 4096 
 
================================================================================================
2. VirtualBox, Thinkpad, T430, 2CPU, 4096B/packet, S:GO, C:C++
go build ./tcp.server.go && ./tcp.server 1990 4096 >/dev/null
g++ tcp.client.cpp -g -O0 -o tcp.client && ./tcp.client 1990 4096 >/dev/null 

----total-cpu-usage---- -dsk/total- ---net/lo-- ---paging-- ---system--
usr sys idl wai hiq siq| read  writ| recv  send|  in   out | int   csw 
  7  17  51   0   0  25|   0     0 | 680M  680M|   0     0 |2207    48k
  7  15  52   0   0  26|   0     0 | 680M  680M|   0     0 |2228    48k
  
  PID USER      PR  NI  VIRT  RES  SHR S %CPU %MEM    TIME+  COMMAND             
  PID USER      PR  NI  VIRT  RES  SHR S %CPU %MEM    TIME+  COMMAND                                                                                                       
 5415 winlin    20   0  169m 2220 1404 R 100.4  0.1   0:27.56 ./tcp.server 1990 4096                                                                                        
 5424 winlin    20   0 11648  900  764 R 85.1  0.0   0:23.47 ./tcp.client 1990 4096   
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
    
    /*if err := conn.SetNoDelay(false); err != nil {
        fmt.Println("set no delay to false failed.")
        return
    }
    fmt.Println("set no delay to false ok.")
    
    SO_SNDBUF := 16384
    if err := conn.SetWriteBuffer(SO_SNDBUF); err != nil {
        fmt.Println("set send SO_SNDBUF failed.")
        return
    }
    fmt.Println("set send SO_SNDBUF to", SO_SNDBUF, "ok.")*/
    
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
