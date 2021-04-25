package dnet

import (
	"fmt"
	"strings"
	"time"

	"github.com/dilungasr/tanzanite/tzcrypt"

	"github.com/mitchellh/mapstructure"
)

// Written by Dilunga SR

// Broadcast sends data to all execept the sender
func (c *Context) Broadcast(statusAndData ...interface{}) {
	dataIndex := 0
	statusCode := 200

	// take user dataIndex from the statusAndCode and assign them to the above variables
	assignData(&dataIndex, &statusCode, statusAndData, "Broadcast")
	// prepare the response to be sent to the client
	res := Response{c.action, statusCode, statusAndData[dataIndex], c.ID}

	// pass to all hub contexts to send to all other contexts
	for context := range c.hub.contexts {
		// send to other contexts except this
		if context != c {
			select {
			case context.send <- res:
			default:
				deleteContext(context)
			}
		}
	}
}

// All sends to anyone connected to the websocket Dnet instance including the sender of the message
func (c *Context) All(statusAndData ...interface{}) {
	dataIndex := 0
	statusCode := 200

	// take user dataIndex from the statusAndCode and assign them to the above variables
	assignData(&dataIndex, &statusCode, statusAndData, "Broadcast")
	// prepare the response to be sent to the client
	res := Response{c.action, statusCode, statusAndData[dataIndex], c.ID}

	//

	// pass to all hub contexts to send to all other contexts
	for context := range c.hub.contexts {
		select {
		case context.send <- res:
		default:
			deleteContext(context)
		}
	}
}

// SendBack sends data back to the sender
func (c *Context) SendBack(statusAndData ...interface{}) {
	dataIndex := 0
	statusCode := 200
	assignData(&dataIndex, &statusCode, statusAndData, "SendBack")

	res := Response{c.action, statusCode, statusAndData[dataIndex], c.ID}

	// send back to the client
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := c.conn.WriteJSON(res); err != nil {
		panic(err)
	}
}

// Send sends to only client
func (c *Context) Send(ID string, statusAndData ...interface{}) {
	dataIndex := 0
	statusCode := 200

	// take user dataIndex from the statusAndCode and assign them to the above variables
	assignData(&dataIndex, &statusCode, statusAndData, "Send")

	//the response to be sent to the client
	res := Response{c.action, statusCode, statusAndData[dataIndex], c.ID}

	// find the user to which the dataIndex should be sent to
	for context := range c.hub.contexts {
		if context.ID == ID {
			select {
			case context.send <- res:
			default:
				deleteContext(context)
			}
		}
	}

}

// Multicast sends to the given users IDs (useful for sharing something to multiple users
func (c *Context) Multicast(userIDs []string, statusAndData ...interface{}) {
	dataIndex := 0
	statusCode := 200

	// take user dataIndex from the statusAndCode and assign them to the above variables
	assignData(&dataIndex, &statusCode, statusAndData, "Send")

	//the response to be sent to the client
	res := Response{c.action, statusCode, statusAndData[dataIndex], c.ID}

	for _, userID := range userIDs {
		// find the matching context
		for context := range c.hub.contexts {
			if userID == context.ID {
				select {
				case context.send <- res:
				default:
					deleteContext(context)
				}

				break
			}
		}
	}
}

/*
   -----------------------------------------------
    ROOM  METHODS GOES HERE
   -----------------------------------------------

*/

// RoomAll sends to the members of the room. (useful for chat rooms.. and sending data to all people under the same role or cartegory )
func (c *Context) RoomAll(ID string, statusAndCode ...interface{}) {
	dataIndex := 0
	statusCode := 200

	// take user dataIndex from the statusAndCode and assign them to the above variables
	assignData(&dataIndex, &statusCode, statusAndCode, "Send")

	//the response to be sent to the client
	res := Response{c.action, statusCode, statusAndCode[dataIndex], c.ID}

	//    find the room and broadcast to all the room members
	for roomID, contexts := range c.hub.rooms {
		if roomID == ID {
			//   broadcast to all the members in the room
			for _, context := range contexts {
				select {
				case context.send <- res:
				default:
					deleteContext(context)
				}
			}

			// break out of the loop if found the room and finished sending to all members of the room
			break
		}
	}
}

// RoomBroadcast sends to all members of the registered room except the sender
func (c *Context) RoomBroadcast(ID string, statusAndCode ...interface{}) {
	dataIndex := 0
	statusCode := 200

	// take user dataIndex from the statusAndCode and assign them to the above variables
	assignData(&dataIndex, &statusCode, statusAndCode, "Send")

	//the response to be sent to the client
	res := Response{c.action, statusCode, statusAndCode[dataIndex], c.ID}

	//    find the room and broadcast to all the room members
	for roomID, contexts := range c.hub.rooms {
		if roomID == ID {
			//   broadcast to all the members in the room
			for _, context := range contexts {
				// send to all members of the room execept the sender
				if context != c {
					select {
					case context.send <- res:
					default:
						deleteContext(context)
					}
				}
			}
		}
	}
}

// CreateRoom is for creating a new room.... if it finds a room exist it only adds the given the room
func (c *Context) CreateRoom(roomID string, usersIDS ...string) {
	isReg := false

	// if the room is already registered
	for room, contexts := range c.hub.rooms {
		if room == roomID {
			isReg = true
			// find active user contexts to add to the room
			for context := range c.hub.contexts {
				for _, id := range usersIDS {
					if id == context.ID {

						// do the room found already added in the room?
						found := false

						//check if the context already added in the room
						for _, roomContext := range contexts {
							if roomContext.ID == id {
								found = true
								break
							}
						}

						// add to the room if hasn't already
						if !found {
							contexts = append(contexts, context)
						}
						break
					}
				}
			}
		}
	}

	// do not the code below if the room exist
	if isReg {
		return
	}

	// if the room not found .... create a new room
	activeUsers := []*Context{}
	for context := range c.hub.contexts {
		for _, id := range usersIDS {
			// if finds an active user
			if id == context.ID {
				activeUsers = append(activeUsers, context)
				break
			}
		}
	}

	// create an active room only when there are active room members
	if len(activeUsers) > 0 {
		c.hub.rooms[roomID] = activeUsers
	}
}

