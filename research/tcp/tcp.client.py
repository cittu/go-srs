'''
python tcp.client.py 127.0.0.1 1990 4096
'''
import socket, sys
if len(sys.argv) <= 3:
    print("Usage: %s <ip> <port> <packet_bytes>"%(sys.argv[0]))
    print("   ip: the ip to connect to.")
    print("   port: the port to connect to.")
    print("   packet_bytes: the bytes for packet to send.")
    print("For example:")
    print("   %s %d %d"%(sys.argv[0], "127.0.0.1", 1990, 4096))
    sys.exit(-1)

server_ip = str(sys.argv[1])
listen_port = int(sys.argv[2])
packet_bytes = int(sys.argv[3])
print("server_ip is %s"%server_ip)
print("listen_port is %d"%listen_port)
print("packet_bytes is %d"%packet_bytes)

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

s.connect((server_ip, listen_port))
print("connect socket %s:%d success."%(server_ip, listen_port))
    
while True:
    data = s.recv(1024)
    if not data: 
        break
s.close()
