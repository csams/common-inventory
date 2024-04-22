package controllers

import "net/http"

func Ready(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
