package dao_test

import (
	"fmt"

	"github.com/yueli-official/foundation/go/identifier"
)

var galleryFixtureNamespace = func() identifier.UUID {
	namespace, err := identifier.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		panic(err)
	}
	return identifier.Derive(namespace, []byte("gallery.dao.integration-fixtures"))
}()

func fixtureIdentifier(kind string, ordinal int) string {
	return identifier.Derive(galleryFixtureNamespace, []byte(fmt.Sprintf("%s:%d", kind, ordinal))).String()
}

func fixtureIdentifiers(kind string, count int) []string {
	values := make([]string, count)
	for index := range count {
		values[index] = fixtureIdentifier(kind, index+1)
	}
	return values
}
