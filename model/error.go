package model

import "fmt"

type UziErr struct {
	Err     string      `json:"err"`
	Message interface{} `json:"message"`
	Code    int         `json:"code"`
}

func (u UziErr) Error() string {
	return fmt.Sprintf("%v: %v", u.Message, u.Err)
}
