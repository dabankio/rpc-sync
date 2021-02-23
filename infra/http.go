package infra

import (
	"encoding/json"
	"io"
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

// return success? (auto write json err when failed)
func (h BaseHandler) BindBody(w http.ResponseWriter, req *http.Request, v interface{}) bool {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		h.WriteErr(w, err)
		return false
	}

	err = json.Unmarshal(b, v)
	if err != nil {
		h.WriteErr(w, err)
		return false
	}
	return true
}

func (_ BaseHandler) WriteJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Add("Content-Type", "application/json")
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

func (_ BaseHandler) RealIP(r *http.Request) string {
	for _, ipHeader := range []string{
		"x-real-ip",
		"x-forwarded-for",
		"x_forwarded_for",
		"x-forwared-for",
	} {
		if h := r.Header.Get(ipHeader); h != "" {
			return h
		}
	}
	return r.RemoteAddr
}

// 旧有.net系统数据结构
// eg: {"Data":1,"Success":true,"Code":1,"Message":"OK"}
type LegacyResult struct {
	Success bool
	Code    int
	Message string
	Data    interface{}
}

const LegacyCodeSuccess = 1
