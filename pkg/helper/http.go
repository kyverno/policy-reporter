package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
)

func CreateJSONRequest(target, method, host string, payload interface{}) (*http.Request, error) {
	body := new(bytes.Buffer)

	if err := json.NewEncoder(body).Encode(payload); err != nil {
		log.Printf("[ERROR] %s : %v\n", target, err.Error())
		return nil, err
	}

	req, err := http.NewRequest(method, host, body)
	if err != nil {
		log.Printf("[ERROR] %s : %v\n", target, err.Error())
		return nil, err
	}

	return req, nil
}

// ProcessHTTPResponse Logs Error or Success messages
func ProcessHTTPResponse(target string, resp *http.Response, err error) {
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		log.Printf("[ERROR] %s PUSH failed: %s\n", target, err.Error())
	} else if resp.StatusCode >= 400 {
		fmt.Printf("StatusCode: %d\n", resp.StatusCode)
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)

		log.Printf("[ERROR] %s PUSH failed [%d]: %s\n", target, resp.StatusCode, buf.String())
	} else {
		log.Printf("[INFO] %s PUSH OK\n", target)
	}
}

func SendJSONResponse(w http.ResponseWriter, list interface{}, err error) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{ "message": "%s" }`, html.EscapeString(err.Error()))

		return
	}

	if err := json.NewEncoder(w).Encode(list); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{ "message": "%s" }`, html.EscapeString(err.Error()))
	}
}
