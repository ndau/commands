package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/oneiro-ndev/chaos/pkg/chaos"
	"github.com/oneiro-ndev/chaos/pkg/chaos/config"
	evndau "github.com/oneiro-ndev/chaos/pkg/ev.ndau"
	"github.com/oneiro-ndev/o11y/pkg/honeycomb"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/server"
	tmlog "github.com/tendermint/tendermint/libs/log"
)

var useNh = flag.Bool("use-ndauhome", false, "if set, keep database within $NDAUHOME/chaos")
var dbspec = flag.String("spec", "", "manually set the noms db spec")
var indexAddr = flag.String("index", "", "search index address")
var socketAddr = flag.String("addr", "0.0.0.0:26658", "socket address for incoming connection from tendermint")
var echoSpec = flag.Bool("echo-spec", false, "if set, echo the DB spec used and then quit")
var echoEmptyHash = flag.Bool("echo-empty-hash", false, "if set, echo the hash of the empty DB and then quit")
var echoHash = flag.Bool("echo-hash", false, "if set, echo the current DB hash and then quit")
var setNdaunode = flag.String("set-ndaunode", "", "set the configured ndau node address and quit")
var unsetNdaunode = flag.Bool("unset-ndaunode", false, "unset ndau node in configuration and quit")

// Bump this any time we need to reset and reindex the chaos chain.  For example, if we change the
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

func getChaosConfigDir() string {
	return filepath.Join(getNdauhome(), "chaos")
}

func getDbSpec() string {
	if len(*dbspec) > 0 {
		return *dbspec
	}
	if *useNh {
		return filepath.Join(getChaosConfigDir(), "data")
	}
	// default to noms server for dockerization
	return "http://noms:8000"
}

func getIndexAddr() string {
	if len(*indexAddr) > 0 {
		return *indexAddr
	}
	if *useNh {
		return filepath.Join(getChaosConfigDir(), "redis")
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

func getConf() *config.Config {
	conf, err := config.LoadDefault(config.DefaultConfigPath(getNdauhome()))
	check(err)
	return conf
}

type siglistener struct {
	sigchan chan os.Signal
}

// Support closing the app with Ctrl+C when running in a shell.
func (s *siglistener) watchSignals(app *chaos.App) {
	go func() {
		if s.sigchan == nil {
			s.sigchan = make(chan os.Signal, 1)
		}
		signal.Notify(s.sigchan, syscall.SIGTERM, syscall.SIGINT)
		for {
			sig := <-s.sigchan
			switch sig {
			case syscall.SIGTERM, syscall.SIGINT:
				app.GetLogger().Info("Terminated with an OS signal")
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

	if *echoHash {
		fmt.Println(getHash())
		os.Exit(0)
	}

	setNdaunodeF(setNdaunode)
	unsetNdaunodeF(unsetNdaunode)

	conf := getConf()

	app, err := chaos.NewApp(getDbSpec(), getIndexAddr(), indexVersion, conf)
	check(err)

	if len(conf.NdauAddress) > 0 {
		app.SetValidator(evndau.New(conf.NdauAddress))
	}

	logger := app.GetLogger()
	logger = logger.WithField("bin", "chaosnode")
	app.SetLogger(logger)
	app.LogState()

	sa := *socketAddr
	server := server.NewSocketServer(sa, app)

	// it's not entirely ideal that we have to generate a separate logger
	// here, but tendermint loggers have an interface incompatible with
	// logrus loggers
	// server.SetLogger(tmlog.NewTMLogger(os.Stderr))
	if logwriter, err := honeycomb.NewWriter(); err != nil {
		server.SetLogger(tmlog.NewTMLogger(os.Stderr))
		app.GetLogger().WithFields(logrus.Fields{
			"warning":       "Unable to initialize Honeycomb for tm server",
			"originalError": err,
		}).Warn("InitServerLog")
		fmt.Println("Can't init server logger for tm: ", err)
	} else {
		server.SetLogger(tmlog.NewTMJSONLogger(logwriter))
	}

	err = server.Start()
	check(err)

	sl := &siglistener{}
	sl.watchSignals(app)

	logger.WithFields(logrus.Fields{
		"address": sa,
		"name":    server.String(),
	}).Info("started ABCI socket server")
	// we want to keep this service running indefinitely
	// if there were more commands to run, we'd probably want to split this into a separate
	// goroutine and deal with closing options, but for now, it's probably fine to actually
	// just let the main routine hang here
	<-server.Quit()
}
