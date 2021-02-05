package main

import (
	"bbcsyncer/infra"

	"github.com/dabankio/civil"
)

func main() {
	calc, err := Initialize()
	infra.PanicErr(err)
	day, err := civil.ParseDate("2021-02-04")
	infra.PanicErr(err)
	_, err = calc.CalcAtDayEast8zoneAndSave(day)
	infra.PanicErr(err)
}
