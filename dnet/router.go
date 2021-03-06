package dnet

import (
	"regexp"
	"time"

	"github.com/dilungasr/tanzanite/tzcrypt"
)

//ActionHandler is a function wich receives Ctx
type ActionHandler func(*Context)

// Options is used to take all the
type Options struct {
	TicketAge time.Duration
	MaxSize   int64
}

// MainRouter is for routing websocket actions
type MainRouter struct {
	// routeMatchers is matching the middleware the given path prefixes
	routeMatchers map[string][]ActionHandler

	// actionHandlers are handlers to executed for the particular action
	actionHandlers map[string][]ActionHandler
	// lastSeenHandler is called when the connection gets closed
	lastSeenHandler ActionHandler
	// tells wether use set the last seen or not
	isLastSeenHandlerSet bool
	// tickets is the store place of all tickets given to the user before they expire
	tickets []string
	// ticketKey is the key for cryptography of the ticket
	ticketIV []byte

	// max message size
	maxSize int64

	// ticketSecrete is the secrete key for tecket encryption
	ticketSecrete string
	// ticketAge is the time the ticket expires
	ticketAge time.Duration
}

//Route is for routing websocket actions based on the incoming action
func (r *MainRouter) Route(IncomingAction string, context *Context) {
	if len(r.routeMatchers) > 0 {
		for path, handlers := range r.routeMatchers {

			// the incoming action must start with the router matcher's path
			regex := regexp.MustCompile("^" + path)

			//check if the incoming action matches the path in the beginning
			if regex.MatchString(IncomingAction) {
				for _, handler := range handlers {
					// set the goNext... c.Next() should be called to change the goNext value to true for the middleware to move to the next one
					// Otherwise the middleware will not proceed to the next middlware.
					context.goNext = false
					// call the handler
					handler(context)

					// stop if the has not passed the middleware... ie c.Next() has not been called
					if !context.goNext {
						return
					}
				}
			}
		}
	}

	//  go on to  the normal event handlers
	// isMatch is for generating 404 error if the message didn't find any action
	isMatch := false
	for handlerAction, handlers := range r.actionHandlers {
		// check for route matching
		if handlerAction == IncomingAction {
			isMatch = true

			// if there are middlewares
			if len(handlers) > 1 {
				for _, handler := range handlers {
					// set the goNext... c.Next() should be called to change the goNext value to true for the middleware to move to the next one
					// Otherwise the middleware will not proceed to the next middlware.
					context.goNext = false
					// call the handler
					handler(context)

					// stop if the has not passed the middleware... ie c.Next() has not been called
					if !context.goNext {
						return
					}

				}
			} else if len(handlers) == 1 {

				//if the action has only one actionHandler
				handlers[0](context)

			} else {
				// if no handler by the user
				panic("Dnet: No action handler passed")
			}
		}
	}

	// if the action matched nothing ..... return the 404 code to the client
	if !isMatch {
		context.conn.WriteJSON(Response{IncomingAction, 404, "Action Not Found", ""})
	}

}

// LastSeen is called when the connection gets closed.
// It's very useful for setting the last seen or apperance of the user
func (r *MainRouter) lastSeen(c *Context) {
	// check to see if there are more than context in the hub
	contextCount := 0

	for context := range c.hub.contexts {
		if context.ID == c.ID {
			contextCount++
		}
	}

	//if there are no any other contexts online ...then call the last seeen handler
	if contextCount == 1 {
		//check if lastSeen hanlder is set
		if r.isLastSeenHandlerSet {
			r.lastSeenHandler(c)
		}
	}
}

// LastSeen is called when the authenticated user gets offline or logs out
// It's very useful for setting the last seen of the user connection
func LastSeen(handler ActionHandler) {
	//  the lastSeenHandler is set
	Router1.isLastSeenHandlerSet = true

	// set it
	Router1.lastSeenHandler = handler
}

// ticketCleaner is for cleaning or removing expired tickets from the router Tickets
func (r *MainRouter) ticketCleaner() {
	ticker := time.NewTicker(r.ticketAge)

	for {
		select {
		case <-ticker.C:

			// enter in the tickets and remove the expired tickets
			for i, ticket := range r.tickets {
				_, _, expireTimeString, _ := ticketParts(ticket)
				expireTime, err := time.Parse(time.RFC3339, expireTimeString)
				if err != nil {
					panic(err)
				}

				// remove the ticket if the has expired
				if time.Now().Local().After(expireTime) {
					r.tickets = append(r.tickets[:i], r.tickets[i+i:]...)
				}
			}
		}
	}
}

// RouterMatcher is for grouping actions, and creating middleware subjected to the particular group
type RouterMatcher struct {
	// path is the path to match
	path string
}

// On method is adding Event handlders to the router by prefixing it with the Matcher path
func (m RouterMatcher) On(action string, handlers ...ActionHandler) {
	Router1.actionHandlers[m.path+action] = handlers
}

//Use is for adding action route-wise middlewares
func (m *RouterMatcher) Use(handlers ...ActionHandler) {
	// append the given action handlers for matching the actioon-path beginning
	Router1.routeMatchers[m.path] = append(Router1.routeMatchers[m.path], handlers...)
}

// Router1 is websocket main router
var Router1 *MainRouter = &MainRouter{
	actionHandlers: make(map[string][]ActionHandler), routeMatchers: make(map[string][]ActionHandler),
	ticketSecrete: tzcrypt.RandString(32), ticketIV: tzcrypt.RandBytes(16),
	ticketAge: 30 * time.Second,
	maxSize:   512,
}
