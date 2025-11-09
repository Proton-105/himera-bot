package keyboard

import (
	"errors"
	"fmt"
	"strings"
)

const (
	CallbackDataSeparator  = ":"
	CallbackDataLimitBytes = 64
)

func EncodeCallback(unique, data string) (string, error) {
	if data == "" {
		if len(unique) > CallbackDataLimitBytes {
			return "", fmt.Errorf("callback data exceeds %d byte limit: got %d", CallbackDataLimitBytes, len(unique))
		}
		return unique, nil
	}

	payload := unique + CallbackDataSeparator + data
	if len(payload) > CallbackDataLimitBytes {
		return "", fmt.Errorf("callback data exceeds %d byte limit: got %d", CallbackDataLimitBytes, len(payload))
	}

	return payload, nil
}

func DecodeCallback(callbackData string) (unique, data string, err error) {
	if callbackData == "" {
		return "", "", errors.New("callback data is empty")
	}

	idx := strings.Index(callbackData, CallbackDataSeparator)
	if idx == -1 {
		return callbackData, "", nil
	}

	return callbackData[:idx], callbackData[idx+len(CallbackDataSeparator):], nil
}
