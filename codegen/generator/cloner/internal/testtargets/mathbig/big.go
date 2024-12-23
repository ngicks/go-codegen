package mathbig

import (
	"crypto/x509/pkix"
	"math/big"
)

type Big struct {
	Int   *big.Int
	Float *big.Float
	Rat   *big.Rat
}

// example for external type that contains *big.Int
type Pkix struct {
	A pkix.TBSCertificateList
}
