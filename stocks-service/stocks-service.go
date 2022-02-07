package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
	"github.com/julienschmidt/httprouter"
)

var (
	logger  = log.New(os.Stdout, "", 0)
	address = getEnvVar("ADDRESS", ":8181")
)

const storename string = "statestore"

type productStock struct {
	Id    string
	stock int
}

func main() {

	s := daprd.NewService(address)

	client, err := dapr.NewClient()
	if err != nil {
		log.Fatalf("Error creating client")
	}

	s.AddBindingInvocationHandler("get", func(ctx context.Context, in *common.BindingEvent) (out []byte, err error) {
		itemId := in.Metadata["itemId"]

		item, err := client.GetStateWithConsistency(context.TODO(), storename, itemId, map[string]string{}, dapr.StateConsistencyStrong)
		if err != nil {
			return nil, err
		}

		stock := &productStock{}
		json.Unmarshal(item.Value, stock)

		return nil, nil
	})

	// err = s.AddServiceInvocationHandler("get", func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {

	// 	json.Unmarshal()

	// 	client.GetStateWithConsistency(context.TODO(), storename, ps.ByName("itemId"), map[string]string{}, dapr.StateConsistencyStrong)

	// 	return nil, nil
	// })

	if err != nil {
		logger.Fatalf("error starting service: %v", err)
		os.Exit(1)
	}

	// start the service
	if err := s.Start(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("error starting service: %v", err)
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

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
