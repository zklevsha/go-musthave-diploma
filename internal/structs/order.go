package structs

import "fmt"

type Order struct {
	Number     string `json:"number,omitempty"`
	Order      string `json:"order,omitempty"`
	Status     string `json:"status"`
	Accrual    *int   `json:"accrual,omitempty"`
	UploadedAt string `json:"uploaded_at,omitempty"`
}

func (o Order) AsText() string {
	var msg string
	if o.Number != "" {
		msg = fmt.Sprintf("number:%s;", o.Number)
	}
	if o.Order != "" {
		msg += fmt.Sprintf("order:%s;", o.Order)
	}
	msg += fmt.Sprintf("status:%s;", o.Status)
	if o.Accrual != nil {
		msg += fmt.Sprintf("accrual:%d;", o.Accrual)
	}
	if o.UploadedAt != "" {
		msg += fmt.Sprintf("uploaded_at:%s;", o.UploadedAt)
	}
	return msg
}
