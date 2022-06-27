package structs

import "fmt"

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (c Credentials) AsText() string {
	return fmt.Sprintf("login:%s;password:%s", c.Login, c.Password)
}
