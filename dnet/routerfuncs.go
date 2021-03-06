package dnet

import (
	"net/http"
	"time"

	tz "github.com/dilungasr/tanzanite"
	"github.com/dilungasr/tanzanite/tzcrypt"
)

// On method is adding the Event handlers to the router
func On(action string, handlers ...ActionHandler) {
	Router1.actionHandlers[action] = handlers
}

// Router is for grouping the actions by matching their paths
func Router(path string) RouterMatcher {
	return RouterMatcher{path}
}

// Use is for adding middlewares to the root of the dnet action path
func Use(handlers ...ActionHandler) {
	Router1.routeMatchers["/"] = append(Router1.routeMatchers["/"], handlers...)
}

// SendTicket sends an encrypted ticket to the use and saves the clean one the router
func SendTicket(r *http.Request, w http.ResponseWriter, ID string) {
	// if the ticketSecrete and the iv set...... wer are ready to go
	secreteKey := Router1.ticketSecrete
	iv := Router1.ticketIV

	// plain data(not encrypted)
	expireTimeBytes, err := time.Now().Local().Add(Router1.ticketAge).MarshalText()
	if err != nil {
		panic(err)
	}
	expireTime := string(expireTimeBytes)
	// get the user IP
	IP, err := tz.GetIP(r)
	if err != nil {
		tz.Send(w, 400, tz.Map{"msg": "Bad request!"})
		return
	}

	// encrypt the tikcet data to be sent to the client
	newTicket := ID + "," + IP + "," + expireTime

	encTicket := tzcrypt.Encrypter(newTicket, secreteKey, iv)

	// save the ticket in the Router1
	Router1.tickets = append(Router1.tickets, newTicket)

	// send the ticket to the client
	tz.Send(w, tz.Map{"ticket": encTicket})

}
