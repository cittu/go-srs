/**

1. set packet to 4096
g++ tcp.server.cpp -g -O0 -o tcp.server && ./tcp.server 1990 4096 >/dev/null 

----total-cpu-usage---- -dsk/total- ---net/lo-- ---paging-- ---system--
usr sys idl wai hiq siq| read  writ| recv  send|  in   out | int   csw 
  6  25  44   0   0  24|   0   224k|2142M 2142M|   0     0 |2439    15k
  5  23  48   0   0  23|   0     0 |2028M 2028M|   0     0 |2803    10k
  
  PID USER      PR  NI  VIRT  RES  SHR S %CPU %MEM    TIME+  COMMAND                                                                                                       
 9758 winlin    20   0 11648  900  764 R 98.3  0.0   0:11.36 ./tcp.client 1990 4096                                                                                         
 9751 winlin    20   0 11648  884  752 R 71.4  0.0   0:07.85 ./tcp.server 1990 4096 
*/
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>

#include <sys/types.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <signal.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>

#define srs_trace(msg, ...)   printf(msg, ##__VA_ARGS__);printf("\n")

int main(int argc, char** argv)
{
    srs_trace("tcp server to send random data to clients.");
    if (argc <= 2) {
        srs_trace("Usage: %s <port> <packet_bytes>", argv[0]);
        srs_trace("   port: the listen port.");
        srs_trace("   packet_bytes: the bytes for packet to send.");
        srs_trace("For example:");
        srs_trace("   %s %d %d", argv[0], 1990, 4096);
        return -1;
    }
    
    int listen_port = ::atoi(argv[1]);
    int packet_bytes = ::atoi(argv[2]);
    srs_trace("listen_port is %d", listen_port);
    srs_trace("packet_bytes is %d", packet_bytes);
    
    int fd = socket(AF_INET, SOCK_STREAM, 0);
    if (fd < 0) {
        srs_trace("create socket failed.");
        return -1;
    }
    
    int reuse_socket = 1;
    if (setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &reuse_socket, sizeof(int)) == -1) {
        srs_trace("setsockopt reuse-addr error.");
        return -1;
    }
    srs_trace("setsockopt reuse-addr success. fd=%d", fd);
    
    sockaddr_in addr;
    addr.sin_family = AF_INET;
    addr.sin_port = htons(listen_port);
    addr.sin_addr.s_addr = INADDR_ANY;
    if (::bind(fd, (const sockaddr*)&addr, sizeof(sockaddr_in)) == -1) {
        srs_trace("bind socket error.");
        return -1;
    }
    srs_trace("bind socket success. fd=%d", fd);
    
    if (::listen(fd, 10) == -1) {
        srs_trace("listen socket error.");
        return -1;
    }
    srs_trace("listen socket success. fd=%d", fd);
    
    // get the sockoptions
    int sock_send_buffer = 0;
    socklen_t nb_sock_send_buffer = sizeof(int);
    if (getsockopt(fd, SOL_SOCKET, SO_SNDBUF, &sock_send_buffer, &nb_sock_send_buffer) < 0) {
        srs_trace("get sockopt failed.");
        return -1;
    }
    srs_trace("SO_SNDBUF=%d", sock_send_buffer);
    
    for (;;) {
        int conn = accept(fd, NULL, NULL);
        if (conn < 0) {
            srs_trace("accept socket error.");
            return -1;
        }
        srs_trace("accept socket ok, conn=%d", conn);
        
        char b[4096];
        for (;;) {
            ssize_t nb_send = send(conn, b, sizeof(b), 0);
            if (nb_send != sizeof(b)) {
                srs_trace("send bytes to socket error.");
                ::close(conn);
                break;
            }
            srs_trace("send bytes ok.");
        }
    }
    
    return 0;
}
