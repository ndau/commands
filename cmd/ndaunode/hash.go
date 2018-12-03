package main

import (
	"fmt"
	"os"

	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
)

// get the hash of an empty database
func getEmptyHash() string {
	// create an in-memory app
	// the ignored variable here is associated mocked data;
	// it's safe to ignore, because these mocks are immediately discarded
	app, _, err := ndau.InitMockApp()
	check(err)
	return app.HashStr()
}

// get the hash of the current database
func getHash(conf *config.Config) string {
	app, err := ndau.NewAppSilent(getDbSpec(), "", -1, *conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "If noms is not running but it is on the local machine, consider the -use-ndauhome flag")
	}
	check(err)
	return app.HashStr()
}
