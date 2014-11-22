/**
g++ tcp.client.cpp -g -O0 -o tcp.client && ./tcp.client 1990 4096 
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
#include <netinet/in.h>

#define srs_trace(msg, ...)   printf(msg, ##__VA_ARGS__);printf("\n")

int main(int argc, char** argv)
{
    srs_trace("tcp client to recv bytes from server");
    if (argc <= 2) {
        srs_trace("Usage: %s <port> <packet_bytes>", argv[0]);
        srs_trace("   port: the port to connect to.");
        srs_trace("   packet_bytes: the bytes for packet to send.");
        srs_trace("For example:");
        srs_trace("   %s %d %d", argv[0], 1990, 4096);
        return -1;
    }
    
    int server_port = ::atoi(argv[1]);
    int packet_bytes = ::atoi(argv[2]);
    srs_trace("server_port is %d", server_port);
    srs_trace("packet_bytes is %d", packet_bytes);
    
    int fd = socket(AF_INET, SOCK_STREAM, 0);
    if (fd < 0) {
        srs_trace("create socket failed.");
        return -1;
    }
    
    sockaddr_in addr;
    addr.sin_family = AF_INET;
    addr.sin_port = htons(server_port);
    addr.sin_addr.s_addr = inet_addr("127.0.0.1");
    if (::connect(fd, (sockaddr*)&addr, sizeof(addr)) == -1) {
        srs_trace("connect server error.");
        return -1;
    }
    srs_trace("connect server success. fd=%d", fd);
    
    // get the sockoptions
    int sock_recv_buffer = 0;
    socklen_t nb_sock_recv_buffer = sizeof(int);
    if (getsockopt(fd, SOL_SOCKET, SO_RCVBUF, &sock_recv_buffer, &nb_sock_recv_buffer) < 0) {
        srs_trace("get sockopt failed.");
        return -1;
    }
    srs_trace("SO_RCVBUF=%d", sock_recv_buffer);
    
    char b[4096];
    for (;;) {
        ssize_t nb_recv = recv(fd, b, sizeof(b), 0);
        if (nb_recv <= 0) {
            srs_trace("recv bytes to socket error.");
            ::close(fd);
            break;
        }
    }
    srs_trace("completed");
    
    return 0;
}
