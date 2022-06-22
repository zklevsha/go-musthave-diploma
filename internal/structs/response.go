package structs

import "fmt"

type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Token   string `json:"token,omitempty"`
}

func (s *Response) AsText() string {
	var msg string
	if s.Message != "" {
		msg = fmt.Sprintf("meassage:%s;", s.Message)
	}
	if s.Error != "" {
		msg += fmt.Sprintf("error:%s;", s.Error)
	}
	if s.Token != "" {
		msg += fmt.Sprintf("token:%s;", s.Token)
	}
	return msg
}
