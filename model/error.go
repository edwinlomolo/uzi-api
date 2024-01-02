package model

import "fmt"

type UziErr struct {
	Error   error       `json:"error"`
	Message interface{} `json:"message"`
	Code    int         `json:"code"`
}

func (u UziErr) ErrorString() string {
	return fmt.Sprintf("%s: %s", u.Message, u.Error.Error())
}
