#!/usr/bin/python
# -*- coding: utf-8 -*-

import sys
# reload sys model to enable the getdefaultencoding method.
reload(sys)
# set the default encoding to utf-8
# using exec to set the encoding, to avoid error in IDE.
exec("sys.setdefaultencoding('utf-8')")
assert sys.getdefaultencoding().lower() == "utf-8"

import os, json, cherrypy, threading, time, random

# the api version.as
api_version = "3.0.14"

# error code defines.
class Errors:
    success = 0

# for root, for instance: http://192.168.1.170:8085
class Root(object):
    exposed = True
    def __init__(self):
        self.api = Api()
    def GET(self):
        return json.dumps({"code":Errors.success, "urls":{"api":"the api root"}})
    def OPTIONS(self, *args, **kwargs):
        pass

# for api, for instance: http://192.168.1.170:8085/api
class Api(object):
    exposed = True
    def __init__(self):
        self.v3 = V3()
    def GET(self):
        return json.dumps({"code":Errors.success,
            "urls": {
                "v3": "the api version 3.0"
            }
        })
    def OPTIONS(self, *args, **kwargs):
        pass

# for v3, for instance: http://192.168.1.170:8085/api/v3
class V3(object):
    exposed = True
    def __init__(self):
        self.json = RESTJson()
    def GET(self):
        return json.dumps({"code":Errors.success, "urls":{
            "json": "test api for cherrypy"
        }})
    def OPTIONS(self, *args, **kwargs):
        pass

# for json, for instance: http://192.168.1.170:8085/api/v3/json
class RESTJson(object):
    exposed = True
    def __init__(self):
        pass
    def GET(self):
        return json.dumps({"code":Errors.success, "desc":"the cherrypy test."})
    def OPTIONS(self, *args, **kwargs):
        pass

# donot support use this module as library.
if __name__ != "__main__":
    raise Exception("embed not support")

# check the user options
if len(sys.argv) <= 1:
    print "Cherrypy web framework"
    print "Usage: python %s <port>"%(sys.argv[0])
    print "    port: the port to listen at."
    print "For example:"
    print "    python %s 1980"%(sys.argv[0])
    print ""
    sys.exit(1)

# parse port from user options.
port = int(sys.argv[1])
static_dir = os.path.abspath(os.path.join(os.path.dirname(sys.argv[0]), "static-dir"))

# cherrypy config.
conf = {
    'global': {
        'server.shutdown_timeout': 3,
        'server.socket_host': '0.0.0.0',
        'server.socket_port': port,
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

# start cherrypy web engine
root = Root()
cherrypy.quickstart(root, '/', conf)
