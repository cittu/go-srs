/**
================================================================================================
go build ./tcp.server.go && ./tcp.server 1 1 1990 4096
g++ tcp.client.cpp -g -O0 -o tcp.client && ./tcp.client 127.0.0.1 1990 4096 

----total-cpu-usage---- -dsk/total- ---net/lo-- ---paging-- ---system--
usr sys idl wai hiq siq| read  writ| recv  send|  in   out | int   csw 
  0   5  93   0   0   2|   0  7509B| 587M  587M|   0     0 |2544   141k
  0   5  93   0   0   2|   0    10k| 524M  524M|   0     0 |2629   123k
 
  PID USER      PR  NI  VIRT  RES  SHR S %CPU %MEM    TIME+  COMMAND                          
 5496 winlin    20   0 98248 1968 1360 S 100.5  0.0   4:40.54 ./tcp.server 1 1 1990 4096      
 5517 winlin    20   0 11740  896  764 S 72.3  0.0   3:24.22 ./tcp.client 127.0.0.1 1990 4096 
 
================================================================================================
go build ./tcp.server.go && ./tcp.server 1 0 1990 4096
g++ tcp.client.cpp -g -O0 -o tcp.client && ./tcp.client 127.0.0.1 1990 4096 

----total-cpu-usage---- -dsk/total- ---net/lo-- ---paging-- ---system--
usr sys idl wai hiq siq| read  writ| recv  send|  in   out | int   csw 
  0   5  93   0   0   1|   0    10k| 868M  868M|   0     0 |2674    79k
  1   5  93   0   0   1|   0    16k| 957M  957M|   0     0 |2660    85k
 
  PID USER      PR  NI  VIRT  RES  SHR S %CPU %MEM    TIME+  COMMAND                         
 3004 winlin    20   0 98248 1968 1360 R 100.2  0.0   2:27.32 ./tcp.server 1 0 1990 4096     
 3030 winlin    20   0 11740  900  764 R 81.0  0.0   1:59.42 ./tcp.client 127.0.0.1 1990 4096
 
================================================================================================
go build ./tcp.server.go && ./tcp.server 10 1 1990 4096
g++ tcp.client.cpp -g -O0 -o tcp.client && for((i=0;i<8;i++)); do (./tcp.client 127.0.0.1 1990 4096 &); done

----total-cpu-usage---- -dsk/total- ---net/lo-- ---paging-- ---system--
usr sys idl wai hiq siq| read  writ| recv  send|  in   out | int   csw 
  4  37  47   0   0  12|   0   105k|3972M 3972M|   0     0 |  14k  995k
  4  37  46   0   0  13|   0  8055B|3761M 3761M|   0     0 |  14k  949k

  PID USER      PR  NI  VIRT  RES  SHR S %CPU %MEM    TIME+  COMMAND                          
 6353 winlin    20   0  517m 6896 1372 R 789.6  0.0  13:24.49 ./tcp.server 10 1 1990 4096     
 6384 winlin    20   0 11740  900  764 S 68.4  0.0   1:11.57 ./tcp.client 127.0.0.1 1990 4096 
 6386 winlin    20   0 11740  896  764 R 67.4  0.0   1:09.53 ./tcp.client 127.0.0.1 1990 4096 
 6390 winlin    20   0 11740  900  764 R 66.7  0.0   1:11.24 ./tcp.client 127.0.0.1 1990 4096 
 6382 winlin    20   0 11740  896  764 R 64.8  0.0   1:11.30 ./tcp.client 127.0.0.1 1990 4096 
 6388 winlin    20   0 11740  896  764 R 64.4  0.0   1:11.80 ./tcp.client 127.0.0.1 1990 4096 
 6380 winlin    20   0 11740  896  764 S 63.4  0.0   1:08.78 ./tcp.client 127.0.0.1 1990 4096 
 6396 winlin    20   0 11740  896  764 R 62.8  0.0   1:09.47 ./tcp.client 127.0.0.1 1990 4096 
 6393 winlin    20   0 11740  900  764 R 61.4  0.0   1:11.90 ./tcp.client 127.0.0.1 1990 4096 
 
================================================================================================
go build ./tcp.server.go && ./tcp.server 10 0 1990 4096
g++ tcp.client.cpp -g -O0 -o tcp.client && for((i=0;i<8;i++)); do (./tcp.client 127.0.0.1 1990 4096 &); done

----total-cpu-usage---- -dsk/total- ---net/lo-- ---paging-- ---system--
usr sys idl wai hiq siq| read  writ| recv  send|  in   out | int   csw 
  5  42  41   0   0  12|   0  8602B|7132M 7132M|   0     0 |  15k  602k
  5  41  41   0   0  12|   0    13k|7426M 7426M|   0     0 |  15k  651k

  PID USER      PR  NI  VIRT  RES  SHR S %CPU %MEM    TIME+  COMMAND                         
 4148 winlin    20   0  528m 9.8m 1376 R 795.5  0.1  81:48.12 ./tcp.server 10 0 1990 4096    
 4167 winlin    20   0 11740  896  764 S 89.8  0.0   8:16.52 ./tcp.client 127.0.0.1 1990 4096
 4161 winlin    20   0 11740  900  764 R 87.8  0.0   8:14.63 ./tcp.client 127.0.0.1 1990 4096
 4174 winlin    20   0 11740  896  764 S 83.2  0.0   8:09.40 ./tcp.client 127.0.0.1 1990 4096
 4163 winlin    20   0 11740  896  764 R 82.6  0.0   8:07.80 ./tcp.client 127.0.0.1 1990 4096
 4171 winlin    20   0 11740  900  764 R 82.2  0.0   8:08.75 ./tcp.client 127.0.0.1 1990 4096
 4169 winlin    20   0 11740  900  764 S 81.9  0.0   8:15.37 ./tcp.client 127.0.0.1 1990 4096
 4165 winlin    20   0 11740  900  764 R 78.9  0.0   8:09.98 ./tcp.client 127.0.0.1 1990 4096
 4177 winlin    20   0 11740  900  764 R 74.0  0.0   8:07.63 ./tcp.client 127.0.0.1 1990 4096
*/
package main
import (
    "fmt"
    "net"
    "os"
    "strconv"
    "runtime"
)

