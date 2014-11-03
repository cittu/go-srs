for((i=0;i<5;i++)); do (ab -n 100000 -c 20 http://127.0.0.1:8080/api/v3/json >req100-$i.go.txt &); done
