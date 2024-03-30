package models

type Music struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	Picture string `json:"picture"`
	Track   string `json:"track"`
}
