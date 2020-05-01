package gateway

type Rename struct {
	Action
	Field string `json:"field"`
	To    string `json:"to"`
}
