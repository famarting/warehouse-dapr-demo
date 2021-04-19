package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/julienschmidt/httprouter"
	// daprd "github.com/dapr/go-sdk/service/grpc"
)

const storename string = "statestore"

func main() {
	// // create a Dapr service server
	// s, err := daprd.NewService(":50001")
	// if err != nil {
	// 	log.Fatalf("failed to start the server: %v", err)
	// }
	// // start the server
	// if err := s.Start(); err != nil {
	// 	log.Fatalf("server error: %v", err)
	// }

	client, err := dapr.NewClient()
	if err != nil {
		log.Fatalf("Error creating client")
	}

	router := httprouter.New()
	router.GET("/api/stocks/:itemId", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		item, err := client.GetState(context.TODO(), storename, ps.ByName("itemId"))
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(200)
			w.Write(item.Value)
		}
	})

	router.POST("/api/stocks/:itemId/:quantity", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		itemID := ps.ByName("itemId")
		item, err := client.GetState(context.TODO(), storename, itemID)
		if err != nil {
			w.WriteHeader(500)
		} else {
			new, err := strconv.Atoi(ps.ByName("quantity"))
			if err != nil {
				w.WriteHeader(500)
				return
			}
			old, err := strconv.Atoi(string(item.Value))
			if err != nil {
				w.WriteHeader(500)
				return
			}
			final := new + old
			finalBytes := []byte(strconv.Itoa(final))
			client.SaveState(context.TODO(), storename, itemID, finalBytes)
			w.WriteHeader(200)
			w.Write(finalBytes)
		}
	})

	log.Println("Server running on localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", router))

}
