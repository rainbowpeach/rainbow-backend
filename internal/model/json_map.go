package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONMap map[string]string

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}

	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshal json map: %w", err)
	}

	return string(data), nil
}

func (m *JSONMap) Scan(value any) error {
	if value == nil {
		*m = JSONMap{}
		return nil
	}

	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("unsupported JSONMap scan type %T", value)
	}

	if len(raw) == 0 {
		*m = JSONMap{}
		return nil
	}

	var result map[string]string
	if err := json.Unmarshal(raw, &result); err != nil {
		return fmt.Errorf("unmarshal json map: %w", err)
	}
	if result == nil {
		result = JSONMap{}
	}

	*m = result
	return nil
}
