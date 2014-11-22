/**
================================================================================================
g++ tcp.server.writev.cpp -g -O0 -o tcp.server && ./tcp.server 64 1990 4096 
g++ tcp.client.readv.cpp -g -O0 -o tcp.client && ./tcp.client 127.0.0.1 1990 64 4096 

----total-cpu-usage---- -dsk/total- ---net/lo-- ---paging-- ---system--
usr sys idl wai hiq siq| read  writ| recv  send|  in   out | int   csw 
  0   6  93   0   0   1|   0    15k|1742M 1742M|   0     0 |2578    30k
  0   6  93   0   0   1|   0    13k|1779M 1779M|   0     0 |2412    30k
  
  PID USER      PR  NI  VIRT  RES  SHR S %CPU %MEM    TIME+  COMMAND                            
 9468 winlin    20   0 12008 1192  800 R 99.8  0.0   1:17.63 ./tcp.server 64 1990 4096          
 9487 winlin    20   0 12008 1192  800 R 80.3  0.0   1:02.49 ./tcp.client 127.0.0.1 1990 64 4096
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
#include <sys/uio.h>

#define srs_trace(msg, ...)   printf(msg, ##__VA_ARGS__);printf("\n")

int main(int argc, char** argv)
{
    srs_trace("tcp server to send random data to clients.");
    if (argc <= 3) {
        srs_trace("Usage: %s <nb_writev> <port> <packet_bytes>", argv[0]);
        srs_trace("   nb_writev: the number of iovec for writev.");
        srs_trace("   port: the listen port.");
        srs_trace("   packet_bytes: the bytes for packet to send.");
        srs_trace("For example:");
        srs_trace("   %s %d %d", argv[0], 64, 1990, 4096);
        return -1;
    }
    
    int nb_writev = ::atoi(argv[1]);
    int listen_port = ::atoi(argv[2]);
    int packet_bytes = ::atoi(argv[3]);
    srs_trace("nb_writev is %d", nb_writev);
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
    
    iovec* iov = new iovec[nb_writev];
    for (int i = 0; i < nb_writev; i++) {
        iovec& item = iov[i];
        item.iov_base = new char[packet_bytes];
        item.iov_len = packet_bytes;
    }
    
    for (;;) {
        int conn = accept(fd, NULL, NULL);
        if (conn < 0) {
            srs_trace("accept socket error.");
            return -1;
        }
        srs_trace("accept socket ok, conn=%d", conn);
        
        for (;;) {
            ssize_t nb_send = writev(conn, iov, nb_writev);
            if (nb_send != packet_bytes * nb_writev) {
                srs_trace("send bytes to socket error.");
                ::close(conn);
                break;
            }
        }
    }
    
    return 0;
}
