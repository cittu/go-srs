/**
go build ./tcp.client.go && ./tcp.client 1990 4096 >/dev/null
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
        server_port, packet_bytes int
        err error
    )
    fmt.Println("tcp client to recv bytes from server")
    if len(os.Args) <= 2 {
        fmt.Println("Usage:", os.Args[0], "<port> <packet_bytes>")
        fmt.Println("   port: the port to connect to.")
        fmt.Println("   packet_bytes: the bytes for packet to send.")
        fmt.Println("For example:")
        fmt.Println("   ", os.Args[0], 1990, 4096)
        return
    }
    
    if server_port, err = strconv.Atoi(os.Args[1]); err != nil {
        fmt.Println("invalid option port", os.Args[1], "and err is", err)
        return
    }
    fmt.Println("server_port is", server_port)
    
    if packet_bytes, err = strconv.Atoi(os.Args[2]); err != nil {
        fmt.Println("invalid packet_bytes port", os.Args[2], "and err is", err)
        return
    }
    fmt.Println("packet_bytes is", packet_bytes)
    
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
    
    /*if err := conn.SetNoDelay(false); err != nil {
        fmt.Println("set no delay to false failed.")
        return
    }
    fmt.Println("set no delay to false ok.")
    
    SO_RCVBUF := 87380
    if err := conn.SetReadBuffer(SO_RCVBUF); err != nil {
        fmt.Println("set send SO_RCVBUF failed.")
        return
    }
    fmt.Println("set send SO_RCVBUF to", SO_RCVBUF, "ok.")*/
    
    b := make([]byte, packet_bytes)
    for {
        n, err := conn.Read(b)
        if err != nil {
            fmt.Println("read failed, err is", err)
            break
        }
        fmt.Println("read bytes, size is", n)
    }
}