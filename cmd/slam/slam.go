package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

type args struct {
	Chain   string `help:"alternative to url to name mainnet, testnet, devnet, or localnet."`
	URL     string `help:"Full base URL to an API endpoint: default is localnet."`
	Min     int64  `help:"The minimum number of napu to transfer each time. Default: 100 napu."`
	Max     int64  `help:"The maximum number of napu to transfer each time. Default: 1000 napu."`
	NAccts  int    `help:"Number of child accounts created from the starting account; default 10."`
	NReq    int    `help:"Number of simultaneous requests that can be in flight at once. Must be <= NAccts; default 5."`
	Name    string `help:"If specified, uses this name to look up account info in the conf file used by ndau tool."`
	Account string `help:"Starting account number if name is not specified."`
	Private string `help:"Private key for signing transactions from the starting account (if name not specified)."`
}

func (args) Description() string {
	return `loadtest runs load tests on ndau

	This starts by reading the specified account, and then generating new child accounts by distributing 50 percent
	of the starting account's value into all of the child accounts.

	Each of the child accounts is then claimed.

	Each of the child accounts then loops -- checking its balance, then transferring a random value between min and max
 	to a randomly chosen other child.

	If an account's balance falls below the ability to transfer more, it stops trying. Eventually all the accounts
	will run out (unless the settlement period is zero).

	Note that this app uses http to the API, not RPC calls. The ndau config is ONLY used to load account information.
	`
}

// RequestManager simplifies making HTTP requests to the ndau api
type RequestManager struct {
	client  *http.Client
	baseurl string
}

// NewRequestManager creates a new RequestManager with sensible defaults
func NewRequestManager(base string) *RequestManager {
	r := RequestManager{
		client: &http.Client{
			Timeout: time.Second * 10,
		},
		baseurl: base,
	}
	return &r
}

// Account keeps track of a subset of account data we care about
type Account struct {
	Balance  types.Ndau           `json:"balance"`
	Addr     address.Address      `json:"address"`
	Private  signature.PrivateKey `json:"private_key"`
	Public   signature.PublicKey  `json:"public_key"`
	Sequence uint64               `json:"sequence"`
}

func (a *Account) String() string {
	as := a.Addr.String()
	return fmt.Sprintf("%s...%s: %d", as[:6], as[len(as)-4:], a.Balance)
}

// Get fetches the path from the baseurl and JSON-decodes it into interface,
// which must be a pointer.
func (r *RequestManager) Get(path string, result interface{}) error {
	resp, err := r.client.Get(r.baseurl + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Print(string(body))
		return fmt.Errorf("Got bad status '%s' from %s", resp.Status, path)
	}
	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			return err
		}
	}
	return nil
}

