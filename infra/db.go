package infra

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

var cachedFieldMap sync.Map

func setDBMapper(db *sqlx.DB) {
	db.MapperFunc(func(s string) string { //snake_case
		v, ok := cachedFieldMap.Load(s)
		if ok {
			return v.(string)
		}
		var ret string
		for i, r := range s {
			if r >= 'A' && r <= 'Z' && i > 0 {
				ret += "_"
			}
			ret += strings.ToLower(string(r))
		}
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
