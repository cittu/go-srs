package main
import (
    "os"
    "fmt"
    "encoding/json"
    "net/http"
    "github.com/go-martini/martini"
)

func main() {
    fmt.Println("go martini web server")
    if len(os.Args) <= 1 {
        fmt.Println("Usage:", os.Args[0], "<port>")
        fmt.Println("   port: the port to listen at.")
        fmt.Println("For example:")
        fmt.Println("   ", os.Args[0], 8080)
        return
    }
    fmt.Println("listen at", os.Args[1])
    
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
    http.ListenAndServe(":" + os.Args[1], m)
}