// Post JSON-encodes the payload and sends it to the baseurl+path,
// then JSON-decodes the response into interface,
// which must be a pointer.
func (r *RequestManager) Post(path string, payload interface{}, result interface{}) error {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(payload)
	resp, err := r.client.Post(r.baseurl+path, "application/json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Print(string(body))
		return fmt.Errorf("Got bad status '%s' from %s", resp.Status, path)
	}
	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetVersion makes sure the blockchain api is reachable and that it's talking to a node,
// and returns the Version string of the blockchain.
func (r *RequestManager) GetVersion() (string, error) {
	version := struct {
		// once the new version code lands this json tag will change or be deleted
		Version string `json:"NdauVersion"`
	}{}
	if err := r.Get("/version", &version); err != nil {
		return "", err
	}
	return version.Version, nil
}

// UpdateAccount takes a pointer to an account and updates its knowledge of the account balance
// and sequence.
func (r *RequestManager) UpdateAccount(acct *Account) error {
	if acct == nil {
		return errors.New("nil acct")
	}
	// can't just r.Get("/account/account", acct), because the ndauapi
	// isn't set up to return a single account. Instead, it returns a map of
	// address to account data.
	amap := make(map[string]Account)
	err := r.Get("/account/account/"+acct.Addr.String(), &amap)
	if err != nil {
		return err
	}
	acct.Balance = amap[acct.Addr.String()].Balance
	acct.Sequence = amap[acct.Addr.String()].Sequence
	return nil
}

// CreateAccount creates a new random account with an initial balance
func (r *RequestManager) CreateAccount(initialBalance types.Ndau) (Account, error) {
	var a Account
	// ownership keys
	pub, prv, err := signature.Generate(signature.Ed25519, nil)
	if err != nil {
		return a, err
	}
	a.Balance = initialBalance
	a.Sequence = 0
	a.Public = pub
	a.Addr, err = address.Generate(address.KindUser, a.Public.KeyBytes())
	if err != nil {
		return a, err
	}

	vpub, vprv, err := signature.Generate(signature.Ed25519, nil)
	if err != nil {
		return a, err
	}

	a.Private = vprv

	fmt.Printf("create child %s... ", a.Addr)

	// claim the newly-created account
	tx := ndau.NewClaimAccount(a.Addr, pub, []signature.PublicKey{vpub}, nil, a.Sequence+1, prv)
	a.Sequence++
	err = r.Post("/tx/submit/ClaimAccount", tx, nil)
	if err == nil {
		fmt.Println("claimed")
	} else {
		fmt.Println("error")
	}

	return a, errors.Wrap(err, "claiming child account: "+a.Addr.String())
}

// Transfer generates a transfer transaction and submits it
// It prevalidates it first.
func (r *RequestManager) Transfer(from *Account, to Account, qty types.Ndau) error {
	nextseq := from.Sequence + 1
	log.Printf(
		"Transfer %s ndau from %s (seq %d) to %s (seq %d) using seq %d",
		qty,
		from.Addr,
		from.Sequence,
		to.Addr,
		to.Sequence,
		nextseq,
	)
	tx := ndau.NewTransfer(from.Addr, to.Addr, qty, nextseq, from.Private)

	err := r.Post("/tx/submit/Transfer", tx, nil)
	if err != nil {
		return err
	}
	from.Sequence = nextseq

	return nil
}

// GenerateChildAccounts creates the appropriate number of child accounts. It uses 50% of the
// balance in the starting account and distributes it equally to all the children.
func (r *RequestManager) GenerateChildAccounts(naccts int, starting Account) ([]Account, error) {
	accts := make([]Account, naccts)
	err := r.UpdateAccount(&starting)
	log.Print("seq ", starting.Sequence)
	starting.Sequence++
	if err != nil {
		return nil, err
	}
	if naccts <= 0 {
		return nil, errors.New("naccts must not be <= 0")
	}
	if starting.Balance == 0 {
		return nil, errors.New("starting.balance must not be 0")
	}
	perAcct := starting.Balance / types.Ndau(2*naccts)
	for i := 0; i < naccts; i++ {
		acct, err := r.CreateAccount(perAcct)
		if err != nil {
			return nil, err
		}
		err = r.Transfer(&starting, acct, perAcct)
		if err != nil {
			return nil, err
		}
		accts[i] = acct
	}
	return accts, nil
}

func main() {
	a := args{
		Min:    100,
		Max:    1000,
		NAccts: 10,
		NReq:   5,
	}

	arg.MustParse(&a)

	if a.URL == "" {
		switch a.Chain {
		case "local", "localnet", "localhost", "":
			a.URL = "http://localhost:3030"
		case "main", "mainnet":
			a.URL = "https://node-0.main.ndau.tech"
		case "test", "testnet":
			a.URL = "https://testnet-0.api.ndau.tech"
		case "dev", "devnet":
			a.URL = "https://devnet-0.api.ndau.tech"
		default:
			log.Fatalf("Unknown chain: %s", a.Chain)
		}
	}

	var startingAddr address.Address
	var startingPrvKey signature.PrivateKey

	if a.Account != "" {
		addr, err := address.Validate(a.Account)
		if err != nil {
			log.Fatalf("Couldn't validate address %s: %s", a.Account, err)
		}
		startingAddr = addr
	}

	if a.Private != "" {
		ppk, err := signature.ParsePrivateKey(a.Private)
		if err != nil {
			log.Fatalf("Couldn't read private key %s: %s", a.Private, err)
		}
		startingPrvKey = *ppk
	}

	if a.Name != "" {
		conf, err := config.Load(config.GetConfigPath())
		if err != nil {
			log.Fatalf("%s loading config.", err)
		}
		if acct, ok := conf.Accounts[a.Name]; ok {
			startingAddr = acct.Address
			pks := acct.TransferPrivate()
			if len(pks) != 1 {
				log.Fatalf(
					"Can't use named account unless it has exactly 1 private transfer key (found %d).",
					len(pks),
				)
			}
			startingPrvKey = pks[0]
		}
	}

	rm := NewRequestManager(a.URL)
	ver, err := rm.GetVersion()
	if err != nil {
		log.Fatalf("Couldn't create child accounts: %s", err)
	}
	log.Printf("Version: %s", ver)

	if startingAddr.String() == "" {
		log.Fatal("Starting address or account name must be specified")
	}

	starting := Account{
		Balance:  0,
		Addr:     startingAddr,
		Private:  startingPrvKey,
		Sequence: 1,
	}
	children, err := rm.GenerateChildAccounts(a.NAccts, starting)
	if err != nil {
		log.Fatalf("Couldn't create child accounts: %s", err)
	}
	for _, a := range children {
		fmt.Printf("%#v\n", a.String())
	}
}
