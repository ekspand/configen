package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_VerifyStdNames(t *testing.T) {
	stdNames := stdTypeNames()
	for _, n := range stdNames {
		ti, exists := stdTypesByName[n]
		require.True(t, exists, "stdTypeNames() return type %s, but that doesn't appear in teh stdTypes map!", n)
		assert.Equal(t, n, ti.Name, "stdType %s is in the stdTypes map, but the typeInfo has type %s", n, ti.Name)
	}

	assert.Equal(t, len(stdNames), len(stdTypes), "Differing number of standard type names, from standard types")
	assert.Equal(t, len(stdNames), len(stdTypesByName), "Differing number of standard type names, from standard types map")
}

func Test_TypeInfoShared(t *testing.T) {
	for _, ti := range stdTypes {
		byname := stdTypesByName[ti.Name]
		assert.Equal(t, ti, byname, "stdType list & byName map should both point to the same typeInfo instance, but got %p/%p", ti, byname)
	}
}

func Test_TypesHasExamples(t *testing.T) {
	for _, ti := range stdTypes {
		assert.True(t, len(ti.ExampleValues) >= 3, "Type %v only has %d ExampleValues should have at least 3", ti.Name, len(ti.ExampleValues))
	}
}

func Test_InvalidOverrideExpr(t *testing.T) {
	bogus := overrideExprType(99)
	expr := bogus.overrideExpr()
	assert.Equal(t, "*UNEXPECTED_OVERRIDE_STYLE_99", expr)
}
