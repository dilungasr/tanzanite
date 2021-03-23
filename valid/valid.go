package valid

import (
	"net/http"
	"reflect"
	"strings"

	tz "github.com/dilungasr/tanzanite"
)

// Errors is the main errors container for tanzanite's validator
var Errors = make(map[string][]string, 1)

// Struct is the validation invoker function
func Struct(val interface{}) (map[string][]string, bool) {
	// empty the Errors first
	Errors = map[string][]string{}
	// check if it's not struct
	if !tz.Is("struct", val) {
		panic("TZ: Non-struct argument passed in Struct()")
	}

	t1 := reflect.TypeOf(val)
	v1 := reflect.ValueOf(val)
	// loop to all fields for validation
	for i := 0; i < t1.NumField(); i++ {
		// if the field is to be validated
		if v, ok := t1.Field(i).Tag.Lookup("vdt"); ok {
			fullField := map[string]interface{}{}

			// use json name if present, or use the original value
			if v, ok := t1.Field(i).Tag.Lookup("json"); ok {
				fullField["json"] = v
			} else {
				fullField["json"] = t1.Field(i).Name
			}

			// assign the value
			fullField["value"] = v1.Field(i).Interface()

			// pass fullField and the tags in the workWithWithRules
			workWithRules(fullField, v, val)
		}
	}

	//check the if there a errors
	if len(Errors) > 0 {
		return Errors, false
	}

	return nil, true
}

// for working with all rules defined as vdt tag values
func workWithRules(field map[string]interface{}, rules string, val interface{}) {
	//  split the string
	rulesSlice := strings.Split(rules, ",")

	// iterate to find all the rules
	for _, v := range rulesSlice {
		rule := strings.TrimSpace(v)
		// for key...value rules
		ruleParts1 := strings.Split(rule, "=")
		ruleParts2 := strings.Split(rule, "-")

		switch {
		case rule == "req":
			required(field)
		case rule == "email":
			isEmaill(field)
		case rule == "phone":
			isPhone(field)
		case rule == "password":
			stdPass(field, 8, 34)
		case rule == "city":
			city(field)
		case rule == "country":
			city(field)
		case rule == "letterOnly":
			letterOnly(field)
		case rule == "letterSpace":
			letterAndSpaceOnly(field)
		case ruleParts1[0] == "eq":
			equalField(field, rule, val)
		case strings.TrimSpace(ruleParts2[0]) == "igt" ||
			strings.TrimSpace(ruleParts2[0]) == "ilt" ||
			strings.TrimSpace(ruleParts2[0]) == "fgt" ||
			strings.TrimSpace(ruleParts2[0]) == "flt":
			gtLt(field, ruleParts2)

		}
	}
}

//Auto checks and automatically sends errors to the client
// return true if performing auto request sending
func Auto(w http.ResponseWriter, v interface{}) bool {
	if err, ok := Struct(v); !ok {
		tz.Send(w, 422, err)
		return true
	}

	return false
}
