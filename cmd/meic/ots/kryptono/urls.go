package kryptono

// URLs for the bitmart API
const (
	API       = "https://openapi.bitmart.com/v2/"
	APIAuth   = API + "authentication"
	APITime   = API + "time"
	APIWallet = API + "wallet"
	APITrades = API + "trades"
	APIOrders = API + "orders"
)

// SignatureHeader names the key used for signing the request body per the Bitmart strategy
const SignatureHeader = "Signature"

const WSSKryptono = "wss://openws.kryptono.com"
