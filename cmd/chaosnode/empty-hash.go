package main

import (
	"fmt"
	"os"

	"github.com/oneiro-ndev/chaos/pkg/chaos"
)

// get the hash of an empty database
func getEmptyHash() string {
	// create an in-memory app
	app, err := chaos.NewApp("mem", "", -1, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	return app.HashStr()
}

// get the hash of the current database
func getHash() string {
	app, err := chaos.NewApp(getDbSpec(), "", -1, getConf())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can't get DBspec.  If noms is not running but it is on the local machine, consider the -use-ndauhome flag")
		return ""
	}
	return app.HashStr()
}
