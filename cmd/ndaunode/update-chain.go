package main

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	generator "github.com/oneiro-ndev/system_vars/pkg/genesis.generator"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

func updateChain(asscpath string, conf *config.Config) {
	assc := make(generator.Associated)
	_, err := toml.DecodeFile(asscpath, &assc)
	check(err)

	app, err := ndau.NewAppSilent(getDbSpec(), "", -1, *conf)
	check(err)

	check(app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)

		for _, sa := range []sv.SysAcct{
			sv.CommandValidatorChange,
			sv.NominateNodeReward,
			sv.ReleaseFromEndowment,
			sv.RecordPrice,
			sv.SetSysvar,
		} {
			addrV, addrok := assc[sa.Address]
			valkeyV, valkeyok := assc[sa.Validation.Public]
			if !(addrok && valkeyok) {
				continue
			}

			// parse address
			addrS, ok := addrV.(string)
			if !ok {
				return st, fmt.Errorf("%s address not stored as string", sa.Name)
			}

			addr, err := address.Validate(addrS)
			if err != nil {
				return st, errors.Wrap(err, "validating "+sa.Address)
			}

			// parse pubkey
			valkeyS, ok := valkeyV.(string)
			if !ok {
				return st, fmt.Errorf("%s validator public key not stored as string", sa.Name)
			}

			var valkey signature.PublicKey
			err = valkey.UnmarshalText([]byte(valkeyS))
			if err != nil {
				return st, errors.Wrap(err, sa.Validation.Public+" invalid")
			}

			// update state
			now, err := math.TimestampFrom(time.Now())
			if err != nil {
				return st, errors.Wrap(err, "computing current timestamp")
			}

			// it would be a pain to ensure that we had system variables here,
			// and this applies only to special accounts anyway, so the best
			// solution is to have them simply start with a 0 settlement period.
			ad, _ := st.GetAccount(addr, now, 0)

			ad.ValidationKeys = append(ad.ValidationKeys, valkey)
			st.Accounts[addr.String()] = ad
		}

		return st, nil
	}))
}
