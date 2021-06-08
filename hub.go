package main

type Message struct {
	data []byte
	room string
}

type Subscription struct {
	conn *Connection
	room string
}

type Hub struct {
	broadcast  chan Message
	unregister chan Subscription
	register   chan Subscription
	rooms      map[string]map[*Connection]bool
}

var h = Hub{
	broadcast:  make(chan Message),
	register:   make(chan Subscription),
	unregister: make(chan Subscription),
	rooms:      make(map[string]map[*Connection]bool),
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		unregister: make(chan Subscription),
		register:   make(chan Subscription),
		rooms:      make(map[string]map[*Connection]bool),
	}
}
func (h *Hub) run() {
	for {
		select {
		case s := <-h.register:
			connections := h.rooms[s.room]
			if connections == nil {
				connections = make(map[*Connection]bool)
				h.rooms[s.room] = connections
			}
			h.rooms[s.room][s.conn] = true
		case s := <-h.unregister:
			connections := h.rooms[s.room]
			if connections != nil {
				if _, ok := connections[s.conn]; ok {
					delete(connections, s.conn)
					close(s.conn.send)
					if len(connections) == 0 {
						delete(h.rooms, s.room)
					}
				}
			}
		case m := <-h.broadcast:
			connections := h.rooms[m.room]
			for c := range connections {
				select {
				case c.send <- m.data:
				default:
					close(c.send)
					delete(connections, c)
					if len(connections) == 0 {
						delete(h.rooms, m.room)
					}
				}
			}

		}
	}

}
