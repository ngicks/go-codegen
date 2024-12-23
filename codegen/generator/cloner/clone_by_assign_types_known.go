package cloner

import "github.com/ngicks/go-codegen/codegen/imports"

// Roughly checked std: from archive to maphash.
// TODO: check rest of them

var knownCloneByAssign = map[imports.TargetType]struct{}{
	{ImportPath: "", Name: "error"}:                    {},
	{ImportPath: "crypto", Name: "PrivateKey"}:         {},
	{ImportPath: "crypto", Name: "PublicKey"}:          {},
	{ImportPath: "crypto/dsa", Name: "PrivateKey"}:     {},
	{ImportPath: "crypto/dsa", Name: "PublicKey"}:      {},
	{ImportPath: "crypto/ecdh", Name: "PrivateKey"}:    {},
	{ImportPath: "crypto/ecdsa", Name: "PrivateKey"}:   {},
	{ImportPath: "crypto/ecdsa", Name: "PublicKey"}:    {},
	{ImportPath: "crypto/ed25519", Name: "PrivateKey"}: {},
	{ImportPath: "crypto/ed25519", Name: "PublicKey"}:  {},
	{ImportPath: "database/sql", Name: "ColumnType"}:   {},
	{ImportPath: "database/sql", Name: "DB"}:           {},
	{ImportPath: "database/sql", Name: "NamedArg"}:     {},
	{ImportPath: "database/sql", Name: "Out"}:          {},
	{ImportPath: "hash/maphash", Name: "Hash"}:         {},
	{ImportPath: "unique", Name: "Handle"}:             {},
}

var knownCloneByAssignPointer = map[imports.TargetType]struct{}{
	{ImportPath: "database/sql", Name: "ColumnType"}: {},
	{ImportPath: "database/sql", Name: "DB"}:         {},
	{ImportPath: "time", Name: "Location"}:           {},
}
