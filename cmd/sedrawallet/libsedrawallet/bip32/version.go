package bip32

import "github.com/pkg/errors"

// BitcoinMainnetPrivate is the version that is used for
// bitcoin mainnet bip32 private extended keys.
// Ecnodes to xprv in base58.
var BitcoinMainnetPrivate = [4]byte{
	0x04,
	0x88,
	0xad,
	0xe4,
}

// BitcoinMainnetPublic is the version that is used for
// bitcoin mainnet bip32 public extended keys.
// Ecnodes to xpub in base58.
var BitcoinMainnetPublic = [4]byte{
	0x04,
	0x88,
	0xb2,
	0x1e,
}

// SedraMainnetPrivate is the version that is used for
// sedra mainnet bip32 private extended keys.
// Ecnodes to xprv in base58.
var SedraMainnetPrivate = [4]byte{
	0x03,
	0x8f,
	0x2e,
	0xf4,
}

// SedraMainnetPublic is the version that is used for
// sedra mainnet bip32 public extended keys.
// Ecnodes to kpub in base58.
var SedraMainnetPublic = [4]byte{
	0x03,
	0x8f,
	0x33,
	0x2e,
}

// SedraTestnetPrivate is the version that is used for
// sedra testnet bip32 public extended keys.
// Ecnodes to ktrv in base58.
var SedraTestnetPrivate = [4]byte{
	0x03,
	0x90,
	0x9e,
	0x07,
}

// SedraTestnetPublic is the version that is used for
// sedra testnet bip32 public extended keys.
// Ecnodes to ktub in base58.
var SedraTestnetPublic = [4]byte{
	0x03,
	0x90,
	0xa2,
	0x41,
}

// SedraDevnetPrivate is the version that is used for
// sedra devnet bip32 public extended keys.
// Ecnodes to kdrv in base58.
var SedraDevnetPrivate = [4]byte{
	0x03,
	0x8b,
	0x3d,
	0x80,
}

// SedraDevnetPublic is the version that is used for
// sedra devnet bip32 public extended keys.
// Ecnodes to xdub in base58.
var SedraDevnetPublic = [4]byte{
	0x03,
	0x8b,
	0x41,
	0xba,
}

// SedraSimnetPrivate is the version that is used for
// sedra simnet bip32 public extended keys.
// Ecnodes to ksrv in base58.
var SedraSimnetPrivate = [4]byte{
	0x03,
	0x90,
	0x42,
	0x42,
}

// SedraSimnetPublic is the version that is used for
// sedra simnet bip32 public extended keys.
// Ecnodes to xsub in base58.
var SedraSimnetPublic = [4]byte{
	0x03,
	0x90,
	0x46,
	0x7d,
}

func toPublicVersion(version [4]byte) ([4]byte, error) {
	switch version {
	case BitcoinMainnetPrivate:
		return BitcoinMainnetPublic, nil
	case SedraMainnetPrivate:
		return SedraMainnetPublic, nil
	case SedraTestnetPrivate:
		return SedraTestnetPublic, nil
	case SedraDevnetPrivate:
		return SedraDevnetPublic, nil
	case SedraSimnetPrivate:
		return SedraSimnetPublic, nil
	}

	return [4]byte{}, errors.Errorf("unknown version %x", version)
}

func isPrivateVersion(version [4]byte) bool {
	switch version {
	case BitcoinMainnetPrivate:
		return true
	case SedraMainnetPrivate:
		return true
	case SedraTestnetPrivate:
		return true
	case SedraDevnetPrivate:
		return true
	case SedraSimnetPrivate:
		return true
	}

	return false
}
