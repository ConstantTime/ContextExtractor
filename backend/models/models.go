package models

type Ticket struct {
    ID         int    `json:"id"`
    CustomerID int    `json:"customer_id"`
    Subject    string `json:"subject"`
    Content    string `json:"content"`
    Context    string `json:"context"`
	Status 	   string `json:"status"`
}

type ContextRule struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Keywords string `json:"keywords"`
}