package tz

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dilungasr/tanzanite/types"
)

// SendJSON for sending error in json format and providing the error code
func SendJSON(w http.ResponseWriter, statusCodeAndData ...interface{}) {
	// the index of data and the default status code
	dataIndex := 0
	statusCode := 200
	positionData(&dataIndex, &statusCode, statusCodeAndData, "SendJSON")

	// send json to the client
	jsonSender(w, statusCode, statusCodeAndData[dataIndex])
}

// Send sends any data. Struct and map are sent as json whereas, other data are treated as plain texts.
// For sending other data as json, you might require SendJSON()
func Send(w http.ResponseWriter, statusCodeAndData ...interface{}) {
	// the index of data and the default status code
	dataIndex := 0
	statusCode := 200
	positionData(&dataIndex, &statusCode, statusCodeAndData, "Send")

	// determine the data and respond accordingly

	switch {
	// for structs and maps... should be ecoded to json
	case types.Is("struct", statusCodeAndData[dataIndex]) || types.Is("map", statusCodeAndData[dataIndex]):
		jsonSender(w, statusCode, statusCodeAndData[dataIndex])
		// for any other data types should be sent as text
	default:
		// send the data in string format
		stringSender(w, statusCode, statusCodeAndData[dataIndex])

	}
}

// SendString  is for sending data in string format to the client
func SendString(w http.ResponseWriter, statusCodeAndData ...interface{}) {
	// the index of data and the default status code
	dataIndex := 0
	statusCode := 200
	positionData(&dataIndex, &statusCode, statusCodeAndData, "SendString")

	// send the data in strin format
	stringSender(w, statusCode, statusCodeAndData[dataIndex])
}

// Redirect for redirecting the client request
func Redirect(w http.ResponseWriter, r *http.Request, statusCodeAndPath ...interface{}) {
	pathIndex := 0
	statusCode := 301
	positionData(&pathIndex, &statusCode, statusCodeAndPath, "Redirect")
	http.Redirect(w, r, fmt.Sprintf("http://%s/%s", r.Host, statusCodeAndPath[pathIndex]), statusCode)
}

// MIME sets the content mime type
func MIME(w http.ResponseWriter, mimeType string) {
	switch mimeType {
	case "json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

	case "text":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	case "webm":
		w.Header().Set("Content-Type", "video/webm")
	case "pdf":
		w.Header().Set("Content-Type", "application/pdf")
	case "html":
		w.Header().Set("Content-Type", "plain/html")
	}

	// makesure the mime type cannot be sniffed by the web browser
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// w.Header().Set("Access-Control-Allow-Credentials", "true")
}

// jsonSender for internal json sending
func jsonSender(w http.ResponseWriter, statusCode int, data interface{}) {
	// set the mime type
	MIME(w, "json")
	// write the status code
	w.WriteHeader(statusCode)

	// Send the json
	json.NewEncoder(w).Encode(data)
}

// stringSender  for internal text sending
func stringSender(w http.ResponseWriter, statusCode int, data interface{}) {
	// set the mime type
	MIME(w, "text")
	// write the  status code
	w.WriteHeader(statusCode)

	// send to the client
	io.WriteString(w, types.String(data))
}

// for checking the data passed and the position to find the specific data within the variadic parameters
func positionData(dataIndex, statusCode *int, statusCodeAndData []interface{}, funcName string) {
	// modify wher to find the data and the status code
	switch length := len(statusCodeAndData); {
	case length == 0:
		panic("TZ: No enough arguments in the " + funcName + "()")
	case length == 2:
		// check if is an int
		ensureStatusCode(statusCodeAndData)
		// change the data index and statusCode
		*dataIndex = 1
		*statusCode = statusCodeAndData[0].(int)
	case length > 2:
		panic("TZ: Too much arguments in " + funcName + "()")
	}
}

// Helps to ensure if the second parameter is a code or not
func ensureStatusCode(statusCodeAndData []interface{}) {
	// check if the code is there
	if !types.Is("int", statusCodeAndData[0]) {
		panic(
			"TZ: Arrangement in the Send() should be Send(w http.ResponseWriter, code int, data any) or you can omit the status code if you want it to be OK.",
		)
	}
}
