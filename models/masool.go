package models

type MasoolData struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type Masool struct {
	Id   int64        `json:"id"`
	Name string       `json:"name"`
	Data []MasoolData `json:"data"` // Stores all fields as key-value pairs
}
