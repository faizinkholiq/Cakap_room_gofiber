package main

import (
	"flag"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"encoding/json"
)

type client struct{
	Username string
	Room string
} // Add more data to this type if needed

type Message map[string]interface{}

var clients = make(map[*websocket.Conn]client) // Note: although large maps with pointer-like types (e.g. strings) as keys are slow, using pointers themselves as keys is acceptable and fast
var register = make(chan *websocket.Conn)
var broadcast = make(chan string)
var unregister = make(chan *websocket.Conn)

func runHub() {
	for {
		select {
		case connection := <-register:
			username := connection.Query("username")
			room := connection.Query("room")
			clients[connection] = client{
				Username: username,
				Room: room,
			}

			message := Message{
				"username": "Bot",
				"text": "Hello " + username + ", Welcome to chat",
				"bot": true,
			}

			jsonMessage, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
			}
			
			if err := connection.WriteMessage(websocket.TextMessage, []byte(jsonMessage)); err != nil {
				log.Println("write error:", err)

				unregister <- connection
				connection.WriteMessage(websocket.CloseMessage, []byte{})
				connection.Close()
			}

			log.Println("connection registered")

		case message := <-broadcast:
			log.Println("message received:", message)

			// Send the message to all clients
			for connection := range clients {
				if err := connection.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					log.Println("write error:", err)

					unregister <- connection
					connection.WriteMessage(websocket.CloseMessage, []byte{})
					connection.Close()
				}
			}

		case connection := <-unregister:
			// Remove the client from the hub
			username := clients[connection].Username

			message := Message{
				"username": "Bot",
				"text": "Ahh, " + username + " left :(",
				"bot": true,
			}

			jsonMessage, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
			}

			log.Println(jsonMessage)
			
			if err := connection.WriteMessage(websocket.TextMessage, []byte(jsonMessage)); err != nil {
				log.Println("write error:", err)

				unregister <- connection
				connection.WriteMessage(websocket.CloseMessage, []byte{})
				connection.Close()
			}

			delete(clients, connection)

			log.Println("connection unregistered")
		}
	}
}

func main() {
	app := fiber.New()

	app.Static("/", "./public")

	app.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) { // Returns true if the client requested upgrade to the WebSocket protocol
			return c.Next()
		}
		return c.SendStatus(fiber.StatusUpgradeRequired)
	})

	go runHub()

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		// When the function returns, unregister the client and close the connection
		defer func() {
			unregister <- c
			c.Close()
		}()

		// Register the client
		register <- c

		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("read error:", err)
				}

				return // Calls the deferred function, i.e. closes the connection on error
			}

			if messageType == websocket.TextMessage {
				// Broadcast the received message
				broadcast <- string(message)
			} else {
				log.Println("websocket message received of type", messageType)
			}
		}
	}))

	addr := flag.String("addr", ":8080", "http service address")
	flag.Parse()
	log.Fatal(app.Listen(*addr))
}