package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var connMap = sync.Map{}

var upgrader = websocket.Upgrader{} // use default options

func newConnectionHandler(w http.ResponseWriter, r *http.Request) {

	connName := mux.Vars(r)["connection_name"]
	log.Println("new conection: ", connName)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	connMap.Store(connName, conn)

	defer func() {
		connMap.Delete(connName)
		conn.Close()
	}()

	for {
	}
}

func requestForwarderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectionName := vars["connection_name"]
	addr := vars["addr"]

	log.Println("connection: ", connectionName, "  addr: ", addr)

	var data map[string]interface{}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.Write([]byte("internal server error"))
		return
	}

	conn, ok := connMap.Load(connectionName)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("connection not found"))
		return
	}

	var response map[string]interface{}

	connection := conn.(*websocket.Conn)
	connection.WriteJSON(map[string]interface{}{"url": addr, "body": data})
	connection.ReadJSON(&response)

	responsePayload, _ := json.Marshal(response)

	w.Write(responsePayload)

}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/connection/{connection_name}", newConnectionHandler)
	r.HandleFunc("/{connection_name}/{addr:.+}", requestForwarderHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
