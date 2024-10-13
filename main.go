package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	Clients = make(map[chan Notif]struct{})
	clientMutex sync.Mutex
)

type Notif struct {
	Username  string `form:"username" json:"username" binding:"required"`
	Activity  string `form:"activity" json:"activity" binding:"required"`
	Post      string `form:"post" json:"post" binding:"required"`
	Timestamp time.Time `form:"timestamp" json:"timestamp"`
}

func main() {
	router := gin.Default()
	router.GET("/notif", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "text/event-stream")
		ctx.Header("Cache-Control", "no-cache")
		ctx.Header("Connection", "keep-alive")

		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type")

		log.Println("Client Connected")
		eventChan := make(chan Notif)
		clientMutex.Lock()
		Clients[eventChan] = struct{}{}
		clientMutex.Unlock()

		defer func() {
			clientMutex.Lock()
			delete(Clients, eventChan)
			clientMutex.Unlock()
			close(eventChan)
		}()

		for {
			data := <-eventChan
			log.Println("Sending Data to Client", data)
			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Println("Error marshaling data:", err)
				continue
			}

			fmt.Fprintf(ctx.Writer, "data: %s\n\n", jsonData)
			ctx.Writer.Flush()
		}
	})

	router.POST("/send-notif", func(ctx *gin.Context) {
		var notif Notif
		ctx.Bind(&notif)
		notif.Timestamp = time.Now()
		Broadcast(notif)
		ctx.JSON(http.StatusOK, gin.H{"message": "Data Sent to Clients"})
	})

	if err := router.Run((":8080")); err != nil {
		log.Panicln(err)
	}
}

func Broadcast(data Notif) {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	for client := range Clients {
		client <- data
	}
}
