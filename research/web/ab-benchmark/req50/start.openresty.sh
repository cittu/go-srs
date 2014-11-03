for((i=0;i<5;i++)); do (ab -n 100000 -c 10 http://127.0.0.1:8080/api/v3/json >req50-$i.openresty.txt &); done
