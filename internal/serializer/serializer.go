package serializer

import (
	"encoding/json"
	"fmt"

	"github.com/zklevsha/go-musthave-diploma/internal/archive"
	"github.com/zklevsha/go-musthave-diploma/internal/structs"
)

func EncodeServerResponse(resp structs.Response, compress bool, asText bool) ([]byte, error) {

	var msg []byte
	var err error

	if asText {
		msg = []byte(resp.AsText())
	} else {
		msg, err = json.Marshal(resp)
		if err != nil {
			return nil, fmt.Errorf("failed to encode server response to json %s", err.Error())
		}
	}

	if !compress {
		return msg, nil
	}

	compressed, err := archive.Compress(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to compress server response %s", err.Error())
	}
	return compressed, nil
}

func EncodeOrdersResponse(orders []structs.Order, compress bool) ([]byte, error) {
	resp, err := json.Marshal(orders)
	if err != nil {
		return nil, fmt.Errorf("failed to encode server response to json %s", err.Error())
	}

	if !compress {
		return resp, nil
	}

	compressed, err := archive.Compress(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to compress server response %s", err.Error())
	}
	return compressed, nil
}
