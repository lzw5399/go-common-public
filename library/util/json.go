package util

import "encoding/json"

func InterfaceToJsonString(obj interface{}) string {
	jsonBytes, err := json.Marshal(obj)
	if err == nil {
		return string(jsonBytes)
	}

	return ""
}

func InterfaceToJsonByte(obj interface{}) []byte {
	jsonBytes, err := json.Marshal(obj)
	if err == nil {
		return jsonBytes
	}

	return []byte{}
}
