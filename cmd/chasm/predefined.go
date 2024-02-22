// Code generated automatically by "make generate"; DO NOT EDIT.

package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

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
		"ACCT_HOLDS":                   "70",
		"ACCT_SEQUENCE":                "71",
		"ACCT_CURRENCYSEATDATE":        "72",
		"ACCT_PARENT":                  "73",
		"ACCT_PROGENITOR":              "74",
		"ACCT_COSTAKERS":               "76",
		"EVENT_DEFAULT":                "0",
		"EVENT_TRANSFER":               "1",
		"EVENT_CHANGEVALIDATION":       "2",
		"EVENT_RELEASEFROMENDOWMENT":   "3",
		"EVENT_CHANGERECOURSEPERIOD":   "4",
		"EVENT_DELEGATE":               "5",
		"EVENT_CREDITEAI":              "6",
		"EVENT_LOCK":                   "7",
		"EVENT_NOTIFY":                 "8",
		"EVENT_SETREWARDSDESTINATION":  "9",
		"EVENT_SETVALIDATION":          "10",
		"EVENT_STAKE":                  "11",
		"EVENT_REGISTERNODE":           "12",
		"EVENT_NOMINATENODEREWARD":     "13",
		"EVENT_CLAIMNODEREWARD":        "14",
		"EVENT_TRANSFERANDLOCK":        "15",
		"EVENT_COMMANDVALIDATORCHANGE": "16",
		"EVENT_UNREGISTERNODE":         "18",
		"EVENT_UNSTAKE":                "19",
		"EVENT_ISSUE":                  "20",
		"EVENT_CREATECHILDACCOUNT":     "21",
		"EVENT_RECORDPRICE":            "22",
		"EVENT_SETSYSVAR":              "23",
		"EVENT_SETSTAKERULES":          "24",
		"EVENT_RECORDENDOWMENTNAV":     "25",
		"EVENT_RESOLVESTAKE":           "26",
		"EVENT_BURN":                   "27",
		"EVENT_BURNANDMINT":			"28",
		"EVENT_CHANGESCHEMA":           "30",
		"LOCK_NOTICEPERIOD":            "91",
		"LOCK_UNLOCKSON":               "92",
		"LOCK_BONUS":                   "93",
		"LOCK":                         "78",
		"RECOURSESETTINGS":             "80",
		"STAKERULES":                   "79",
		"TX_SOURCE":                    "1",
		"TX_DESTINATION":               "2",
		"TX_TARGET":                    "3",
		"TX_NODE":                      "4",
		"TX_STAKETO":                   "5",
		"TX_NAME":                      "6",
		"TX_VALUE":                     "7",
		"TX_RULES":                     "8",
		"TX_QUANTITY":                  "11",
		"TX_BURN":                      "12",
		"TX_POWER":                     "17",
		"TX_PERIOD":                    "21",
		"TX_NEWKEYS":                   "31",
		"TX_VALIDATIONSCRIPT":          "32",
		"TX_DISTRIBUTIONSCRIPT":        "33",
		"TX_OWNERSHIP":                 "34",
		"TX_STAKERULES":                "35",
		"TX_RANDOM":                    "41",
		"TX_ETHADDR":					"51",
	}
	return k
}
