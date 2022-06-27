package structs

import "fmt"

type Balance struct {
	Current   int `json:"current"`
	Withdrawn int `json:"withdrawn"`
}

func (b Balance) AsText() string {
	return fmt.Sprintf("balance:%d;withdrawn:%d", b.Current, b.Withdrawn)
}
