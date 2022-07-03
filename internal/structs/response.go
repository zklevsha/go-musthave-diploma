package structs

import "fmt"

type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (r Response) AsText() string {
	var msg string
	if r.Message != "" {
		msg = fmt.Sprintf("meassage:%s;", r.Message)
	}
	if r.Error != "" {
		msg += fmt.Sprintf("error:%s;", r.Error)
	}
	return msg
}
