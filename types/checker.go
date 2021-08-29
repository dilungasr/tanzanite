package types

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
)

// Is reports if data is of type typ or not
func Is(typ string, data interface{}) bool {
	wg := sync.WaitGroup{}
	testChannel := make(chan bool, 1)

	// spawn a goroutine to make the checking process faster
	wg.Add(1)
	go func() {
		t := reflect.TypeOf(data)
		v := reflect.ValueOf(data)
		kt := t.Kind()
		kv := reflect.Indirect(v).Kind()

		switch strings.ToLower(typ) {
		case "struct":
			testChannel <- kt == reflect.Struct || kv == reflect.Struct
		case "map":
			testChannel <- kt == reflect.Map || kv == reflect.Map
		case "all":
			// checks if type is interface{}
			testChannel <- t == nil
		case "string":
			testChannel <- kt == reflect.String || kv == reflect.String
		case "ptr":
			testChannel <- kt == reflect.Ptr
		case "array":
			testChannel <- kt == reflect.Array || kv == reflect.Array
		case "int":
			testChannel <- kt == reflect.Int || kv == reflect.Int
		case "bool":
			testChannel <- kt == reflect.Bool || kv == reflect.Bool
		default:
			panic("TZ: Unknown test type =>" + typ)

		}

		wg.Done()
	}()

	wg.Wait()

	// return the boolean value
	return <-testChannel
}

// String coverts v to string
func String(v interface{}) string {
	jsn, err := json.Marshal(v)
	if err != nil {
		panic("TZ: Error encoding the data to JSON")
	}

	// return the resulted string from converstion
	return string(jsn)
}
