package wire

// AlgorithmSpec specifies block height for given algorithm version
type AlgorithmSpec struct {
	// Height is block height at which the algorithm version is activated
	Height uint32

	// HeaderSize is block header size in bytes as an input to Equihash
	HeaderSize int

	// Version is numeric identifier of the algorithm
	Version uint8

	// Bits is the new difficulty compact representation at the point of algorithm change
	Bits uint32
}
