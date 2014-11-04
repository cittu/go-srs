#!/usr/bin/python
# -*- coding: utf-8 -*-

import sys
# reload sys model to enable the getdefaultencoding method.
reload(sys)
# set the default encoding to utf-8
# using exec to set the encoding, to avoid error in IDE.
exec("sys.setdefaultencoding('utf-8')")
assert sys.getdefaultencoding().lower() == "utf-8"

import os, json, cherrypy, threading, time, random, MySQLdb, datetime

# the api version.as
api_version = "3.0.14"

def enable_crossdomain():
    cherrypy.response.headers["Access-Control-Allow-Origin"] = "*"
    cherrypy.response.headers["Access-Control-Allow-Methods"] = "GET, POST, HEAD, PUT, DELETE"
    # generate allow headers for crossdomain.
    allow_headers = ["Cache-Control", "X-Proxy-Authorization", "X-Requested-With", "Content-Type"]
    cherrypy.response.headers["Access-Control-Allow-Headers"] = ",".join(allow_headers)

# error code defines.
class Errors:
    Success = 0
    
    SystemMysqlFailed = 302
    SystemMysqlEmpty = 303
    SystemMysqlInsert = 304
    SystemMysqlUnInitialized = 305

class MysqlClient:
    def __init__(self):
        self.__host = None
        self.__port = None
        self.__user = None
        self.__passwd = None
        self.__db = None
        self.__charset = None
        self.__initialized = False

    def initialize(self, host, port, user, passwd, db, charset='utf8'):
        self.__host = host
        self.__port = port
        self.__user = user
        self.__passwd = passwd
        self.__db = db
        self.__charset = charset
        self.__initialized = True

    def execute(self, sql_format_str, sql_value_tuple=None):
        (code, rows_affected, result) = (Errors.Success, 0, None)

        if not self.__initialized:
            code = Errors.SystemMysqlUnInitialized
            return (code, rows_affected, result)

        if sql_format_str is None:
            code = Errors.SystemMysqlEmpty
            return (code, rows_affected, result)

        (conn, cur) = (None, None)
        try:
            conn = self.__connection()
            cur = conn.cursor(MySQLdb.cursors.DictCursor)

            if sql_value_tuple:
                rows_affected = cur.execute(sql_format_str, sql_value_tuple)
            else:
                rows_affected = cur.execute(sql_format_str)
            result = cur.fetchall()
            conn.commit()

            return (code, rows_affected, result)
        except Exception, ex:
            code = Errors.SystemMysqlFailed
            print("sql=%s, ex=%s"%(sql_format_str, ex))
            return (code, rows_affected, result)
        finally:
            if cur is not None:
                cur.close()
            if conn is not None:
                conn.close()
                
    def execute_get_id(self, sql_format_str, sql_value_tuple=None):
        (code, id) = (Errors.Success, None)

        if not self.__initialized:
            code = Errors.SystemMysqlUnInitialized
            return (code, id)

        if sql_format_str is None:
            code = Errors.SystemMysqlEmpty
            return (code, id)

        (conn, cur) = (None, None)
        try:
            conn = self.__connection()
            cur = conn.cursor(MySQLdb.cursors.DictCursor)

            if sql_value_tuple:
                rows_affected = cur.execute(sql_format_str, sql_value_tuple)
            else:
                rows_affected = cur.execute(sql_format_str)

            if rows_affected != 1:
                code = Errors.SystemMysqlInsert
                return (code, id)

            id = conn.insert_id()
            conn.commit()

            return (code, id)
        except Exception, ex:
            code = Errors.SystemMysqlFailed
            return (code, id)
        finally:
            if cur is not None:
                cur.close()
            if conn is not None:
                conn.close()

    def __connection(self):
        # @see: http://eventlet.net/doc/modules/db_pool.html
        # @remark: we directly use the MySQLdb, for the api and cycle thread will invoke.
        conn = MySQLdb.connect(host=self.__host, port=self.__port,
            user=self.__user, passwd=self.__passwd, db=self.__db, charset=self.__charset)
        return conn

# for root, for instance: http://192.168.1.170:8085
class Root(object):
    exposed = True
    def __init__(self):
        self.api = Api()
    def GET(self):
        enable_crossdomain()
        return json.dumps({"code":Errors.Success, "urls":{"api":"the api root"}})
    def OPTIONS(self, *args, **kwargs):
        enable_crossdomain()
        pass

