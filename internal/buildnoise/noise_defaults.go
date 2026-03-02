//go:build !buildnoisegen

package buildnoise

// Deterministic fallback entropy for regular builds and tests.
var (
	entropy0 uint64 = 0x243f6a8885a308d3
	entropy1 uint64 = 0x13198a2e03707344
	entropy2 uint64 = 0xa4093822299f31d0
	entropy3 uint64 = 0x082efa98ec4e6c89
)
