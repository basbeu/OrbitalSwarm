// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gs

type wsMessage struct {
	data   []byte
	client *Client
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	wsReceived chan wsMessage

	// Outbound messages for the clients.
	wsBroadcast chan []byte

	// Callbacks
	onClientJoin      func() []byte
	onMessageReceived func([]byte) []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub(onClientJoin func() []byte, onMessageReceived func([]byte) []byte) *Hub {
	return &Hub{
		wsReceived: make(chan wsMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),

		onClientJoin:      onClientJoin,
		onMessageReceived: onMessageReceived,
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			client.send <- h.onClientJoin()
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.wsBroadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case message := <-h.wsReceived:
			select {
			case message.client.send <- h.onMessageReceived(message.data):
			default:
				close(message.client.send)
				delete(h.clients, message.client)
			}
		}
	}
}
