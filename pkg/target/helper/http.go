package helper

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
)

// HandleHTTPResponse Logs Error or Success messages
func HandleHTTPResponse(target string, resp *http.Response, err error) {
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
