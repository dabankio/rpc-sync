package infra

import (
	"encoding/json"
	"log"
	"net/http"
)

type BaseHandler struct{}

func (_ BaseHandler) WriteErr(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	_, we := w.Write([]byte(err.Error()))
	if we != nil {
		log.Println("[err] write json response: ", we)
	}
}

func (_ BaseHandler) WriteJSON(w http.ResponseWriter, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	_, err = w.Write(b)
	if err != nil {
		log.Println("[err] write json response: ", err)
	}
}
