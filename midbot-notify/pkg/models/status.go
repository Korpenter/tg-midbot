package models

type Status struct {
	UID            string         `json:"uid"`
	ReceptionDate  string         `json:"receptionDate"`
	PassportStatus PassportStatus `json:"passportStatus"`
	InternalStatus InternalStatus `json:"internalStatus"`
}

type InternalStatus struct {
	Name    string `json:"name"`
	Percent int    `json:"percent"`
}

type PassportStatus struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Color        string `json:"color"`
	Subscription bool   `json:"subscription"`
}
