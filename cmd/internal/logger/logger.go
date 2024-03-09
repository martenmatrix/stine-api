package logger

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func LogResponse(response *http.Response) {
	resDump, err := httputil.DumpResponse(response, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("RESPONSE:\n%s\n", string(resDump))
}

func LogRequest(request *http.Request) {
	reqDump, err := httputil.DumpRequest(request, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("REQUEST:\n%s\n", string(reqDump))
}
