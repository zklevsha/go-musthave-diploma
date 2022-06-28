package structs

type Withdraw struct {
	Order string `json:"order"`
	Sum   int    `json:"sum"`
}
