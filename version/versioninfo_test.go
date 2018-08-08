package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo_ParseBuild(t *testing.T) {
	v := Info{Build: "1.2-4-314"}
	v.PopulateFromBuild()
	assert.Equal(t, uint(1), v.Major)
	assert.Equal(t, uint(2), v.Minor)
	assert.Equal(t, float32(1.2), v.Float())
	assert.Equal(t, "1.2", v.String())
}

func TestInfo_GreaterOrEqual(t *testing.T) {
	v01 := Info{0, 1, "", 0.1}
	v02 := Info{0, 2, "", 0.2}
	v10 := Info{1, 0, "", 1.0}
	v12 := Info{1, 2, "", 1.2}
	v20 := Info{2, 0, "", 2.0}
	f := func(v, other Info, expected bool) {
		act := v.GreaterOrEqual(other)
		assert.Equal(t, expected, act, "%v GreaterOrEqual (%v) return wrong result of %v, expecting %v", v, other, act, expected)
	}
	f(v01, v01, true)
	f(v02, v01, true)
	f(v10, v01, true)
	f(v20, v12, true)
	f(v02, v10, false)
	f(v01, v02, false)
}

func TestInfo_Current(t *testing.T) {
	assert.NotEmpty(t, Current().String())
}
