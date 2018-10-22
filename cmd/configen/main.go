package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/go-phorce/configen/version"
	"github.com/juju/errors"
)

// usage config-gen -c <config_def.json> -d <dest path>
// see the configDef type for details of what the config_def.json file should contain
func main() {
	ver := flag.Bool("v", false, "Print version")
	def := flag.String("c", "", "Filename of the configuration definition file")
	dest := flag.String("d", ".", "Directory to write generated files(s) to")
	// i.e. if config-gen is in foo/bar/src/github.com/go-phorce/configen
	// then you can specify -pkg foo/bar
	pkgDir := flag.String("pkg", "", "Directory containing root of path to this package [by default find it from $GOPATH]")
	flag.Parse()

	if *ver {
		fmt.Printf("configen %v\n", version.Current())
		os.Exit(0)
	}

	if *def == "" {
		log.Fatal("must specify the name of the configuration definition file")
	}
	if _, err := os.Stat(*def); os.IsNotExist(err) {
		log.Fatalf("unable to find supplied configuration defintition file: %v", *def)
	}
	pdir := ""
	if *pkgDir == "" {
		var err error
		pdir, err = findPackageDir()
		if err != nil {
			log.Println(err.Error())
			os.Exit(-1)
		}
	} else {
		pdir = filepath.Join(*pkgDir, "src/github.com/go-phorce/configen")
	}
	err := generateConfig(pdir, *def, *dest)
	if err != nil {
		log.Println(err.Error())
		os.Exit(-1)
	}
}

type commentable struct {
	Comment string
}

func (c commentable) PrefixedComment() string {
	if strings.TrimSpace(c.Comment) == "" {
		return ""
	}
	lines := strings.Split(c.Comment, "\n")
	for i := range lines {
		lines[i] = `// ` + strings.TrimSpace(lines[i])
	}
	return strings.Join(lines, "\n")
}

// fieldInfo is metadata about a single field declared in the config definition
type fieldInfo struct {
	commentable
	Name string
	Type string
	// GoType will be populated by code, not from the json [this is exported so the template can access it]
	GoType *typeInfo
}

// OverrideImpl pipe returns overrideFrom implementation
func (f *fieldInfo) OverrideImpl() string {
	if f.GoType.overrideStyle == osStruct {
		return fmt.Sprintf("c.%s.overrideFrom(&o.%s)", f.Name, f.Name)
	}
	return fmt.Sprintf("%s(&c.%s, &o.%s)", f.GoType.OverrideFunc, f.Name, f.Name)
}

// IsStruct returns true for Structure type
func (f *fieldInfo) IsStruct() bool {
	return f.GoType.overrideStyle == osStruct
}

// IsBoolPtr returns true for *bool type
func (f *fieldInfo) IsBoolPtr() bool {
	return f.GoType.overrideStyle == osCompareNil && f.Type == "*bool"
}

// IsDuration returns true for Duration type
func (f *fieldInfo) IsDuration() bool {
	return f.GoType.overrideStyle == osCompareZero && f.Type == "Duration"
}

// structInfo is metadata about a collection of fields that are mapped to a single go struct
type structInfo struct {
	commentable
	WithGetter bool
	Fields     []fieldInfo
	// GoType will be created & populated via the config post processing, not from the json [this is exported so the template can access it]
	GoType *typeInfo
}

// GettersImpl pipe returns Getters implementation
func (s *structInfo) GettersImpl() string {
	list := []string{}

	intName := s.GoType.Name + "Config"
	list = append(list, strings.Replace(s.PrefixedComment(), s.GoType.Name, intName, -1))
	list = append(list, fmt.Sprintf("type %s interface {", intName))
	for _, f := range s.Fields {
		var ft string
		mn := "Get" + f.Name
		if f.GoType.overrideStyle == osStruct {
			mn += "Cfg"
			ft = fmt.Sprintf("%s() *%s", mn, f.Type)
		} else if f.IsBoolPtr() {
			mn += "Cfg"
			ft = fmt.Sprintf("%s() bool", mn)
		} else if f.IsDuration() {
			mn += "Cfg"
			ft = fmt.Sprintf("%s() time.Duration", mn)
		} else {
			ft = fmt.Sprintf("%s() %s", mn, f.Type)
		}
		list = append(list, strings.Replace(f.PrefixedComment(), f.GoType.Name, mn, -1))
		list = append(list, ft)
	}
	list = append(list, "}")

	for _, f := range s.Fields {
		var fs string
		mn := "Get" + f.Name
		if f.GoType.overrideStyle == osStruct {
			mn += "Cfg"
			fs = fmt.Sprintf("func (c *%s) %s() *%s {\n\treturn &c.%s\n}",
				s.GoType.Name,
				mn,
				f.Type,
				f.Name,
			)
			strings.Replace(f.PrefixedComment(), f.Name, mn, -1)
		} else if f.IsBoolPtr() {
			fs = fmt.Sprintf("func (c *%s) %s() bool {\n\treturn c.%s != nil && *c.%s\n}",
				s.GoType.Name,
				mn,
				f.Name,
				f.Name,
			)
		} else if f.IsDuration() {
			fs = fmt.Sprintf("func (c *%s) %s() time.Duration {\n\treturn c.%s.TimeDuration()\n}",
				s.GoType.Name,
				mn,
				f.Name,
			)
		} else {
			fs = fmt.Sprintf("func (c *%s) %s() %s {\n\treturn c.%s\n}",
				s.GoType.Name,
				mn,
				f.Type,
				f.Name,
			)
		}
		list = append(list, strings.Replace(f.PrefixedComment(), f.Name, mn, -1))
		list = append(list, fs)
	}
	return strings.Join(list, "\n")
}

