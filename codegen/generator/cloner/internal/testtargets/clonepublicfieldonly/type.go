package clonepublicfieldonly

import "archive/tar"

// nolint
type example struct {
	f tar.Header
}
