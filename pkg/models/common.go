package models

import (
	"database/sql"
	"encoding/json"
)

func NullString(str string) sql.NullString {
	return sql.NullString{
		String: str,
		Valid:  str != "",
	}
}

func NullStringExport(nullString sql.NullString) string {
	if nullString.Valid {
		return nullString.String
	}
	return ""
}

func ArrayToArrayJSON[E any](array []E) JSONArray {
	init := make(JSONArray, len(array))
	for i := range array {
		_bytes, err := json.Marshal(array[i])
		if err != nil {
			continue
		}

		init[i] = string(_bytes)
	}

	return init
}

func StructToJSONB[E any](val E) JSONB {
	_bytes, err := json.Marshal(val)
	if err != nil {
		return nil
	}

	var jsb JSONB
	err = json.Unmarshal(_bytes, &jsb)
	if err != nil {
		return nil
	}

	return jsb
}
