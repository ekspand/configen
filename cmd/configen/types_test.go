package main

import (
	"testing"
)

func Test_VerifyStdNames(t *testing.T) {
	stdNames := stdTypeNames()
	for _, n := range stdNames {
		if ti, exists := stdTypesByName[n]; !exists {
			t.Errorf("stdTypeNames() return type %s, but that doesn't appear in teh stdTypes map!", n)
		} else {
			if ti.Name != n {
				t.Errorf("stdType %s is in the stdTypes map, but the typeInfo has type %s", n, ti.Name)
			}
		}
	}
	if len(stdNames) != len(stdTypes) {
		t.Errorf("Differing number of standard type names, from standard types! %d / %d", len(stdNames), len(stdTypes))
	}
	if len(stdNames) != len(stdTypesByName) {
		t.Errorf("Differing number of standard type names, from standard types map! %d / %d", len(stdNames), len(stdTypesByName))
	}
}

func Test_TypeInfoShared(t *testing.T) {
	for _, ti := range stdTypes {
		byname := stdTypesByName[ti.Name]
		if ti != byname {
			t.Errorf("stdType list & byName map should both point to the same typeInfo instance, but got %p/%p", ti, byname)
		}
	}
}

func Test_TypesHasExamples(t *testing.T) {
	for _, ti := range stdTypes {
		if len(ti.ExampleValues) < 3 {
			t.Errorf("Type %v only has %d ExampleValues should have at least 3", ti.Name, len(ti.ExampleValues))
		}
	}
}

func Test_InvalidOverrideExpr(t *testing.T) {
	bogus := overrideExprType(99)
	expr := bogus.overrideExpr()
	if expr != "*UNEXPECTED_OVERRIDE_STYLE_99" {
		t.Errorf("invalid overrideExprType value returned unexpected expr of %v", expr)
	}
}
