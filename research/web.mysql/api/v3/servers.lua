local cjson = require "cjson"
local mysql = require "mysql"

local db, err = mysql:new()
if not db then
    ngx.say("failed to instantiate mysql: ", err)
    return
end

db:set_timeout(10000) -- 10 sec

--------------------------------------------------
-- db: localhost 
-- port: 3306 
-- user: root 
-- password: test 
-- db_name: srs_go
--------------------------------------------------
local ok, err, errno, sqlstate = db:connect{
    host = "127.0.0.1",
    port = 3306,
    database = "srs_go",
    user = "root",
    password = "test",
    max_packet_size = 1024 * 1024 }

if not ok then
    ngx.say("failed to connect: ", err, ": ", errno, " ", sqlstate)
    return
end

local params = ngx.req.get_uri_args()
local ret = {}
ret["code"] = 0
-- request:
--   action=create&mac_addr=08:00:27:EF:39:DF&ip_addr=192.168.1.173&hostname=dev
-- response:
--   {"code":0, "id":201}
--
-- request:
--   action=get&start=0&count=10&sort=desc
-- response:
--   {"code":0, "data":[
--       {"id":200, "mac_addr":"08:00:27:EF:39:DF", "ip_addr":"192.168.1.173", "hostname":"dev"}
--   ]}
if (params.action == "get") then
    local res, err, errno, sqlstate = 
        db:query("select * from srs_server order by server_id " 
            .. params.sort .. " limit " .. params.start .. "," .. params.count)
    if not res then
        ngx.say("bad result: ", err, ": ", errno, ": ", sqlstate, ".")
        return
    end
    
    ret["data"] = {}
    for i, name in ipairs(res) do
        ret["data"][i] = {}
        ret["data"][i]["id"] = name["server_id"]
        ret["data"][i]["mac_addr"] = name["server_mac_addr"]
        ret["data"][i]["ip_addr"] = name["server_ip_addr"]
        ret["data"][i]["hostname"] = name["server_hostname"]
        ret["data"][i]["create"] = name["server_create_time"]
        ret["data"][i]["modify"] = name["server_last_modify_time"]
        ret["data"][i]["heartbeat"] = name["server_last_heartbeat_time"]
    end
else
    local now = os.date("%Y-%m-%d %H:%M:%S")
    local res, err, errno, sqlstate = 
        db:query("insert into srs_server " 
            .. "(server_mac_addr,server_ip_addr,server_hostname,server_create_time,server_last_modify_time,server_last_heartbeat_time) "
            .. "values('" .. params.mac_addr .. "','" .. params.ip_addr .. "','" 
            .. params.hostname .. "','" .. now .. "','" 
            .. now .. "','" .. now .. "')")
    if not res then
        ngx.say("bad result: ", err, ": ", errno, ": ", sqlstate, ".")
        return
    end
    ret["id"] = res["insert_id"]
end

local ok, err = db:close()
if not ok then
    ngx.say("failed to close: ", err)
    return
end

ngx.say(cjson.encode(ret))
ngx.exit(ngx.HTTP_OK)
