package main
import (
    "encoding/json"
    "net/http"
    "github.com/go-martini/martini"
)

func main() {
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
    http.ListenAndServe(":8080", m)
}
