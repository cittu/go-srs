/**
go build ./tcp.client.go && ./tcp.client 1 0 1990 4096
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
        nb_cpus, no_delay, server_port, packet_bytes int
        err error
    )
    fmt.Println("tcp client to recv bytes from server")
    if len(os.Args) <= 2 {
        fmt.Println("Usage:", os.Args[0], "<cpus> <no_delay> <port> <packet_bytes>")
        fmt.Println("   cpus: how many cpu to use.")
        fmt.Println("   no_delay: whether tcp no delay. go default 1, maybe performance hurt.")
        fmt.Println("   port: the port to connect to.")
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
    
    if server_port, err = strconv.Atoi(os.Args[3]); err != nil {
        fmt.Println("invalid option port", os.Args[3], "and err is", err)
        return
    }
    fmt.Println("server_port is", server_port)
    
    if packet_bytes, err = strconv.Atoi(os.Args[4]); err != nil {
        fmt.Println("invalid packet_bytes port", os.Args[4], "and err is", err)
        return
    }
    fmt.Println("packet_bytes is", packet_bytes)
    
    runtime.GOMAXPROCS(nb_cpus)
    
    serverEP := fmt.Sprintf(":%d", server_port)
    addr, err := net.ResolveTCPAddr("tcp4", serverEP)
    if err != nil {
        fmt.Println("resolve addr failed, err is", err)
        return
    }
    conn, err := net.DialTCP("tcp", nil, addr)
    if err != nil {
        fmt.Println("connect server failed, err is", err)
        return
    }
    defer conn.Close()
    fmt.Println("connected at", serverEP)
    
    if no_delay == 0 {
        if err := conn.SetNoDelay(false); err != nil {
            fmt.Println("set no delay to false failed.")
            return
        }
        fmt.Println("set no delay to false ok.")
    }
    
    /*SO_RCVBUF := 87380
    if err := conn.SetReadBuffer(SO_RCVBUF); err != nil {
        fmt.Println("set send SO_RCVBUF failed.")
        return
    }
    fmt.Println("set send SO_RCVBUF to", SO_RCVBUF, "ok.")*/
    
    b := make([]byte, packet_bytes)
    for {
        n, err := conn.Read(b)
        if err != nil {
            fmt.Println("read failed, err is", err, "and n is", n)
            break
        }
    }
}