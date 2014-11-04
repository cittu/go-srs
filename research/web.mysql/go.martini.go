package main
import (
    "os"
    "fmt"
    "strconv"
    "runtime"
    "encoding/json"
    "net/http"
    "github.com/go-martini/martini"
)

func main() {
    fmt.Println("go martini web server")
    if len(os.Args) <= 2 {
        fmt.Println("Usage:", os.Args[0], "<cpus> <port>")
        fmt.Println("   port: the port to listen at.")
        fmt.Println("For example:")
        fmt.Println("   ", os.Args[0], 1, 8080)
        return
    }
    fmt.Println("use cpus", os.Args[1])
    fmt.Println("listen at", os.Args[2])
    
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
    runtime.GOMAXPROCS(nb_cpus)
    
    m := martini.Classic()
    m.Get("/api/v3/json", func(res http.ResponseWriter) string {
        res.Header().Add("Server", "GoMartini/1.0")
        b, err := json.Marshal(map[string]interface{}{
            "code": 0, "desc": "test api for go martini",
        })
        if err != nil {
            // error
        }
        return string(b)
    })
    http.ListenAndServe(fmt.Sprintf(":%d", listen_port), m)
}
