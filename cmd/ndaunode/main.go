package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
	"github.com/oneiro-ndev/ndau/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/server"
	tmlog "github.com/tendermint/tendermint/libs/log"
)

var useNh = flag.Bool("use-ndauhome", false, "if set, keep database within $NDAUHOME/ndau")
var dbspec = flag.String("spec", "", "manually set the noms db spec")
var indexAddr = flag.String("index", "", "search index address")
var socketAddr = flag.String("addr", "0.0.0.0:26658", "socket address for incoming connection from tendermint")
var echoSpec = flag.Bool("echo-spec", false, "if set, echo the DB spec used and then quit")
var echoEmptyHash = flag.Bool("echo-empty-hash", false, "if set, echo the hash of the empty DB and then quit")
var echoHash = flag.Bool("echo-hash", false, "if set, echo the current DB hash and then quit")
var echoVersion = flag.Bool("version", false, "if set, echo the current version and exit")
var genesisfilePath = flag.String("genesisfile", "", "if set, update system variables from the genesisfle and exit")
var asscfilePath = flag.String("asscfile", "", "if set, create special accounts from the given associated data file and exit")

// Bump this any time we need to reset and reindex the ndau chain.  For example, if we change the
// format of something in the index, say, needing to use unsorted sets instead of sorted sets; if
// our new searching code doesn't expect the old format in the index, we can bump this to cause a
// wipe and full reindex of the blockchain using the new format that the new search code expects.
// That is why this is tied to code here, rather than a variable we pass in.
// History:
//   0 = initial version
//   1 = new format for indxing transaction fee/sib
const indexVersion = 1

func getNdauhome() string {
	nh := os.ExpandEnv("$NDAUHOME")
	if len(nh) > 0 {
		return nh
	}
	return filepath.Join(os.ExpandEnv("$HOME"), ".ndau")
}

func getNdauConfigDir() string {
	return filepath.Join(getNdauhome(), "ndau")
}

func getDbSpec() string {
	if len(*dbspec) > 0 {
		return *dbspec
	}
	if *useNh {
		return filepath.Join(getNdauConfigDir(), "noms")
	}
	// default to noms server for dockerization
	return "http://noms:8000"
}

func getIndexAddr() string {
	if len(*indexAddr) > 0 {
		return *indexAddr
	}
	if *useNh {
		return filepath.Join(getNdauConfigDir(), "redis")
	}
	// default to redis server for dockerization
	return "redis:6379"
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	if *echoSpec {
		fmt.Println(getDbSpec())
		os.Exit(0)
	}

	if *echoEmptyHash {
		fmt.Println(getEmptyHash())
		os.Exit(0)
	}

	if *echoVersion {
		version.Emit()
	}

	ndauhome := getNdauhome()
	configPath := config.DefaultConfigPath(ndauhome)

	conf, err := config.LoadDefault(configPath)
	check(err)

	if *echoHash {
		fmt.Println(getHash(conf))
		os.Exit(0)
	}

	if len(*asscfilePath) > 0 || len(*genesisfilePath) > 0 {
		updateFromGenesis(*genesisfilePath, *asscfilePath, conf)
		os.Exit(0)
	}

	app, err := ndau.NewApp(getDbSpec(), getIndexAddr(), indexVersion, *conf)
	check(err)

	app.LogState()

	server := server.NewSocketServer(*socketAddr, app)

	// Don't let the server log anything (it's very noisy) unless we're at debug level.
	if app.GetLogger().(*logrus.Logger).Level == logrus.DebugLevel {
		var tmLogger tmlog.Logger
		switch os.Getenv("LOG_FORMAT") {
		case "json", "":
			tmLogger = tmlog.NewTMJSONLogger(os.Stdout)
		case "text", "plain":
			tmLogger = tmlog.NewTMLogger(os.Stdout)
		default:
			tmLogger = tmlog.NewTMJSONLogger(os.Stdout)
		}
		server.SetLogger(tmLogger)
	}

	err = server.Start()
	check(err)

	entry := app.GetLogger().WithFields(logrus.Fields{
		"address": *socketAddr,
		"name":    server.String(),
	})

	v, err := version.Get()
	if err == nil {
		entry = entry.WithField("version", v)
	} else {
		entry = entry.WithError(err)
	}
	entry.Info("started ABCI socket server")

	// This gives us a mechanism to kill off the server with an OS signal (for example, Ctrl-C)
	app.App.WatchSignals()

	// This runs forever until a signal happens
	<-server.Quit()
}
