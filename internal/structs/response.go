package structs

import "fmt"

type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (s *Response) AsText() string {
	var msg string
	if s.Message != "" {
		msg = fmt.Sprintf("meassage:%s;", s.Message)
	}
	if s.Error != "" {
		msg += fmt.Sprintf("error:%s;", s.Error)
	}
	return msg
}
