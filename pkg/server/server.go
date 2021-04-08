// ------------------------------------------------------ {COPYRIGHT-TOP} ---
// Licensed Materials - Property of IBM
// 5900-AEO
//
// Copyright IBM Corp. 2020, 2021. All Rights Reserved.
//
// US Government Users Restricted Rights - Use, duplication, or
// disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
// ------------------------------------------------------ {COPYRIGHT-END} ---
package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

const (
	Port = 8080
)

func Start() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// DumpRequest will put the request information into readable plain-text
		reqBytes, err := httputil.DumpRequest(req, true)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 error
			fmt.Fprintln(w, "Unable to process request")
			fmt.Println(err)
			return
		}

		// Echo the request back, and log it.
		fmt.Fprintln(w, string(reqBytes))
		fmt.Println(string(reqBytes))
	})

	fmt.Println("Now listening...")
	http.ListenAndServe(fmt.Sprintf(":%d", Port), nil)
}
