package main

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	generator "github.com/oneiro-ndev/chaos_genesis/pkg/genesis.generator"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

func updateChain(asscpath string, conf *config.Config) {
	asscFile := make(generator.AssociatedFile)
	_, err := toml.DecodeFile(asscpath, &asscFile)
	check(err)

	if len(asscFile) != 1 {
		check(errors.New("assc datafile must have exactly one key"))
	}

	var assc generator.Associated
	// because we know the length is 1, we can get the only entry
	for _, v := range asscFile {
		assc = v
	}

	app, err := ndau.NewAppSilent(getDbSpec(), "", -1, *conf)
	check(err)

	check(app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)

		for _, sa := range []sv.SysAcct{sv.CommandValidatorChange, sv.NominateNodeReward, sv.ReleaseFromEndowment} {
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

			ad, _ := st.GetAccount(addr, now)

			ad.ValidationKeys = append(ad.ValidationKeys, valkey)
			st.Accounts[addr.String()] = ad
		}

		return st, nil
	}))
}