// configDef is the data loaded from the configuration defintion json file.
// this must define a Configuration type, and may also define additional related types to generate
type configDef struct {
	// PackageName the name of the go package in the generated code, defaults based on the last path segment if not set
	PackageName string
	// Configuration the primary/top level configuration type.
	Configuration *structInfo
	// RelatedTypes contains any additional types that are referenced from the Configuration defintion, go structs will be generated for these
	RelatedTypes map[string]*structInfo

	// customTypeInfos is populated during post processing
	customTypeInfos map[string]*typeInfo
}

// this is the collection of data exposed to the template to generate the new config.go
type templateData struct {
	PackageName string
	Structs     map[string]*structInfo
	BaseTypes   typeInfos
}

// findPackageDir will look for this package from the GOPATH, and return the directory that contains it
// [this is used to load the template files from]
func findPackageDir() (string, error) {
	wd, _ := os.Getwd()
	p, err := build.Import("github.com/go-phorce/configen", wd, build.FindOnly)
	if err != nil {
		return "", errors.Trace(err)
	}
	return p.Dir, nil
}

// generateConfig will load the config definition file, and generate the resulting
// config.go file & config_test.go files in the supplied directory.
func generateConfig(packageDir, defFile, destDir string) error {
	def, err := loadConfig(defFile)
	if err != nil {
		return errors.Trace(err)
	}
	if def.PackageName == "" {
		defName, err := defaultPackageName(destDir)
		if err != nil {
			return errors.Trace(err)
		}
		def.PackageName = defName
	}
	t, err := template.ParseFiles(filepath.Join(packageDir, "cmd/configen/config.go.template"),
		filepath.Join(packageDir, "cmd/configen/config_test.go.template"))
	if err != nil {
		return fmt.Errorf("unable to open/parse template files: %v", err)
	}
	gen := func(destFilename, templateName string) (string, error) {
		destFile := filepath.Join(destDir, destFilename)
		df, err := os.Create(destFile)
		if err != nil {
			return "", fmt.Errorf("unable to create destination file: %v", err)
		}
		defer df.Close()
		if err := t.ExecuteTemplate(df, templateName, def); err != nil {
			return "", fmt.Errorf("unable to generate code: %v", err)
		}
		return destFile, nil
	}
	configFile, err := gen("config.go", "config.go.template")
	if err != nil {
		return errors.Trace(err)
	}
	testFile, err := gen("config_test.go", "config_test.go.template")
	if err != nil {
		return errors.Trace(err)
	}
	gofmt(configFile)
	gofmt(testFile)
	return nil
}

// defaultPackageName returns the default name for the package based on the
// directory its contained in
func defaultPackageName(dir string) (string, error) {
	absDest, err := filepath.Abs(dir)
	if err != nil {
		return "", errors.Errorf("unable to resolve destination directory: %v", err)
	}
	return filepath.Base(absDest), nil
}

