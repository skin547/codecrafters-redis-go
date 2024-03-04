package util

import "strconv"

// ToRespSimpleStrings converts a simple string response to the RESP format
func ToRespSimpleStrings(value string) string {
	return "+" + value + "\r\n"
}

// ToRespBulkStrings converts a bulk string response to the RESP format
func ToRespBulkStrings(value string) string {
	return "$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n"
}

// ToRespErrorBulkStrings generates an error response in the RESP format
func ToRespErrorBulkStrings(value string) string {
	return "-" + value + "\r\n"
}
