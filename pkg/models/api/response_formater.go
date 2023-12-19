package api

import "net/http"

func HttpResponse(r *http.Request, w http.ResponseWriter, statuscode int, obj APIResponseObject) (int, error) {
	// eveery request needs to sett this header so they know how to process api errors
	w.Header().Set("Content-Type", ContentHeaderFromRequest(r))

	switch ContentTypeFromRequest(r) {
	case ContentTypeJSON:
		w.WriteHeader(statuscode)
		if obj != nil {
			return w.Write(obj.EncodeJSON())
		}
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	return 0, nil
}