// loadConfig will load the supplied config file, parse it, and do some post processing
// to build the templateData instance.
func loadConfig(defFile string) (*templateData, error) {
	// parse the supplied json config file
	var def configDef
	f, err := os.Open(defFile)
	if err != nil {
		return nil, errors.Errorf("unable to open supplied configuration definition file %v: %v", defFile, err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&def); err != nil {
		return nil, errors.Errorf("unable to parse configuration definition file %v: %v", defFile, err)
	}
	return def.processConfig()
}

// getRelatedType will retrieve the related typeinfo based on the type name
func (def *configDef) getRelatedType(fType string) (*structInfo, bool) {
	if strings.HasPrefix(fType, "[]") {
		return def.getRelatedType(fType[2:])
	}
	rt, ok := def.RelatedTypes[fType]
	return rt, ok
}

// processConfig does our post processing to generate the templateData instance
// we need to set the GoType in all the Fields & Structs, and verify that all
// the referenced types exist
func (def *configDef) processConfig() (*templateData, error) {
	def.customTypeInfos = make(map[string]*typeInfo)
	usedTypes := make(map[string]*typeInfo) // type Name -> Type
	processField := func(f *fieldInfo) error {
		ti, ok := stdTypesByName[f.Type]
		if !ok {
			rt, ok := def.getRelatedType(f.Type)
			if !ok {
				return errors.Errorf("field %v has type %v which isn't valid (valid types are %v)", f.Name, f.Type, strings.Join(stdTypeNames(), ","))
			}
			ti = def.ensureRelatedTypeInfo(f.Type, rt)
		}
		if ti.RequiresOverrideImpl() {
			usedTypes[f.Type] = ti
		}
		f.GoType = ti
		return nil
	}
	for idx := range def.Configuration.Fields {
		if err := processField(&def.Configuration.Fields[idx]); err != nil {
			return nil, errors.Trace(err)
		}
	}
	def.ensureRelatedTypeInfo("Configuration", def.Configuration)
	res := templateData{
		PackageName: def.PackageName,
		Structs:     map[string]*structInfo{"Configuration": def.Configuration},
	}
	for tn, td := range def.RelatedTypes {
		for idx := range td.Fields {
			if err := processField(&td.Fields[idx]); err != nil {
				return nil, errors.Trace(err)
			}
		}
		res.Structs[tn] = td
		def.ensureRelatedTypeInfo(tn, td)
	}

	def.populateStructExamples()
	typeList := make([]*typeInfo, 0, len(usedTypes))
	for _, t := range usedTypes {
		typeList = append(typeList, t)
	}
	sort.Sort(typeInfos(typeList))
	res.BaseTypes = typeList
	return &res, nil
}

// populateStructExamples goes through all the struct types that we're going to generate
// and populates the ExampleValues field for them [the ExampleValues are used to generate
// unit tests].
// We can only generate an ExampleValue for a type once there are ExampleValues available
// for all the fields in the Type, so to handle cases where struct A contains a struct B
// we loop over all the types, doing the ones we can, then go back and try them again
// and so on until they're all done.
func (def *configDef) populateStructExamples() {
	for tries := 0; tries <= len(def.customTypeInfos); tries++ {
		completed := true
		for _, t := range def.customTypeInfos {
			if len(t.ExampleValues) == 0 {
				// note that its in this order so that populate gets called regardless of the value of completed
				completed = populateStructExample(t) && completed
			}
		}
		if completed {
			return
		}
	}
	missing := make([]string, 0, 10)
	for _, t := range def.customTypeInfos {
		if len(t.ExampleValues) == 0 {
			missing = append(missing, t.Name)
		}
	}
	log.Fatalf("after %d tries, still unable to generate all the ExampleValues, %v still don't have any", len(def.customTypeInfos), missing)
}

// populateStructExample updates the typeInfo with ExampleValues and returns true
// or returns false if it was unable to [because there are missing example from
// other types]
// the base types get 3 examples, we get 6, 3 are fully populated, and 3 only populate
// a subset of fields
func populateStructExample(t *typeInfo) bool {
	r := &bytes.Buffer{}
	build := func(idx int, maxFields int) (string, bool) {
		r.Reset()
		isArray := strings.HasPrefix(t.Name, "[]")
		if isArray {
			fmt.Fprintf(r, "%s{\n{\n", t.Name)
		} else {
			fmt.Fprintf(r, "%s{\n", t.Name)
		}
		delim := ""
		for fidx, f := range t.structDef.Fields {
			if fidx >= maxFields {
				break
			}
			fe := f.GoType.ExampleValues
			if len(fe) == 0 {
				return "", false
			}
			fmt.Fprintf(r, "%s%s: %s", delim, f.Name, fe[idx])
			delim = ",\n"
		}
		if isArray {
			fmt.Fprint(r, "},\n}")
		} else {
			r.WriteByte('}')
		}
		return r.String(), true
	}
	res := make([]string, 0, 6)
	for i := 0; i < 3; i++ {
		ex, ok := build(i, 5000)
		if !ok {
			return false
		}
		res = append(res, ex)
	}
	for i := 0; i < 3; i++ {
		ex, _ := build(i, i+1)
		res = append(res, ex)
	}
	t.ExampleValues = res
	return true
}

// we need a typeInfo instance for this custom struct type (or slice of custom struct),
// this method will create it [or return a previously created one]
func (def *configDef) ensureRelatedTypeInfo(name string, structDef *structInfo) *typeInfo {
	if existing, exists := def.customTypeInfos[name]; exists {
		return existing
	}
	t := typeInfo{
		Name:          name,
		overrideStyle: osStruct,
		structDef:     structDef,
	}
	if strings.HasPrefix(name, "[]") {
		typ := name[2:]
		t.OverrideFunc = "override" + typ + "Slice"
		t.overrideStyle = osLen
		def.ensureRelatedTypeInfo(typ, structDef)
	} else {
		structDef.GoType = &t
	}
	def.customTypeInfos[name] = &t
	return &t
}

// gofmt runs gofmt on the supplied source code file
func gofmt(f string) {
	fg := exec.Command("gofmt", "-w", "-s", f)
	fg.Run()
}