# for api, for instance: http://192.168.1.170:8085/api
class Api(object):
    exposed = True
    def __init__(self):
        self.v3 = V3()
    def GET(self):
        enable_crossdomain()
        return json.dumps({"code":Errors.Success,
            "urls": {
                "v3": "the api version 3.0"
            }
        })
    def OPTIONS(self, *args, **kwargs):
        enable_crossdomain()
        pass

# for v3, for instance: http://192.168.1.170:8085/api/v3
class V3(object):
    exposed = True
    def __init__(self):
        self.servers = RESTServer()
    def GET(self):
        enable_crossdomain()
        return json.dumps({"code":Errors.Success, "urls":{
            "servers": "all servers installed SRS"
        }})
    def OPTIONS(self, *args, **kwargs):
        enable_crossdomain()
        pass

# for servers, for instance: http://192.168.1.170:8085/api/v3/servers
class RESTServer(object):
    exposed = True
    def __init__(self):
        pass
    # request:
    #   action=create&mac_addr=08:00:27:EF:39:DF&ip_addr=192.168.1.173&hostname=dev
    # response:
    #   {"code":0, "id":201}
    #
    # request:
    #   action=get&start=0&count=10&sort=desc
    # response:
    #   {"code":0, "data":[
    #       {"id":200, "mac_addr":"08:00:27:EF:39:DF", "ip_addr":"192.168.1.173", "hostname":"dev"}
    #   ]}
    def GET(self, action, 
        start=0, count=10, sort='desc',
        mac_addr='08:00:27:EF:39:DF', ip_addr='192.168.1.173', hostname='dev'
    ):
        enable_crossdomain()
        if sort != "desc":
            sort = "asc"
        # GET, to query
        if action == 'get':
            (code, nb_rows, rows) = mysql.execute(
                "select * from srs_server order by server_id " + sort + " limit %s,%s",
                (int(start), int(count))
            )
            ret = {"code":code}
            if code == Errors.Success:
                ret["data"] = []
                for row in rows:
                    ret["data"].append({
                        "id": row["server_id"],
                        "mac_addr": row["server_mac_addr"],
                        "ip_addr": row["server_ip_addr"],
                        "hostname": row["server_hostname"],
                        "create": str(row["server_create_time"]),
                        "modify": str(row["server_last_modify_time"]),
                        "heartbeat": str(row["server_last_heartbeat_time"])
                    })
            return json.dumps(ret)
        # POST, to create
        else:
            now = datetime.datetime.now()
            (code, id) = mysql.execute_get_id(
                "insert into srs_server "
                "(server_mac_addr,server_ip_addr,server_hostname,server_create_time,server_last_modify_time,server_last_heartbeat_time) "
                "values(%s,%s,%s,%s,%s,%s)", 
                (mac_addr, ip_addr, hostname, now, now, now)
            )
            ret = {"code":code}
            if code == Errors.Success:
                ret["id"] = id
            return json.dumps(ret)
    def OPTIONS(self, *args, **kwargs):
        enable_crossdomain()
        pass

# donot support use this module as library.
if __name__ != "__main__":
    raise Exception("embed not support")

# check the user options
if len(sys.argv) <= 6:
    print "Cherrypy web framework"
    print "Usage: python %s <port> <db_host> <db_port> <db_user> <db_passwd> <db_name>"%(sys.argv[0])
    print "    port: the port to listen at."
    print "For example:"
    print "    python %s 8080 localhost 3306 root test srs_go"%(sys.argv[0])
    print ""
    sys.exit(1)

# parse port from user options.
(port, db_host, db_port, db_user, db_passwd, db_name) = sys.argv[1:]
static_dir = os.path.abspath(os.path.join(os.path.dirname(sys.argv[0]), "static-dir"))

# cherrypy config.
conf = {
    'global': {
        'server.shutdown_timeout': 3,
        'server.socket_host': '0.0.0.0',
        'server.socket_port': int(port),
        'tools.encode.on': True,
        'tools.staticdir.on': True,
        'tools.encode.encoding': "utf-8"
    },
    '/': {
        'tools.staticdir.dir': static_dir,
        'tools.staticdir.index': "index.html",
        # for cherrypy RESTful api support
        'request.dispatch': cherrypy.dispatch.MethodDispatcher()
    }
}

# initialize kernel objects.
mysql = MysqlClient()
mysql.initialize(db_host, int(db_port), db_user, db_passwd, db_name)

# start cherrypy web engine
root = Root()
cherrypy.quickstart(root, '/', conf)
