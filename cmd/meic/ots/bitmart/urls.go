package bitmart

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
const SignatureHeader = "X-Bm-Signature"

// WSSBitmart is bitmart's websocket URL
const WSSBitmart = "wss://openws.bitmart.com"
