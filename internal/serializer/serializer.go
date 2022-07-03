package serializer

import (
	"encoding/json"
	"fmt"

	"github.com/zklevsha/go-musthave-diploma/internal/archive"
)

func EncodeResponse(str interface{}, compress bool) ([]byte, error) {
	resp, err := json.Marshal(str)
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
