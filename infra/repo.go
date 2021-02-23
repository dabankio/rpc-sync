package infra

import (
	"time"

	"github.com/jmoiron/sqlx"
)

func NewRepo(db *sqlx.DB) *Repo { return &Repo{db: db} }

type Repo struct {
	db *sqlx.DB
}

type ApiLog struct {
	ID     uint64
	Time   time.Time
	URL    string
	Method string
	IP     string
	Body   string
}

func (r *Repo) CreateApiLog(log ApiLog) error {
	_, err := r.db.Exec(`insert into api_log (time, url, method, ip, body) 
	values ($1, $2, $3, $4, $5)`, log.Time, log.URL, log.Method, log.IP, log.Body)
	return err
}
