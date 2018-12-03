// Code generated automatically by "make generate"; DO NOT EDIT.

package main

// Predefined constants available to chasm programs.

func predefinedConstants() map[string]string {
	k := map[string]string{
		"ACCT_BALANCE":                 "61",
		"ACCT_VALIDATIONKEYS":          "62",
		"ACCT_REWARDSTARGET":           "63",
		"ACCT_INCOMINGREWARDSFROM":     "64",
		"ACCT_DELEGATIONNODE":          "65",
		"ACCT_LASTEAIUPDATE":           "66",
		"ACCT_LASTWAAUPDATE":           "67",
		"ACCT_WEIGHTEDAVERAGEAGE":      "68",
		"ACCT_VALIDATIONSCRIPT":        "69",
		"ACCT_SETTLEMENTS":             "70",
		"EVENT_DEFAULT":                "0",
		"EVENT_TRANSFER":               "1",
		"EVENT_CHANGEVALIDATION":       "2",
		"EVENT_RELEASEFROMENDOWMENT":   "3",
		"EVENT_CHANGESETTLEMENTPERIOD": "4",
		"EVENT_DELEGATE":               "5",
		"EVENT_CREDITEAI":              "6",
		"EVENT_LOCK":                   "7",
		"EVENT_NOTIFY":                 "8",
		"EVENT_SETREWARDSDESTINATION":  "9",
		"EVENT_CLAIMACCOUNT":           "10",
		"EVENT_STAKE":                  "11",
		"EVENT_REGISTERNODE":           "12",
		"EVENT_NOMINATENODEREWARD":     "13",
		"EVENT_CLAIMNODEREWARD":        "14",
		"EVENT_TRANSFERANDLOCK":        "15",
		"EVENT_GTVALIDATORCHANGE":      "255",
		"LOCK_NOTICEPERIOD":            "91",
		"LOCK_UNLOCKSON":               "92",
		"LOCK_BONUS":                   "93",
		"SETTLEMENTSETTINGS_PERIOD":    "111",
		"SETTLEMENTSETTINGS_CHANGESAT": "112",
		"SETTLEMENTSETTINGS_NEXT":      "113",
		"TX_SOURCE":                    "1",
		"TX_DESTINATION":               "2",
		"TX_TARGET":                    "3",
		"TX_NODE":                      "4",
		"TX_QUANTITY":                  "11",
		"TX_PERIOD":                    "21",
		"TX_NEWKEYS":                   "31",
		"TX_VALIDATIONSCRIPT":          "32",
		"TX_DISTRIBUTIONSCRIPT":        "33",
		"TX_RPCADDRESS":                "34",
		"TX_RANDOM":                    "41",
	}
	return k
}
