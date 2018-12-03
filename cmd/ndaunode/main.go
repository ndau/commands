package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
	"github.com/oneiro-ndev/ndau/pkg/version"
	"github.com/oneiro-ndev/o11y/pkg/honeycomb"
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
var updateConfFrom = flag.String("update-conf-from", "", "if set, update app configuration from the given genesisfile and exit")
var updateChainFrom = flag.String("update-chain-from", "", "if set, update noms from the given associated data file and exit")

// Bump this any time we need to reset and reindex the ndau chain.  For example, if we change the
// format of something in the index, say, needing to use unsorted sets instead of sorted sets; if
// our new searching code doesn't expect the old format in the index, we can bump this to cause a
// wipe and full reindex of the blockchain using the new format that the new search code expects.
// That is why this is tied to code here, rather than a variable we pass in.
// History:
//   0 = initial version
const indexVersion = 0

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

type siglistener struct {
	sigchan chan os.Signal
}

// Support closing the app with Ctrl+C when running in a shell.
func (s *siglistener) watchSignals() {
	go func() {
		if s.sigchan == nil {
			s.sigchan = make(chan os.Signal, 1)
		}
		signal.Notify(s.sigchan, syscall.SIGTERM, syscall.SIGINT)
		for {
			sig := <-s.sigchan
			switch sig {
			case syscall.SIGTERM, syscall.SIGINT:
				os.Exit(0)
			}
		}
	}()
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

	if updateConfFrom != nil && len(*updateConfFrom) > 0 {
		err = conf.UpdateFrom(*updateConfFrom)
		check(err)
		err = conf.Dump(configPath)
		check(err)
		os.Exit(0)
	}

	if *echoHash {
		fmt.Println(getHash(conf))
		os.Exit(0)
	}

	if updateChainFrom != nil && len(*updateChainFrom) > 0 {
		updateChain(*updateChainFrom, conf)
		os.Exit(0)
	}

	app, err := ndau.NewApp(getDbSpec(), getIndexAddr(), indexVersion, *conf)
	check(err)

	logger := app.GetLogger()
	logger = logger.WithField("bin", "ndaunode")
	app.SetLogger(logger)
	app.LogState()

	server := server.NewSocketServer(*socketAddr, app)

	// it's not entirely ideal that we have to generate a separate logger
	// here, but tendermint loggers have an interface incompatible with
	// logrus loggers
	// server.SetLogger(tmlog.NewTMLogger(os.Stderr))
	if logwriter, err := honeycomb.NewWriter(); err != nil {
		server.SetLogger(tmlog.NewTMLogger(os.Stderr))
		app.GetLogger().WithFields(logrus.Fields{
			"warning":       "Unable to initialize Honeycomb for tm server",
			"originalError": err,
			"bin":           "ndautendermint",
		}).Warn("InitServerLog")
		fmt.Println("Can't init server logger for tm: ", err)
	} else {
		server.SetLogger(tmlog.NewTMJSONLogger(logwriter))
	}

	err = server.Start()
	check(err)

	entry := logger.WithFields(logrus.Fields{
		"address": *socketAddr,
		"name":    server.String(),
	})

	v, err := version.Get()
	if err == nil {
		entry = entry.WithField("version", v)
	} else {
		entry = entry.WithError(err)
	}
	sl := &siglistener{}
	sl.watchSignals()
	entry.Info("started ABCI socket server")
	// we want to keep this service running indefinitely
	// if there were more commands to run, we'd probably want to split this into a separate
	// goroutine and deal with closing options, but for now, it's probably fine to actually
	// just let the main routine hang here
	<-server.Quit()
}
