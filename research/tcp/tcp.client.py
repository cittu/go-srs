'''
python tcp.client.py 1990 4096
'''
import socket, sys
if len(sys.argv) <= 2:
    print("Usage: %s <port> <packet_bytes>"%(sys.argv[0]))
    print("   port: the port to connect to.")
    print("   packet_bytes: the bytes for packet to send.")
    print("For example:")
    print("   %s %d %d"%(sys.argv[0], 1990, 4096))
    sys.exit(-1)

listen_port = int(sys.argv[1])
packet_bytes = int(sys.argv[2])
print("listen_port is %d"%listen_port)
print("packet_bytes is %d"%packet_bytes)

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

s.connect(('', listen_port))
print("connect socket success.")
    
while True:
    data = s.recv(1024)
    if not data: 
        break
s.close()
