package model

type SigninInput struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Courier   bool   `json:"courier"`
}
