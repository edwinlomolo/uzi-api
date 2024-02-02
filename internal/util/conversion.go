package util

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
)

func FloatToString(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func Base64Key(key interface{}) string {
	keyString, err := json.Marshal(key)
	if err != nil {
		panic(err)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(keyString))

	return encoded
}
