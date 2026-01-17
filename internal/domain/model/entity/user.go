package entity

type UserEntity struct {
	ID              string
	Name            string
	DisplayName     string
	Avatar          string
	RefferalChannel UserRefferalChannel
	Department      string
}

type UserRefferalChannel struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}
