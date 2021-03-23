package tz

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dilungasr/tanzanite/types"
)

// Bind is for binding client data into the struct, map or interface{}
func Bind(r *http.Request, v interface{}) error {
	// check if pointer or not
	if !types.Is("ptr", v) {
		return errors.New("TZ: Only pointer arguments are allowed in the Bind()")
	}

	// decode the data only to the struct, map or interface{}
	switch {
	case types.Is("struct", v) || types.Is("map", v) || types.Is("all", v):
		json.NewDecoder(r.Body).Decode(v)
	default:
		return errors.New("TZ: You can only bind data to the struct, map or interface{}")
	}

	return nil
}

// Binder does the same as Bind() but handles panic the error, hence useful if you don't want to check error and modify it but just panic it
func Binder(r *http.Request, v interface{}) {
	// check if pointer or not
	if !types.Is("ptr", v) {
		panic("TZ: Only pointer arguments are allowed in the Bind()")
	}

	// decode the data only to the struct, map or interface{}
	switch {
	case types.Is("struct", v) || types.Is("map", v) || types.Is("all", v):
		json.NewDecoder(r.Body).Decode(v)
	default:
		panic("TZ: You can only bind data to the struct, map or interface{}")
	}
}
