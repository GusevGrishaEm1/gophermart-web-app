package http

import (
	"encoding/json"
	"errors"
	"net/http"

	customerr "github.com/GusevGrishaEm1/gophermart-web-app.git/internal/app/error"
)

func sendServerErr(err error, w http.ResponseWriter) {
	customErr := &customerr.CustomError{}
	if errors.As(err, &customErr) {
		w.WriteHeader(customErr.HTTPStatus)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}

func sendClientErr(err error, w http.ResponseWriter) {
	customErr := &customerr.CustomError{}
	if errors.As(err, &customErr) {
		w.WriteHeader(customErr.HTTPStatus)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func sendOKWithBody(w http.ResponseWriter, responseBody any) {
	data, err := json.Marshal(responseBody)
	if err != nil {
		sendServerErr(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func sendOKWithCookie(token string, w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:  string("USER_ID"),
		Value: token,
	}
	w.Header().Set("Content-Type", "application/json")
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}
