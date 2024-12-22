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

type Pkix struct {
	A pkix.TBSCertificateList
}
