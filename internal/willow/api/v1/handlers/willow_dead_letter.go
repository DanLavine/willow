package handlers

import "net/http"

func (qh *queueHandler) DeadLetterList(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (qh *queueHandler) DeadLetterGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
