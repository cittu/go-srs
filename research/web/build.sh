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
    echo "build ngx_openresty-1.7.0.1" &&
    cd $objs &&
    rm -rf ngx_openresty-1.7.0.1 &&
    wget $WEB_ROOT/3rdparty/ngx_openresty-1.7.0.1.tar.gz -O ngx_openresty-1.7.0.1.tar.gz &&
    tar xf ngx_openresty-1.7.0.1.tar.gz && 
    cd ngx_openresty-1.7.0.1 &&
    ./configure --prefix=$release \
        --with-luajit --with-http_stub_status_module --without-http_redis2_module \
        --with-http_iconv_module --with-http_mp4_module --with-http_flv_module --with-http_realip_module &&
    echo "dynamic link lua" &&
    make && make install &&
    echo "static link lua" &&
    rm -f build/nginx-1.7.0/objs/nginx &&
    sed -i "s|-lluajit-5.1|$release/luajit/lib/libluajit-5.1.a|g" build/nginx-1.7.0/objs/Makefile &&
    make && make install &&
    echo "ngx_openresty-1.7.0.1 ok"
else
    echo "ngx_openresty-1.7.0.1 ok"
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
echo "go get martini" &&
go get github.com/go-martini/martini &&
cd $GOPATH/src/github.com/go-martini/martini &&
(git checkout master; git branch -d 1.0; git checkout v1.0 -b 1.0) &&

# about
echo "about nginx-lua(v2):" &&
echo "      sudo killall nginx; sudo ./objs/_release/nginx/sbin/nginx" &&
echo "      sudo killall -1 nginx" &&
echo "      tailf objs/_release/nginx/logs/error.log" &&
echo "      http://dev:8080/api/v3/json" &&
echo "about Cherrypy:" &&
echo "      python cherrypy.api.py 8080" &&
echo "about go martini:" &&
echo "      go build -gcflags '-N -l' -o objs/go.martini ./go.martini.go && ./objs/go.martini 1 8080"
echo "about benchmarks:" &&
echo "      ab-benchmark"
