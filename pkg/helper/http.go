package helper

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
)

func SendJSONResponse(w http.ResponseWriter, list interface{}, err error) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{ "message": "%s" }`, html.EscapeString(err.Error()))

		return
	}

	if err := json.NewEncoder(w).Encode(list); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{ "message": "%s" }`, html.EscapeString(err.Error()))
	}
}
