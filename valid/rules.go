package valid

import (
	"net"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// stdPass is for validating app-standard passwords
func stdPass(field map[string]interface{}, min, max int) {
	fieldName := field["json"].(string)
	value := field["value"].(string)
	errs := []string{}

	var (
		isMin   bool
		special bool
		number  bool
		upper   bool
		lower   bool
	)
	// append error
	appendError := func(err string) {
		errs = append(errs, err)
	}

	//test for the muximum and minimum characters required for the password string
	if len(value) < min || len(value) > max {
		isMin = false
		appendError("length should be " + strconv.Itoa(min) + " to " + strconv.Itoa(max))
	}

	for _, c := range value {
		// Optimize the code if all become true before reaching the end
		if special && number && upper && lower && isMin {
			break
		}

		// else go switching
		switch {
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsLower(c):
			lower = true
		case unicode.IsNumber(c):
			number = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		}
	}

	// Add custom error messages
	if !special {
		appendError("should contain at least a single special character")
	}
	if !number {
		appendError("should contain at least a single digit")
	}
	if !lower {
		appendError("should contain at least a single lowercase letter")
	}
	if !upper {
		appendError("should contain at least single uppercase letter")
	}

	// if there is any error
	if len(errs) > 0 {
		Errors[fieldName] = append(Errors[fieldName], errs...)
	}
}

//email for validatin an email
func isEmaill(field map[string]interface{}) {
	value := field["value"].(string)
	fieldName := field["json"].(string)
	regex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	isValid := true
	// basic rules
	if len(value) < 3 || len(value) > 254 {
		isValid = false

	} else if !regex.MatchString(value) {
		isValid = false
	}

	// check the DNS MX record to see if the domain is  a valid mailserver
	emailParts := strings.Split(value, "@")
	if len(emailParts) != 2 {
		isValid = false
	} else if mx, err := net.LookupMX(emailParts[1]); err != nil || len(mx) == 0 {
		isValid = false
	}
	// valid or not

	if !isValid {
		msg := "Not valid email"
		Errors[fieldName] = append(Errors[fieldName], msg)
	}
}

// // isINt test if the value is int
// func isInt(field map[string]interface{}) {
// 	fieldName := field["json"].(string)

// 	// check if int or not
// 	if value, ok := field["value"].(int); !ok || value == 0 {
// 		msg := "Not integer"
// 		Errors[fieldName] = append(Errors[fieldName], msg)
// 	}
// }

// // isBool test if the value is int
// func isBool(field map[string]interface{}) {
// 	fieldName := field["json"].(string)

// 	// check if int or not
// 	if _, ok := field["value"].(bool); !ok {
// 		msg := "Not Boolean"
// 		Errors[fieldName] = append(Errors[fieldName], msg)
// 	}
// }

//isPhone check if the value is a phone number or not
func isPhone(field map[string]interface{}) {
	test(field, "Not valid phone number", `^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){0,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
}

// letterOnly
func letterOnly(field map[string]interface{}) {
	test(field, "should contain only letters", `^[A-Za-z]+$`)
}

// letterOnly
func letterAndSpaceOnly(field map[string]interface{}) {
	test(field, "should contain only contain letters and spaces if needed", `^([A-Za-z]+)(\s[A-Za-z]+)*$`)
}

// regexy...
func test(field map[string]interface{}, msg string, regexpr string) {
	fieldName := field["json"].(string)
	reg := regexp.MustCompile(regexpr)
	isError := false
	isSlice := false
	value, ok := field["value"].(string)

	if !ok {
		// check if it's a slice of strings
		val, ok := field["value"].([]string)
		if ok {
			isSlice = true
			// if lenght is zero .. the validation should automatically fail
			if len(val) == 0 {
				isError = true
			} else {

				for _, v := range val {
					if !reg.MatchString(v) {
						isError = true
						break
					}
				}
			}
		}
	}

	// is not valid
	if isError {
		Errors[fieldName] = append(Errors[fieldName], "Array elements "+msg)
	} else if !reg.MatchString(value) && !isSlice {
		Errors[fieldName] = append(Errors[fieldName], msg)
	}
}

// city for validating the city
func city(field map[string]interface{}) {
	test(field, "Not a valid city or country name", `^\p{L}+(?:([\ \-\']|(\.\ ))\p{L}+)*$`)
}

// required
func required(field map[string]interface{}) {
	value := field["value"]
	fieldName := field["json"].(string)

	if value == "" {
		Errors[fieldName] = append(Errors[fieldName], "This field cannot be blank")
	}
}

// gtLt for working with number ranges
func gtLt(field map[string]interface{}, ruleParts []string) {
	fieldName := field["json"].(string)

	ruleName := strings.TrimSpace(ruleParts[0])

	// react according to the rules
	if ruleName == "flt" || ruleName == "fgt" {
		fieldValue, ok := field["value"].(float64)
		if !ok {
			panic("TZ-VALID: The field value of '" + fieldName + "' should be  float64")
		}

		//convert the given ruleNumber to to float64
		ruleValue, err := strconv.ParseFloat(strings.TrimSpace(ruleParts[1]), 64)
		if err != nil {
			panic("TZ-VALID: '" + ruleName + "' in field '" + fieldName + "' should contain should contain a floating point point number eg. gt-10.7")
		}
		// work with the rules
		if ruleName == "flt" && fieldValue >= ruleValue {
			Errors[fieldName] = append(Errors[fieldName], "should be a number with decimal points and it's less than "+strconv.FormatFloat(ruleValue, 'f', 2, 64))

		} else if ruleName == "fgt" && fieldValue <= ruleValue {
			Errors[fieldName] = append(Errors[fieldName], "should be decimal point number which is greater than "+strconv.FormatFloat(ruleValue, 'f', 2, 64))
		}
	} else if ruleName == "ilt" || ruleName == "igt" {
		fieldValue, _ := field["value"].(int)
		// convert the given ruleNumber to int
		var ruleValue int
		if strings.TrimSpace(ruleParts[1]) == "yn" {
			ruleValue = time.Now().Local().Year()
		} else {
			// try to convert the ruleValue
			ruleIntNumber, err := strconv.Atoi(strings.TrimSpace(ruleParts[1]))
			if err != nil {
				panic("TZ-VALID: The field value of " + fieldName + " should be  int")
			}

			ruleValue = ruleIntNumber
		}

		// work with the rules
		// for less than
		if ruleName == "ilt" && fieldValue >= ruleValue {
			Errors[fieldName] = append(Errors[fieldName], "should be an integer less than "+strconv.Itoa(ruleValue))
		} else if ruleName == "igt" && fieldValue <= ruleValue {
			Errors[fieldName] = append(Errors[fieldName], "should be an integer greater than "+strconv.Itoa(ruleValue))
		}

	}
}

// equalField to make sure two fields are the similar
func equalField(field map[string]interface{}, rule string, val interface{}) {
	name1 := field["json"].(string)
	name2 := strings.Split(rule, "=")[1]
	var value1 interface{}
	var value2 interface{}

	t1 := reflect.TypeOf(val)
	v1 := reflect.ValueOf(val)

	// loop to find the fields
	for i := 0; i < t1.NumField(); i++ {
		if v, ok := t1.Field(i).Tag.Lookup("json"); ok {
			if v == name1 {
				value1 = v1.Field(i).Interface()
			} else if v == name2 {
				value2 = v1.Field(i).Interface()
			}
		}
	}

	// check if the value are the same
	if value1 != value2 {
		Errors[name1] = append(Errors[name1], name1+" don't match "+name2)
	}

}
