package infra

import (
	"context"
	_ "embed"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

//go:embed db.sql
var schemaSQL string

var cachedFieldMap sync.Map

func ToSnakeCase(s string) (ret string) {
	var uppers []bool
	for _, r := range s {
		uppers = append(uppers, r >= 'A' && r <= 'Z')
	}
	le := len(s)
	for i, r := range s {
		if i > 0 && uppers[i] {
			if !uppers[i-1] { //前一个是小写
				ret += "_"
			} else { //前一个是大写
				if i != le-1 && !uppers[i+1] { //后一个是小写
					ret += "_"
				}
			}
		}
		ret += strings.ToLower(string(r))
	}
	return
}

func setDBMapper(db *sqlx.DB) {
	db.MapperFunc(func(s string) string { //snake_case
		v, ok := cachedFieldMap.Load(s)
		if ok {
			return v.(string)
		}
		ret := ToSnakeCase(s)
		cachedFieldMap.Store(s, ret)
		return ret
	})
}

func NewPGDB(conf Conf) (*sqlx.DB, error) {
	log.Println("connecting db...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	db, err := sqlx.ConnectContext(ctx, "postgres", conf.DB)
	if err != nil {
		return nil, err
	}
	setDBMapper(db)
	return db, nil
}

func NewMysqlDB(conf Conf) (*sqlx.DB, error) {
	return sqlx.Connect("mysql", conf.DB)
}

func RunInTx(db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err := recover(); err != nil {
			_ = tx.Rollback()
			panic(err)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
