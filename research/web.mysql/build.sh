#!/bin/bash

WEB_ROOT=http://winlinvip.github.io/srs.release

# calc the dir
echo "argv[0]=$0"
if [[ ! -f $0 ]]; then 
    echo "directly execute the scripts on shell.";
    work_dir=`pwd`
else 
    echo "execute scripts in file: $0";
    work_dir=`dirname $0`; work_dir=`(cd ${work_dir} && pwd)`
fi &&

workdir=$work_dir &&
objs=${workdir}/objs &&
release=$objs/_release &&
mkdir -p $objs &&
echo "objs: $objs" &&
echo "release: $release" &&

if [[ ! -f $release/nginx/sbin/nginx ]]; then
    echo "build ngx_openresty-1.7.4.1" &&
    cd $objs &&
    rm -rf ngx_openresty-1.7.4.1 &&
    wget $WEB_ROOT/3rdparty/ngx_openresty-1.7.4.1.tar.gz -O ngx_openresty-1.7.4.1.tar.gz &&
    tar xf ngx_openresty-1.7.4.1.tar.gz && 
    cd ngx_openresty-1.7.4.1 &&
    ./configure --prefix=$release \
        --with-luajit --with-http_stub_status_module --without-http_redis2_module \
        --with-http_iconv_module --with-http_mp4_module --with-http_flv_module --with-http_realip_module &&
    echo "dynamic link lua" &&
    make && make install &&
    echo "static link lua" &&
    rm -f build/nginx-1.7.4/objs/nginx &&
    sed -i "s|-lluajit-5.1|$release/luajit/lib/libluajit-5.1.a|g" build/nginx-1.7.4/objs/Makefile &&
    make && make install &&
    echo "ngx_openresty-1.7.4.1 ok"
else
    echo "ngx_openresty-1.7.4.1 ok"
fi &&

# for nginx-lua mysql
# https://github.com/openresty/lua-resty-mysql
if [[ ! -f $workdir/api/v3/mysql.lua ]]; then
    cd $objs &&
    rm -rf lua-resty-mysql &&
    git clone https://github.com/openresty/lua-resty-mysql.git &&
    cd lua-resty-mysql &&
    (git checkout master; git branch -d v0.15; git checkout v0.15 -b v0.15) &&
    cd $workdir/api/v3 && rm -f mysql.lua && ln -sf ../../objs/lua-resty-mysql/lib/resty/mysql.lua . &&
    echo "lua-resty-mysql ok"
else
    echo "lua-resty-mysql ok"
fi &&

if [[ ! -f $objs/CherryPy-3.2.2/setup.py ]]; then
    echo "build CherryPy-3.2.2" &&
    cd $objs &&
    rm -rf CherryPy-3.2.2 &&
    wget $WEB_ROOT/3rdparty/CherryPy-3.2.2.tar.gz -O CherryPy-3.2.2.tar.gz &&
    tar xf CherryPy-3.2.2.tar.gz && 
    cd CherryPy-3.2.2 &&
    sudo python setup.py install &&
    echo "CherryPy-3.2.2 ok"
else
    echo "CherryPy-3.2.2 ok"
fi &&

if [[ ! -f $objs/MySQL-python-1.2.3c1/setup.py ]]; then
    echo "build MySQL-python-1.2.3c1" &&
    cd $objs &&
    rm -rf MySQL-python-1.2.3c1 &&
    wget $WEB_ROOT/3rdparty/MySQL-python-1.2.3c1.zip -O MySQL-python-1.2.3c1.zip &&
    unzip -q MySQL-python-1.2.3c1.zip && 
    cd MySQL-python-1.2.3c1 &&
    sudo python setup.py install &&
    echo "MySQL-python-1.2.3c1 ok"
else
    echo "MySQL-python-1.2.3c1 ok"
fi &&

# lua api
# @see bravoserver/trunk/src/p2p/api
echo "build lua api" &&
cd $objs/_release/nginx &&
rm -f api && ln -sf $workdir/api &&

# nginx conf
# @see bravoserver/trunk/src/p2p/api/nginx.conf
echo "build nginx conf" &&
cd $objs/_release/nginx/conf &&
rm -f nginx.conf && cp $workdir/conf/nginx.conf . &&
sed -i "s/nobody/`whoami`/g" nginx.conf &&
sed -i "s|ROOT_DIR|$objs/_release/nginx|g" nginx.conf &&

# for cherrypy
echo "create static-dir for cherrypy" &&
cd $workdir && mkdir -p static-dir &&

# for go martini
if [[ ! -f $GOPATH/src/github.com/go-martini/martini/martini.go ]]; then
    echo "go get martini" &&
    go get github.com/go-martini/martini &&
    cd $GOPATH/src/github.com/go-martini/martini &&
    (git checkout master; git branch -d v1.0; git checkout v1.0 -b v1.0) &&
    echo "go-martini ok"
else
    echo "go-martini ok"
fi &&

# for go mysql
if [[ ! -d $GOPATH/src/github.com/go-sql-driver/mysql ]]; then
    echo "go get mysql" &&
    go get github.com/go-sql-driver/mysql &&
    cd $GOPATH/src/github.com/go-sql-driver/mysql &&
    (git checkout master; git branch -d v1.2; git checkout v1.2 -b v1.2) &&
    echo "go-mysql ok"
else
    echo "go-mysql ok"
fi &&

# about
echo "for mysql-python, install:" &&
echo "      sudo yum install -y mysql-devel python-devel" &&
echo "install the database:" &&
echo "      mysql -uroot -ptest < db.sql" &&
echo "about nginx-lua(v2):" &&
echo "      vi api/v3/servers.lua # to modify the db parameters" &&
echo "      sudo killall nginx; sudo ./objs/_release/nginx/sbin/nginx" &&
echo "      sudo killall -1 nginx" &&
echo "      tailf objs/_release/nginx/logs/error.log" &&
echo "about Cherrypy:" &&
echo "      python cherrypy.api.py 8080 localhost 3306 root test srs_go" &&
echo "about go martini:" &&
echo "      go build -o objs/go.martini ./go.martini.go && ./objs/go.martini 1 8080 root:test@/srs_go"
echo "about benchmarks:" &&
echo "      ab-benchmark"
echo "access the url:" &&
echo "      http://dev:8080/api/v3/servers?action=create&mac_addr=08:00:27:EF:39:DF&ip_addr=192.168.1.173&hostname=dev"
echo "      http://dev:8080/api/v3/servers?action=get&start=0&count=10&sort=desc"
echo "access the url by cli:" &&
echo "      http://dev:8080/api/v3/servers?action=create\&mac_addr=08:00:27:EF:39:DF\&ip_addr=192.168.1.173\&hostname=dev"
echo "      http://dev:8080/api/v3/servers?action=get\&start=0\&count=10\&sort=desc"
