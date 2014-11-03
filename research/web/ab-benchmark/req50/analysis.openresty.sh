vs=`ls req50-*.openresty.txt|xargs cat|grep "Requests"|awk '{print $4}'|awk -F '.' '{print $1}'`; sum=0; for v in $vs; do let sum=$sum+$v; done; echo $sum
