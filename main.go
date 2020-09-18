package main

import (
    "fmt"
    "net/http"
    "os"
    "context"
    "encoding/json"
    "io/ioutil"
    "strconv"
    "github.com/jackc/pgx/v4"
    "github.com/gorilla/mux"
)

type Object_json struct {
	Peer_id int64   `json:"peer_id"`
	Text    string  `json:"text"`
}

type Json_json struct {
  	Type   string        `json:"type"`
	Object Object_json   `json:"object"`
}


func hashHandler(w http.ResponseWriter, r *http.Request) {

    // Read body
    b, err := ioutil.ReadAll(r.Body)
    defer r.Body.Close()
    if err != nil {
	    http.Error(w, err.Error(), 500)
	    return
    }

    vars := mux.Vars(r)
    hash := vars["hash"]

    var var_Json = &Json_json{}
    var b_b = []byte(b)
    err = json.Unmarshal(b_b, var_Json)
	if err != nil {
        fmt.Fprintf(os.Stderr,"Unable to parse json: %v\n", err)
    }

    fmt.Println(b)

    os.Setenv("DATABASE_URL", "postgresql://user:password@host/db")
    conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
        os.Exit(1)
    }
    defer conn.Close(context.Background())

    var id int64
    var response string
    var token string

    fmt.Println(var_Json.Object.Text);

    err = conn.QueryRow(context.Background(), "SELECT bots.id, bots.token, triggers.response FROM bots left join triggers ON bots.id=triggers.bot_id WHERE bots.hash=$1 AND trigger_name=$2", hash, var_Json.Object.Text).Scan(&id, &token, &response)
    if err != nil {
        fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
    } else {
        fmt.Println(var_Json.Object.Peer_id)
        pear_id_str := strconv.FormatInt(var_Json.Object.Peer_id, 10)
        resp, err := http.Get("https://api.vk.com/method/messages.send?access_token="+token+"&message="+response+"&peer_id="+pear_id_str+"&v=5.87")
        
        if err != nil {
            fmt.Fprintf(os.Stderr, "Unable to make GET request" ,err)
        }

        body, err := ioutil.ReadAll(resp.Body)

        if err != nil {
            fmt.Fprintf(os.Stderr, "Unable to read body response vk",err)
        }

    }

    w.Write([]byte("OK"))
}

func main() {
    router := mux.NewRouter()
    router.HandleFunc("/dev/{hash}", hashHandler)
    http.Handle("/",router)

    fmt.Println("Server is listening...")
    http.ListenAndServe(":8080", nil)
}
