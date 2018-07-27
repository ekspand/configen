package version

import (
	"testing"
)

func TestInfo_ParseBuild(t *testing.T) {
	v := Info{Build: "1.2-4-314"}
	v.PopulateFromBuild()
	if v.Major != 1 {
		t.Errorf("Parsed Major version should be 1, was %d", v.Major)
	}
	if v.Minor != 2 {
		t.Errorf("Parsed Minor version should be 2, was %d", v.Minor)
	}
	if v.Float() != 1.2 {
		t.Errorf("Parsed Float version should be 1.2 was %f", v.Float())
	}
}

func TestInfo_GreaterOrEqual(t *testing.T) {
	v01 := Info{0, 1, "", 0.1}
	v02 := Info{0, 2, "", 0.2}
	v10 := Info{1, 0, "", 1.0}
	v12 := Info{1, 2, "", 1.2}
	v20 := Info{2, 0, "", 2.0}
	f := func(v, other Info, expected bool) {
		act := v.GreaterOrEqual(other)
		if act != expected {
			t.Errorf("%v GreaterOrEqual (%v) return wrong result of %v, expecting %v", v, other, act, expected)
		}
	}
	f(v01, v01, true)
	f(v02, v01, true)
	f(v10, v01, true)
	f(v20, v12, true)
	f(v02, v10, false)
	f(v01, v02, false)
}
