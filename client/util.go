package main

import (
	"fmt"
	"strings"
)

func getErrorStringFromRecovery(r interface{}) string {
	errStr := ""
	switch t := r.(type) {
	case error:
		errStr = t.Error()
		break
	default:
		errStr = fmt.Sprintf("%#v", r)
	}
	return errStr
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func combineServerUrl(relUrl string) string {
	return strings.TrimRight(*serverFlag, "/") + "/" + strings.TrimLeft(relUrl, "/")
}