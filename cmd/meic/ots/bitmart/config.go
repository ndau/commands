package bitmart

// BMArgs encapsulates the arguments which configure bitmart
type BMArgs struct {
	APIKeyPath string `arg:"--bitmart-api-key-path,required" help:"path to bitmart api key file"`
}

// HasBMArgs is implemented for types which contain BMArgs
type HasBMArgs interface {
	GetBMArgs() BMArgs
}
