package conv

import "github.com/pkg/errors"

func DataToString(data any) ([]byte, error) {
	switch v := data.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return nil, errors.WithStack(ErrDataIsNotStringable)
	}
}