func main() {
    var (
        nb_cpus, no_delay, listen_port, packet_bytes int
        err error
    )
    
    fmt.Println("tcp server to send random data to clients.")
    if len(os.Args) <= 2 {
        fmt.Println("Usage:", os.Args[0], "<cpus> <no_delay> <port> <packet_bytes>")
        fmt.Println("   cpus: how many cpu to use.")
        fmt.Println("   no_delay: whether tcp no delay. go default 1, maybe performance hurt.")
        fmt.Println("   port: the listen port.")
        fmt.Println("   packet_bytes: the bytes for packet to send.")
        fmt.Println("For example:")
        fmt.Println("   ", os.Args[0], 1, 0, 1990, 4096)
        return
    }
    
    if nb_cpus, err = strconv.Atoi(os.Args[1]); err != nil {
        fmt.Println("invalid option cpus", os.Args[1], "and err is", err)
        return
    }
    fmt.Println("nb_cpus is", nb_cpus)
    
    if no_delay, err = strconv.Atoi(os.Args[2]); err != nil {
        fmt.Println("invalid option no_delay", os.Args[2], "and err is", err)
        return
    }
    fmt.Println("no_delay is", no_delay)
    
    if listen_port, err = strconv.Atoi(os.Args[3]); err != nil {
        fmt.Println("invalid option port", os.Args[3], "and err is", err)
        return
    }
    fmt.Println("listen_port is", listen_port)
    
    if packet_bytes, err = strconv.Atoi(os.Args[4]); err != nil {
        fmt.Println("invalid packet_bytes port", os.Args[4], "and err is", err)
        return
    }
    fmt.Println("packet_bytes is", packet_bytes)
    
    runtime.GOMAXPROCS(nb_cpus)
    
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
        
        go handleConnection(conn, no_delay, packet_bytes)
    }
}

func handleConnection(conn *net.TCPConn, no_delay int, packet_bytes int) {
    defer conn.Close()
    fmt.Println("handle connection", conn)
    
    if no_delay == 0 {
        if err := conn.SetNoDelay(false); err != nil {
            fmt.Println("set no delay to false failed.")
            return
        }
        fmt.Println("set no delay to false ok.")
    }
    
    /*SO_SNDBUF := 16384
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
    }
}
