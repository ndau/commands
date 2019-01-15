package main

import (
	cli "github.com/jawher/mow.cli"
)

func getAccount(verbose bool, keys int, emitJSON, pretty bool) func(*cli.Cmd) {
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
			"recover",
			"recover an account from its recovery phrase",
			getAccountRecover(verbose),
		)

		cmd.Command(
			"claim",
			"claim this account on the blockchain",
			getAccountClaim(verbose),
		)

		cmd.Command(
			"validation",
			"change the account's validation",
			getAccountValidation(verbose, keys, emitJSON, pretty),
		)

		cmd.Command(
			"query",
			"query the ndau chain about this account",
			getAccountQuery(verbose),
		)

		cmd.Command(
			"change-settlement-period",
			"change the settlement period for outbound transfers from this account",
			getAccountChangeSettlement(verbose, keys, emitJSON, pretty),
		)

		cmd.Command(
			"delegate",
			"delegate EAI calculation to a node",
			getAccountDelegate(verbose, keys, emitJSON, pretty),
		)

		cmd.Command(
			"credit-eai",
			"credit EAI for accounts which have delegated to this one",
			getAccountCreditEAI(verbose, keys, emitJSON, pretty),
		)

		cmd.Command(
			"lock",
			"lock this account with a specified notice period",
			getLock(verbose, keys, emitJSON, pretty),
		)

		cmd.Command(
			"notify",
			"notify that this account should be unlocked once its notice period expires",
			getNotify(verbose, keys, emitJSON, pretty),
		)

		cmd.Command(
			"set-rewards-target",
			"set the rewards target for this account",
			getSetRewardsDestination(verbose, keys, emitJSON, pretty),
		)

		cmd.Command(
			"stake",
			"stake this account to another",
			getStake(verbose, keys, emitJSON, pretty),
		)

		cmd.Command(
			"register-node",
			"register this node to activate it",
			getRegisterNode(verbose, keys, emitJSON, pretty),
		)

		cmd.Command(
			"claim-node-reward",
			"claim node reward for this node",
			getClaimNodeReward(verbose, keys, emitJSON, pretty),
		)
	}
}
