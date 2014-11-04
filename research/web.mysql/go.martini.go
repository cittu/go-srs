package main
import (
    "os"
    "fmt"
    "strconv"
    "runtime"
    "time"
)
import (
    "encoding/json"
    "net/http"
    "github.com/go-martini/martini"
)
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    fmt.Println("go martini web server")
    if len(os.Args) <= 3 {
        fmt.Println("Usage:", os.Args[0], "<cpus> <port> <db_url>")
        fmt.Println("   cpus: the cpus to use for go.")
        fmt.Println("   port: the port to listen at.")
        fmt.Println("   db_url: the db url to use for connect mysql.")
        fmt.Println("For example:")
        fmt.Println("   ", os.Args[0], 1, 8080, "root:test@/srs_go")
        return
    }
    fmt.Println("use cpus", os.Args[1])
    fmt.Println("listen at", os.Args[2])
    fmt.Println("db_url is", os.Args[3])
    
    nb_cpus, err := strconv.Atoi(os.Args[1])
    if err != nil {
        fmt.Println("invalid option cpus", os.Args[1])
        return
    }
    listen_port, err := strconv.Atoi(os.Args[2])
    if err != nil {
        fmt.Println("invalid option port", os.Args[2])
        return
    }
    db_url := os.Args[3]
    
    runtime.GOMAXPROCS(nb_cpus)
    
    // request:
    //   action=create&mac_addr=08:00:27:EF:39:DF&ip_addr=192.168.1.173&hostname=dev
    // response:
    //   {"code":0, "id":201}
    //
    // request:
    //   action=get&start=0&count=10&sort=desc
    // response:
    //   {"code":0, "data":[
    //       {"id":200, "mac_addr":"08:00:27:EF:39:DF", "ip_addr":"192.168.1.173", "hostname":"dev"}
    //   ]}
    m := martini.Classic()
    m.Get("/api/v3/servers", func(res http.ResponseWriter, req *http.Request) string {
        query := req.URL.Query()
        res.Header().Add("Server", "GoMartini/1.0")
        ret := map[string]interface{}{"code":0}
        if query.Get("sort") != "desc" {
            query.Set("sort", "asc")
        }
        
        db, err := sql.Open("mysql", db_url)
        if err != nil {
            return "open mysql server failed"
        }
        defer db.Close()
        
        if query.Get("action") == "get" {
            rows, err := db.Query("select server_id, server_mac_addr, server_ip_addr, " +
                    "server_hostname, server_create_time, server_last_modify_time, " +
                    "server_last_heartbeat_time from srs_server order by server_id " + query.Get("sort") + " limit ?,?",
                query.Get("start"), query.Get("count"))
            if err != nil {
                return fmt.Sprintf("query data failed. %s", err)
            }
            data := make([]interface{}, 0)
            for rows.Next() {
                obj := make(map[string]interface{})
                var server_id int
                var server_mac_addr, server_ip_addr, server_hostname string
                var server_create_time, server_last_modify_time, server_last_heartbeat_time string
                if err := rows.Scan(&server_id, &server_mac_addr, 
                    &server_ip_addr, &server_hostname, &server_create_time, 
                    &server_last_modify_time, &server_last_heartbeat_time); err != nil {
                    return fmt.Sprintf("get server_last_heartbeat_time failed. %s", err)
                }
                obj["id"] = server_id
                obj["mac_addr"] = server_mac_addr
                obj["ip_addr"] = server_ip_addr
                obj["hostname"] = server_hostname
                obj["create"] = server_create_time
                obj["modify"] = server_last_modify_time
                obj["heartbeat"] = server_last_heartbeat_time
                data = append(data, obj)
            }
            ret["data"] = data
        } else {
            now := time.Now()
            result, err := db.Exec(
                "insert into srs_server " +
                "(server_mac_addr,server_ip_addr,server_hostname,server_create_time,server_last_modify_time,server_last_heartbeat_time) " +
                "values(?,?,?,?,?,?)", 
                query.Get("mac_addr"), query.Get("ip_addr"), query.Get("hostname"), now, now, now)
            if err != nil {
                return fmt.Sprintf("insert data failed. %s", err)
            }
            ret["id"], err = result.LastInsertId()
            if err != nil {
                return fmt.Sprintf("insert data id failed. %s", err)
            }
        }
        b, err := json.Marshal(ret)
        if err != nil {
            return fmt.Sprintf("to json failed. %s", err)
        }
        return string(b)
    })
    http.ListenAndServe(fmt.Sprintf(":%d", listen_port), m)
}
