package responser

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Errorf(w http.ResponseWriter, code int, format string, args ...interface{}) {
	Error(w, code, fmt.Sprintf(format, args...))
}

func Error(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"code": %d, "msg":"%s", "error": true}`, code, message)
}

func ErrorMap(w http.ResponseWriter, code int, message string, values map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json, _ := json.Marshal(values)
	fmt.Fprintf(w, `{"code": %d, "msg":"%s", "error": true, "data":%s}`, code, message, json)
}

func ErrorJson(w http.ResponseWriter, code int, message string, json []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"code": %d, "msg":"%s", "error": true, "data":%s}`, code, message, json)
}

func ErrorJsonParser(w http.ResponseWriter, code int, message string, toParse interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json, err := json.Marshal(toParse)
	if err != nil {
		Errorf(w, http.StatusInternalServerError, "error parsing the json: %v", err)
		return
	}
	fmt.Fprintf(w, `{"code": %d, "msg":"%s", "error": true, "data":%s}`, code, message, json)
}

func Successf(w http.ResponseWriter, code int, format string, args ...interface{}) {
	Success(w, code, fmt.Sprintf(format, args...))
}

func Success(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"code": %d, "msg":"%s", "error": false}`, code, message)
}

func SuccessMap(w http.ResponseWriter, code int, message string, values map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json, _ := json.Marshal(values)
	fmt.Fprintf(w, `{"code": %d, "msg":"%s", "error": false, "data":%s}`, code, message, json)
}

func SuccessJson(w http.ResponseWriter, code int, message string, json []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"code": %d, "msg":"%s", "error": false, "data":%s}`, code, message, json)
}

func SuccessJsonParser(w http.ResponseWriter, code int, message string, toParse interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	//convert to json the to parse
	json, err := json.Marshal(toParse)
	if err != nil {
		Errorf(w, http.StatusInternalServerError, "error parsing the json: %v", err)
		return
	}
	fmt.Fprintf(w, `{"code": %d, "msg":"%s", "error": false, "data":%s}`, code, message, json)
}
