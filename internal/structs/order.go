package structs

type Order struct {
	Number     string `json:"number,omitempty"`
	Order      string `json:"order,omitempty"`
	Status     string `json:"status"`
	Accrual    *int   `json:"accrual,omitempty"`
	UploadedAt string `json:"uploaded_at,omitempty"`
}
