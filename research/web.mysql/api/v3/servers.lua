local cjson = require "cjson"
local mysql = require "mysql"

local ret = {}
ret["code"] = 0
ret["msg"] = "web test for openresty"

ngx.say(cjson.encode(ret))
ngx.exit(ngx.HTTP_OK)
