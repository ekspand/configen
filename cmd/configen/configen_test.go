package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var getConfigDef = []struct {
	configDef     *configDef
	configDefFunc func() *configDef
	expected      int
	name          string
}{
	{
		configDefFunc: func() *configDef {
			return &configDef{
				Configuration: &structInfo{Fields: []fieldInfo{{Name: "a", Type: "[]int"}, {Name: "b", Type: "Bob"}, {Name: "c", Type: "Peters"}}},
				RelatedTypes: map[string]*structInfo{
					"Bob":    {Fields: []fieldInfo{{Name: "a", Type: "int"}, {Name: "b", Type: "Alice"}, {Name: "c", Type: "Alice"}}},
					"Alice":  {Fields: []fieldInfo{{Name: "a", Type: "string"}, {Name: "b", Type: "string"}, {Name: "b", Type: "uint64"}}},
					"Peter":  {Fields: []fieldInfo{{Name: "a", Type: "string"}, {Name: "b", Type: "string"}, {Name: "b", Type: "uint64"}}},
					"Peters": {Fields: []fieldInfo{{Name: "peters", Type: "[]Peter"}}},
				},
			}
		},
		expected: 5,
		name:     "StructValuesWithArrayInRelatedTypes",
	},
	{
		configDefFunc: func() *configDef {
			return &configDef{
				Configuration: &structInfo{Fields: []fieldInfo{{Name: "a", Type: "[]int"}, {Name: "b", Type: "Bob"}, {Name: "c", Type: "[]Peter"}}},
				RelatedTypes: map[string]*structInfo{
					"Bob":   {Fields: []fieldInfo{{Name: "a", Type: "int"}, {Name: "b", Type: "Alice"}, {Name: "c", Type: "Alice"}}},
					"Alice": {Fields: []fieldInfo{{Name: "a", Type: "string"}, {Name: "b", Type: "string"}, {Name: "b", Type: "uint64"}}},
					"Peter": {Fields: []fieldInfo{{Name: "a", Type: "string"}, {Name: "b", Type: "string"}, {Name: "b", Type: "uint64"}}},
				},
			}
		},
		expected: 4,
		name:     "StructValuesWithArrayInConfigurationObject",
	},
	{
		configDefFunc: func() *configDef {
			return &configDef{
				Configuration: &structInfo{Fields: []fieldInfo{{Name: "a", Type: "[]int"}, {Name: "b", Type: "Bob"}}},
				RelatedTypes: map[string]*structInfo{
					"Bob":   {Fields: []fieldInfo{{Name: "a", Type: "int"}, {Name: "b", Type: "Alice"}, {Name: "c", Type: "Alice"}}},
					"Alice": {Fields: []fieldInfo{{Name: "a", Type: "string"}, {Name: "b", Type: "string"}, {Name: "b", Type: "uint64"}}},
				},
			}
		},
		expected: 3,
		//Test_StructExampleValues defines some custom types, and verifies that ExampleValues are built for all the types
		name: "StructExampleValuesWithCustomTypes",
	},
}

func Test_StructValues(t *testing.T) {
	for _, testStruct := range getConfigDef {
		t.Run(testStruct.name, func(t *testing.T) {
			testStruct.configDef = testStruct.configDefFunc()
			td, err := testStruct.configDef.processConfig()
			require.NoError(t, err)

			hasExamples := func(ty *typeInfo) {
				assert.True(t, len(ty.ExampleValues) >= 6,
					"Struct type %s only has %d examples, should have at least 6 for %s ",
					ty.Name, len(ty.ExampleValues), testStruct.name)
			}
			for _, ct := range td.Structs {
				hasExamples(ct.GoType)
			}
			assert.Equal(t, testStruct.expected, len(td.Structs))
		})
	}
}

func Test_defaultPacakgeName(t *testing.T) {
	pn, err := defaultPackageName(".")
	require.NoError(t, err)
	require.Equal(t, "configen", pn)
}

func pkgDir(t *testing.T) string {
	d, err := findPackageDir()
	require.NoError(t, err)
	return d
}

func Test_InvalidType(t *testing.T) {
	err := generateConfig(pkgDir(t), "testdata/invalid_type_test.json", ".")
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "Field bob has type Alice which isn't valid"))

	err = generateConfig(pkgDir(t), "testdata/invalid_type_rel_test.json", ".")
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "Field bob has type Alice which isn't valid"))
}

func Test_InvalidJson(t *testing.T) {
	err := generateConfig(pkgDir(t), "testdata/invalid_json_test.json", ".")
	assert.Error(t, err)
	assert.True(t,
		strings.HasPrefix(err.Error(), "Unable to parse configuration definition file testdata/invalid_json_test.json: invalid character 'b' looking for beginning of object"),
		"got: "+err.Error())

	err = generateConfig(pkgDir(t), "testdata/missing.json", ".")
	assert.Error(t, err)
	assert.True(t,
		strings.HasPrefix(err.Error(), "Unable to open supplied configuration definition file testdata/missing.json: open testdata/missing.json: no such file or directory"),
		"got: "+err.Error())
}

// Test_GenCompileTest does a end 2 end test, it does a code-gen and then builds & tests the generated code
// it does this for all the gen*.json files in this package.
func Test_GenCompileTest(t *testing.T) {
	testConfigs := findTestConfigs(t)
	for _, tc := range testConfigs {
		t.Run(tc.Name(), func(t *testing.T) {
			genCompileTest(t, tc)
		})
	}
}

func genCompileTest(t *testing.T, tc os.FileInfo) {
	pkgName := strings.TrimRight(tc.Name(), ".json")
	destDir := pkgDir(t) + "/.tmp/" + pkgName

	err := os.RemoveAll(destDir)
	require.NoError(t, err)

	err = os.MkdirAll(destDir, 0775)
	require.NoError(t, err)

	defer func() {
		if !t.Failed() {
			os.RemoveAll(destDir)
		}
	}()
	err = generateConfig(pkgDir(t), "testdata/"+tc.Name(), destDir)
	require.NoError(t, err)

	goBuild(t, pkgName, destDir)
	goTest(t, pkgName, destDir)
}

func goBuild(t *testing.T, testName, dir string) {
	b := exec.Command("go", "build", "-a", "-v", ".")
	b.Dir = dir
	res, err := b.CombinedOutput()
	require.NoError(t, err)

	sr := string(res)
	t.Logf("go build output: %s", sr)
	assert.True(t, strings.Contains(sr, testName))
}

func goTest(t *testing.T, testName, dir string) {
	b := exec.Command("go", "test", ".")
	b.Dir = dir
	res, err := b.CombinedOutput()
	require.NoError(t, err)

	sr := string(res)
	t.Logf("go test output: %s", sr)
	assert.True(t, strings.HasPrefix(sr, "ok "))
}

// returns all files in the package directory that match the gen*.json pattern
func findTestConfigs(t *testing.T) []os.FileInfo {
	p := regexp.MustCompile("gen.*\\.json")
	files, err := ioutil.ReadDir("testdata")
	require.NoError(t, err)

	res := make([]os.FileInfo, 0, 10)
	for _, f := range files {
		if !f.IsDir() && p.MatchString(f.Name()) {
			res = append(res, f)
		}
	}
	require.True(t, len(res) > 0)
	return res
}
