// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package sync

import (
	_ "github.com/lib/pq"
)

// Injectors from wire.go:

func InitializeWorker() (*Worker, error) {
	conf := ParseConf()
	db, err := NewPGDB(conf)
	if err != nil {
		return nil, err
	}
	repo := NewRepo(db)
	client, err := NewBBCClient(conf)
	if err != nil {
		return nil, err
	}
	worker := NewWorker(repo, client)
	return worker, nil
}
