package main

import (
	"fmt"
	"os"

	tc "github.com/oneiro-ndev/ndau/pkg/tool.config"
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func main() {
	config, err := tc.LoadDefault(tc.GetConfigPath())
	check(err)
	fmt.Println("successfully loaded config from:")
	fmt.Println(tc.GetConfigPath())
	fmt.Println("acccounts by name:")
	for name, acct := range config.Accounts {
		fmt.Printf(" %s: %s\n", name, acct.Name)
	}
	fmt.Println("getaccounts:")
	for _, acct := range config.GetAccounts() {
		fmt.Println("", acct.Name)
	}
	err = config.Save()
	check(err)
	fmt.Println("successfully saved config")
}
