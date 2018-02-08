package main

import "log"
import "net/http"

func RunAPIServer(manager *Manager) error {
	http.Handle("/", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("[recover]", err)
			}
		}()
		NewMessenger(manager).ServeHTTP(res, req)
	}))
	return http.ListenAndServe(*HTTP_ADDR, http.DefaultServeMux)
}
