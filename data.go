package tz

// Map is for simplifying sending of json data
type Map map[string]interface{}

// Msg models the msg to a simple msg:string data
func Msg(msg string) (simpleMsgResponse Map) {
	return Map{"msg": msg}
}
