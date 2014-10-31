'''
1. set packet to 4096
python tcp.server.py 1990 4096 >/dev/null

----total-cpu-usage---- -dsk/total- ---net/lo-- ---paging-- ---system--
usr sys idl wai hiq siq| read  writ| recv  send|  in   out | int   csw 
 20   5  63   0   0  12|   0   144k| 245M  245M|   0     0 |2597  2896 
 21   6  63   0   0  10|   0  4096B| 251M  251M|   0     0 |2714  3015 

  PID USER      PR  NI  VIRT  RES  SHR S %CPU %MEM    TIME+  COMMAND                                                                                                       
 5157 winlin    20   0  157m 5780 2808 R 100.0  0.3   0:34.11 python tcp.client.py 1990 4096                                                                                
 5140 winlin    20   0  157m 5932 2824 S 28.2  0.3   0:09.84 python tcp.server.py 1990 4096 
'''
import socket, sys
if len(sys.argv) <= 2:
    print("Usage: %s <port> <packet_bytes>"%(sys.argv[0]))
    print("   port: the listen port.")
    print("   packet_bytes: the bytes for packet to send.")
    print("For example:")
    print("   %s %d %d"%(sys.argv[0], 1990, 4096))
    sys.exit(-1)

listen_port = int(sys.argv[1])
packet_bytes = int(sys.argv[2])
print("listen_port is %d"%listen_port)
print("packet_bytes is %d"%packet_bytes)

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
print("setsockopt reuse-addr success.")

s.bind(('', listen_port))
print("bind socket success.")

s.listen(10)
print("listen socket success.")

b = ''
for i in range(0, packet_bytes):
    b += str(i)
    
while True:
    conn, addr = s.accept()
    while True:
        try:
            conn.send(b)
            print("send %d bytes"%len(b))
        except Exception, ex:
            print("ex:%s"%ex)
            break
    conn.close()
