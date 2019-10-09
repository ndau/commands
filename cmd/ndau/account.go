package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	cli "github.com/jawher/mow.cli"
)

func getAccount(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Command(
			"list",
			"list known accounts",
			getAccountList(verbose),
		)

		cmd.Command(
			"new",
			"create a new account",
			getAccountNew(verbose),
		)

		cmd.Command(
			"addr",
			"get the address of an account",
			getAccountAddr(verbose),
		)

		cmd.Command(
			"set-validation",
			"set this account's validation rules",
			getAccountSetValidation(verbose, emitJSON, compact),
		)

		cmd.Command(
			"create-child",
			"create this child account on the blockchain",
			getAccountCreateChild(verbose, keys, emitJSON, compact),
		)

		cmd.Command(
			"destroy",
			"remove all local knowledge of this account",
			getAccountDestroy(verbose),
		)

		cmd.Command(
			"recover",
			"recover an account from its recovery phrase",
			getAccountRecover(verbose),
		)

		cmd.Command(
			"validation",
			"change the account's validation",
			getAccountValidation(verbose, keys, emitJSON, compact),
		)

		cmd.Command(
			"query",
			"query the ndau chain about this account",
			getAccountQuery(verbose, emitJSON, compact),
		)

		cmd.Command(
			"change-recourse-period",
			"change the recourse period for outbound transfers from this account",
			getAccountChangeRecourse(verbose, keys, emitJSON, compact),
		)

		cmd.Command(
			"delegate",
			"delegate EAI calculation to a node",
			getAccountDelegate(verbose, keys, emitJSON, compact),
		)

		cmd.Command(
			"credit-eai",
			"credit EAI for accounts which have delegated to this one",
			getAccountCreditEAI(verbose, keys, emitJSON, compact),
		)

		cmd.Command(
			"lock",
			"lock this account with a specified notice period",
			getLock(verbose, keys, emitJSON, compact),
		)

		cmd.Command(
			"notify",
			"notify that this account should be unlocked once its notice period expires",
			getNotify(verbose, keys, emitJSON, compact),
		)

		cmd.Command(
			"set-rewards-target",
			"set the rewards target for this account",
			getSetRewardsDestination(verbose, keys, emitJSON, compact),
		)

		cmd.Command(
			"stake",
			"stake ndau from this account to another",
			getStake(verbose, keys, emitJSON, compact),
		)

		cmd.Command(
			"register-node",
			"register this node to activate it",
			getRegisterNode(verbose, keys, emitJSON, compact),
		)

		cmd.Command(
			"claim-node-reward",
			"claim node reward for this node",
			getClaimNodeReward(verbose, keys, emitJSON, compact),
		)

		cmd.Command(
			"set-stake-rules",
			"set stake rules for this account",
			getSetStakeRules(verbose, keys, emitJSON, compact),
		)
	}
}
