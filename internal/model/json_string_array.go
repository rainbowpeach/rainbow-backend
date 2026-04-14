package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONStringArray []string

func (a JSONStringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}

	data, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("marshal json string array: %w", err)
	}

	return string(data), nil
}

func (a *JSONStringArray) Scan(value any) error {
	if value == nil {
		*a = JSONStringArray{}
		return nil
	}

	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("unsupported JSONStringArray scan type %T", value)
	}

	if len(raw) == 0 {
		*a = JSONStringArray{}
		return nil
	}

	var result []string
	if err := json.Unmarshal(raw, &result); err != nil {
		return fmt.Errorf("unmarshal json string array: %w", err)
	}

	*a = JSONStringArray(result)
	return nil
}
