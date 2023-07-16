package codegen

import (
	"encoding/json"
	"fmt"

	"github.tesla.cn/itapp/lines/errorx"
)

const (
	extPropGoType = "x-go-type"
)

func extTypeName(extPropValue interface{}) (string, error) {
	raw, ok := extPropValue.(json.RawMessage)
	if !ok {
		return "", fmt.Errorf("failed to convert type: %T", extPropValue)
	}
	var name string
	if err := json.Unmarshal(raw, &name); err != nil {
		return "", errorx.Wrap(err, "failed to unmarshal json")
	}

	return name, nil
}