// Next pushes the next middleware in the list
func (c *Context) Next() {
	c.goNext = true
}

// Rooms assigns this connnection to the chatrooms it relates to
func (c *Context) Rooms(roomsIDs ...string) {

	// if user has rooms
	if len(roomsIDs) > 0 {

		for _, room := range roomsIDs {
			// find if already added and only append user if found
			isReg := false
			for registeredRoom, contexts := range c.hub.rooms {
				if registeredRoom == room {
					isReg = true

					//  check if the context already added to the room
					found := false
					for _, context := range contexts {
						if context == c {
							found = true
						}
					}

					// add the context to the room if no registered
					if !found {
						// append user to the rooms
						contexts = append(contexts, c)

					}
					break
				}
			}

			//    if chat room is added for the first time in the hub
			if !isReg {
				c.hub.rooms[room] = []*Context{c}
			}
		}
	}
}

/*
   -----------------------------------------------
   TICKET METHODS GEOS HERE
   -----------------------------------------------

*/

// ticket is for converting the received golang interface{} data to ticket ..... get the fields
type clientTicket struct {
	Ticket string `json:"ticket"`
}

// AuthTicket authenticates the ticket and  return the user id to associate and query in the database
func (c *Context) AuthTicket() (ID string, valid bool) {
	var ticketFromClient clientTicket

	c.Binder(&ticketFromClient)

	// split the ticket into parts
	ticketParts := strings.Split(ticketFromClient.Ticket, ",")
	if len(ticketParts) == 0 || len(ticketParts) > 3 {
		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		c.conn.WriteJSON(Response{c.action, 400, "Bad request", c.ID})
		return "", false
	}

	//exctract the encrypted pieces
	encID := ticketParts[0]
	encIP := ticketParts[1]
	encExpireTime := ticketParts[2]

	// decrypt the data
	ID = tzcrypt.Decrypter(encID, Router1.ticketSecrete, Router1.ticketIV)
	IP := tzcrypt.Decrypter(encIP, Router1.ticketSecrete, Router1.ticketIV)
	expireTime := tzcrypt.Decrypter(encExpireTime, Router1.ticketSecrete, Router1.ticketIV)

	// validate the ticket
	valid = false
	for i, ticket := range Router1.tickets {
		if ticket.id == ID && ticket.ip == IP && ticket.expireTime == expireTime {
			//  compare the expireTime and this time to see if the ticket expired or not
			ticketExpireTime, err := time.Parse(time.RFC3339, ticket.expireTime)
			if err != nil {
				panic(err)
			}

			// if the ticket expired
			if time.Now().Local().After(ticketExpireTime) {
				//  delete the ticket
				Router1.tickets = append(Router1.tickets[:i], Router1.tickets[i+1:]...)
			} else {
				// if the ticket is valid
				valid = true
				break
			}
		}
	}

	// if the ticket is valid
	// set the ip address in the context
	if valid {
		c.IP = IP
		c.authed = true
		c.ID = ID
		return ID, valid
	}

	// if not valid ... close the connection
	c.SendBack(401, "Please login to access this resource")
	c.Dispose()
	return "", false
}

// Authed  tells if the connection is authenticated or not
func (c *Context) Authed() bool {
	return c.ID != ""
}

/*
   -----------------------------------------------
    FIRE METHODS GOE HERE
   -----------------------------------------------

*/

// Fire sets which action to fire to the client. It's recommended to keep the action in path form to maintain maintain uniformity,
// If you do not set the action to fire, the action you listened for it will be fired backward to the client too.
func (c *Context) Fire(action string) {
	c.action = action
}

/*
   -----------------------------------------------
    BINDING METHODS GOE HERE
   -----------------------------------------------

*/

// Binder is for extracting data from the client and storing it to the passed pointer v
func (c *Context) Binder(v interface{}) {
	if err := mapstructure.Decode(c.Data, v); err != nil {
		panic(err)
	}
}

/*
   -----------------------------------------------
    CLOSE METHOD GOES HERE
   -----------------------------------------------

*/

// Dispose discards the client connection without calling LastSeen for saving any lastSeen info for the clinet connection
// Useful for expired unauthorized client connection
func (c *Context) Dispose() {
	c.disposed = true
	c.hub.unregister <- c
	c.conn.Close()
}

// Logout calls the LastSeen function to ensure user last seen data is saved before discarding the client connection
func (c *Context) Logout() {
	c.loggedout = true
	// set the last seen of the clinet connection
	Router1.lastSeen(c)

	// unregister the clinet context
	c.hub.unregister <- c
	c.conn.Close()
}

/*
  -----------------------------------------------------------
  |  SETTING VALUES IN AND GETTING VALUES FROM CONNECTION   |
  ----------------------------------------------------------
*/

// Set stores value in the connection.
func (c *Context) Set(key string, val interface{}) {
	c.values[key] = val
}

// Get gets data stored in the connection
func (c *Context) Get(key string) (val interface{}, err error) {
	val, ok := c.values[key]
	if !ok {
		return val, fmt.Errorf("Dnet: value not registered in the connection")
	}

	return val, nil

}

// By Dilunga SR
