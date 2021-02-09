// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package main

import (
	"bbcsyncer/infra"
	"bbcsyncer/pow"
	"bbcsyncer/reward"
	"bbcsyncer/sync"
)

// Injectors from wire.go:

func InitializeApp() (*App, error) {
	conf := infra.ParseConf()
	db, err := infra.NewPGDB(conf)
	if err != nil {
		return nil, err
	}
	repo := sync.NewRepo(db)
	client, err := infra.NewBBCClient(conf)
	if err != nil {
		return nil, err
	}
	worker := sync.NewWorker(repo, client)
	rewardRepo := reward.NewRepo(db)
	calc := reward.NewCalc(rewardRepo, repo)
	v := NewJobs(worker, calc)
	sched, err := infra.NewSched(v)
	if err != nil {
		return nil, err
	}
	handler := reward.NewHandler(rewardRepo)
	powRepo := pow.NewRepo(db)
	powHandler := pow.NewHandler(powRepo)
	mux := NewRouter(handler, powHandler)
	app := NewApp(sched, mux)
	return app, nil
}
