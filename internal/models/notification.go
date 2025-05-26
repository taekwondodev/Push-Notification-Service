package models

type Notification struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Message string `json:"message"`
}
