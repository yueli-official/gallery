package dao

import "github.com/yueli-official/foundation/go/identifier"

func newIdentifier() string {
	return identifier.MustNew().String()
}
