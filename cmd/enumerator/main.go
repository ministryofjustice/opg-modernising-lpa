// The code for enumerator is based on golang.org/x/tools/cmd/stringer,
// copyright 2014 The Go Authors. All rights reserved. Use of that source code
// is governed by a BSD-style license that follows:
//
// Copyright (c) 2009 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
//    * Redistributions of source code must retain the above copyright notice,
//      this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above copyright
//      notice, this list of conditions and the following disclaimer in the
//      documentation and/or other materials provided with the distribution.
//    * Neither the name of Google Inc. nor the names of its contributors may be
//      used to endorse or promote products derived from this software without
//      specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

// Enumerator is a tool to automate the creation of simple enums. Given the name
// of a (signed or unsigned) integer type T that has constants defined,
// enumerator will create a new self-contained Go source file implementing
//
//	func ParseT(string) (T, error)
//	func (t T) String() string
//	func (t T) MarshalText() ([]byte, error)
//	func (t *T) UnmarshalText([]byte) error
//
// and for each value X
//
//	func (t T) IsX() bool
//
// The file is created in the same package and directory as the package that
// defines T. This tool is designed to be used with go generate.
//
// For example, given this snippet,
//
//	package painkiller
//
//	type Pill int
//
//	const (
//		Placebo Pill = iota
//		Aspirin
//		Ibuprofen
//		Paracetamol
//		Acetaminophen = Paracetamol
//	)
//
// running this command
//
//	enumerator -type=Pill
//
// in the same directory will create the file enum_pill.go, in package painkiller,
// containing definitions of
//
//	func ParsePill() (Pill, error)
//	func (Pill) String() string
//	func (Pill) MarshalText() ([]byte, error)
//	func (*Pill) UnmarshalText([]byte) error
//	func (Pill) IsPlacebo() bool
//	func (Pill) IsAspirin() bool
//	func (Pill) IsIbuprofen() bool
//	func (Pill) IsParacetamol() bool
//
// The String method will translate the value of a Pill constant to the string
// representation of the respective constant name, so that the call
// fmt.Print(painkiller.Aspirin) will print the string "Aspirin".
//
// The Parse method performs the inverse, so that the call ParsePill("Aspirin")
// will return painkiller.Aspirin, nil.
//
// Typically this process would be run using go generate, like this:
//
//	//go:generate stringer -type=Pill
//
// If multiple constants have the same value, the lexically first matching name will
// be used (in the example, Acetaminophen will print as "Paracetamol").
//
// The -type flag is required to contain the type to generate methods for.
//
// The -linecomment flag tells enumerator to generate the text of any line
// comment, trimmed of leading spaces, instead of the constant name. For
// instance, if it was desired to have the names in lower case
//
//	Aspirin // aspirin
//
// The -trimprefix flag tells enumerator to remove any type name prefixes. For
// instance, if we prefixed our values with Pill, like
//
//	PillAspirin
//
// an IsAspirin() method would still be generated,
// painkiller.PillAspirin.String() would return "Aspirin" and
// ParsePill("Aspirin") would return painkiller.PillAspirin.
//
// The -empty flag tells enumerator to generate a method to check whether the
// underlying value is 0. This is useful when an enum is defined using iota+1.
// For instance, with the following
//
//	type YesNo uint8
//
//	const (
//		Yes YesNo = iota + 1
//		No
//	)
//
// we would be able to check whether a value x had not been set to Yes or No by
// using the boolean value returned by x.Empty(). It also alters Parse to accept
// "" as, in this case, YesNo(0).
//
// The -bits flag tells enumerator to consider the type as a field of bits. This
// is useful when an emum is defined using 1<<iota. It causes the following
// methods to be generated instead of the usual behaviour:
//
//	func ParseT([]string) (T, error)
//	func (t T) String() string
//	func (t T) Strings() []string
//
// and for each value X
//
//	func (t T) HasX() bool
//
// The String() method will only return sensible values for uncombined values of
// T, i.e. X.String(), not (X|Y).String(). When dealing with combined values use
// Strings(), where (X|Y).Strings() == []string{X.String(), Y.String()}.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/constant"
	"go/format"
	"go/token"
	"go/types"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/packages"
)

var (
	typeName    = flag.String("type", "", "type name; must be set")
	trimprefix  = flag.Bool("trimprefix", false, "trim the type prefix from the generated constant names")
	linecomment = flag.Bool("linecomment", false, "use line comment text as printed text when present")
	empty       = flag.Bool("empty", false, "generate method to check for empty value")
	bits        = flag.Bool("bits", false, "consider the type to be a bit field")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of enumerator:\n")
	fmt.Fprintf(os.Stderr, "\tenumerator [flags] -type T\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("enumerator: ")
	flag.Usage = Usage
	flag.Parse()
	if len(*typeName) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	// Parse the package once.
	g := Generator{
		trimPrefix:  *trimprefix,
		lineComment: *linecomment,
		empty:       *empty,
		bits:        *bits,
	}

	g.parsePackage(".")

	// Print the header and package clause.
	g.Printf(`// Code generated by "enumerator %[1]s"; DO NOT EDIT.

package %[2]s

import (
	"strconv"
	"fmt"
)
`, strings.Join(os.Args[1:], " "), g.pkg.name)

	// Run generate for the type.
	g.generate(*typeName)
	// Format the output.
	src := g.format()

	// Write to file.
	outputName := strings.ToLower(fmt.Sprintf("enum_%s.go", *typeName))
	err := os.WriteFile(outputName, src, 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

// Generator holds the state of the analysis. Primarily used to buffer
// the output for format.Source.
type Generator struct {
	buf bytes.Buffer // Accumulated output.
	pkg *Package     // Package we are scanning.

	trimPrefix  bool
	lineComment bool
	empty       bool
	bits        bool
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

// File holds a single parsed file and associated data.
type File struct {
	pkg  *Package  // Package to which this file belongs.
	file *ast.File // Parsed AST.
	// These fields are reset for each type being generated.
	typeName string  // Name of the constant type.
	values   []Value // Accumulator for constant values of that type.

	trimPrefix  bool
	lineComment bool
}

type Package struct {
	name  string
	defs  map[*ast.Ident]types.Object
	files []*File
}

// parsePackage analyzes the single package constructed from the patterns and tags.
// parsePackage exits if there is an error.
func (g *Generator) parsePackage(path string) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		// TODO: Need to think about constants in test files. Maybe write type_string_test.go
		// in a separate pass? For later.
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	g.addPackage(pkgs[0])
}

// addPackage adds a type checked Package and its syntax files to the generator.
func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		name:  pkg.Name,
		defs:  pkg.TypesInfo.Defs,
		files: make([]*File, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		g.pkg.files[i] = &File{
			file:        file,
			pkg:         g.pkg,
			trimPrefix:  g.trimPrefix,
			lineComment: g.lineComment,
		}
	}
}

// generate produces the String method for the named type.
func (g *Generator) generate(typeName string) {
	values := make([]Value, 0, 100)
	for _, file := range g.pkg.files {
		// Set the state for this run of the walker.
		file.typeName = typeName
		file.values = nil
		if file.file != nil {
			ast.Inspect(file.file, file.genDecl)
			values = append(values, file.values...)
		}
	}

	if len(values) == 0 {
		log.Fatalf("no values defined for type %s", typeName)
	}
	// Generate code that will fail if the constants change value.
	g.Printf("func _() {\n")
	g.Printf("\t// An \"invalid array index\" compiler error signifies that the constant values have changed.\n")
	g.Printf("\t// Re-run the stringer command to generate them again.\n")
	g.Printf("\tvar x [1]struct{}\n")
	for _, v := range values {
		g.Printf("\t_ = x[%s - %s]\n", v.originalName, v.str)
	}
	g.Printf("}\n")

	runs := splitIntoRuns(values)
	// The decision of which pattern to use depends on the number of
	// runs in the numbers. If there's only one, it's easy. For more than
	// one, there's a tradeoff between complexity and size of the data
	// and code vs. the simplicity of a map. A map takes more space,
	// but so does the code. The decision here (crossover at 10) is
	// arbitrary, but considers that for large numbers of runs the cost
	// of the linear scan in the switch might become important, and
	// rather than use yet another algorithm such as binary search,
	// we punt and use a map. In any case, the likelihood of a map
	// being necessary for any realistic example other than bitmasks
	// is very low. And bitmasks probably deserve their own analysis,
	// to be done some other day.
	switch {
	case len(runs) == 1:
		g.buildOneRun(runs, typeName)
	case len(runs) <= 10:
		g.buildMultipleRuns(runs, typeName)
	default:
		g.buildMap(runs, typeName)
	}
	if g.bits {
		g.buildStrings(runs, typeName)
	}

	g.buildTextMethods(typeName)
	g.buildIsMethods(runs, typeName)
	g.buildParseMethod(runs, typeName)
	g.buildValues(runs, typeName)
	if g.empty {
		g.buildEmpty(typeName)
	}
}

// splitIntoRuns breaks the values into runs of contiguous sequences.
// For example, given 1,2,3,5,6,7 it returns {1,2,3},{5,6,7}.
// The input slice is known to be non-empty.
func splitIntoRuns(values []Value) [][]Value {
	// We use stable sort so the lexically first name is chosen for equal elements.
	sort.Stable(byValue(values))
	// Remove duplicates. Stable sort has put the one we want to print first,
	// so use that one. The String method won't care about which named constant
	// was the argument, so the first name for the given value is the only one to keep.
	// We need to do this because identical values would cause the switch or map
	// to fail to compile.
	j := 1
	for i := 1; i < len(values); i++ {
		if values[i].value != values[i-1].value {
			values[j] = values[i]
			j++
		}
	}
	values = values[:j]
	runs := make([][]Value, 0, 10)
	for len(values) > 0 {
		// One contiguous sequence per outer loop.
		i := 1
		for i < len(values) && values[i].value == values[i-1].value+1 {
			i++
		}
		runs = append(runs, values[:i])
		values = values[i:]
	}
	return runs
}

// format returns the gofmt-ed contents of the Generator's buffer.
func (g *Generator) format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}

// Value represents a declared constant.
type Value struct {
	originalName string // The name of the constant.
	name         string // The name with trimmed prefix.
	// The value is stored as a bit pattern alone. The boolean tells us
	// whether to interpret it as an int64 or a uint64; the only place
	// this matters is when sorting.
	// Much of the time the str field is all we need; it is printed
	// by Value.String.
	value  uint64 // Will be converted to int64 when needed.
	signed bool   // Whether the constant is a signed type.
	str    string // The string representation given by the "go/constant" package.
}

func (v *Value) String() string {
	return v.str
}

// byValue lets us sort the constants into increasing order.
// We take care in the Less method to sort in signed or unsigned order,
// as appropriate.
type byValue []Value

func (b byValue) Len() int      { return len(b) }
func (b byValue) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byValue) Less(i, j int) bool {
	if b[i].signed {
		return int64(b[i].value) < int64(b[j].value)
	}
	return b[i].value < b[j].value
}

// genDecl processes one declaration clause.
func (f *File) genDecl(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.CONST {
		// We only care about const declarations.
		return true
	}
	// The name of the type of the constants we are declaring.
	// Can change if this is a multi-element declaration.
	typ := ""
	// Loop over the elements of the declaration. Each element is a ValueSpec:
	// a list of names possibly followed by a type, possibly followed by values.
	// If the type and value are both missing, we carry down the type (and value,
	// but the "go/types" package takes care of that).
	for _, spec := range decl.Specs {
		vspec := spec.(*ast.ValueSpec) // Guaranteed to succeed as this is CONST.
		if vspec.Type == nil && len(vspec.Values) > 0 {
			// "X = 1". With no type but a value. If the constant is untyped,
			// skip this vspec and reset the remembered type.
			typ = ""

			// If this is a simple type conversion, remember the type.
			// We don't mind if this is actually a call; a qualified call won't
			// be matched (that will be SelectorExpr, not Ident), and only unusual
			// situations will result in a function call that appears to be
			// a type conversion.
			ce, ok := vspec.Values[0].(*ast.CallExpr)
			if !ok {
				continue
			}
			id, ok := ce.Fun.(*ast.Ident)
			if !ok {
				continue
			}
			typ = id.Name
		}
		if vspec.Type != nil {
			// "X T". We have a type. Remember it.
			ident, ok := vspec.Type.(*ast.Ident)
			if !ok {
				continue
			}
			typ = ident.Name
		}
		if typ != f.typeName {
			// This is not the type we're looking for.
			continue
		}
		// We now have a list of names (from one line of source code) all being
		// declared with the desired type.
		// Grab their names and actual values and store them in f.values.
		for _, name := range vspec.Names {
			if name.Name == "_" {
				continue
			}
			// This dance lets the type checker find the values for us. It's a
			// bit tricky: look up the object declared by the name, find its
			// types.Const, and extract its value.
			obj, ok := f.pkg.defs[name]
			if !ok {
				log.Fatalf("no value for constant %s", name)
			}
			info := obj.Type().Underlying().(*types.Basic).Info()
			if info&types.IsInteger == 0 {
				log.Fatalf("can't handle non-integer constant type %s", typ)
			}
			value := obj.(*types.Const).Val() // Guaranteed to succeed as this is CONST.
			if value.Kind() != constant.Int {
				log.Fatalf("can't happen: constant is not an integer %s", name)
			}
			i64, isInt := constant.Int64Val(value)
			u64, isUint := constant.Uint64Val(value)
			if !isInt && !isUint {
				log.Fatalf("internal error: value of %s is not an integer: %s", name, value.String())
			}
			if !isInt {
				u64 = uint64(i64)
			}
			v := Value{
				originalName: name.Name,
				value:        u64,
				signed:       info&types.IsUnsigned == 0,
				str:          value.String(),
			}
			if c := vspec.Comment; f.lineComment && c != nil && len(c.List) == 1 {
				v.name = strings.TrimSpace(c.Text())
			} else {
				v.name = strings.TrimPrefix(v.originalName, f.typeName)
			}
			f.values = append(f.values, v)
		}
	}
	return false
}

// Helpers

// usize returns the number of bits of the smallest unsigned integer
// type that will hold n. Used to create the smallest possible slice of
// integers to use as indexes into the concatenated strings.
func usize(n int) int {
	switch {
	case n < 1<<8:
		return 8
	case n < 1<<16:
		return 16
	default:
		// 2^32 is enough constants for anyone.
		return 32
	}
}

// declareIndexAndNameVars declares the index slices and concatenated names
// strings representing the runs of values.
func (g *Generator) declareIndexAndNameVars(runs [][]Value, typeName string) {
	var indexes, names []string
	for i, run := range runs {
		index, name := g.createIndexAndNameDecl(run, typeName, fmt.Sprintf("_%d", i))
		if len(run) != 1 {
			indexes = append(indexes, index)
		}
		names = append(names, name)
	}
	g.Printf("const (\n")
	for _, name := range names {
		g.Printf("\t%s\n", name)
	}
	g.Printf(")\n\n")

	if len(indexes) > 0 {
		g.Printf("var (")
		for _, index := range indexes {
			g.Printf("\t%s\n", index)
		}
		g.Printf(")\n\n")
	}
}

// declareIndexAndNameVar is the single-run version of declareIndexAndNameVars
func (g *Generator) declareIndexAndNameVar(run []Value, typeName string) {
	index, name := g.createIndexAndNameDecl(run, typeName, "")
	g.Printf("const %s\n", name)
	g.Printf("var %s\n", index)
}

// declareNameVars declares the concatenated names string representing all the values in the runs.
func (g *Generator) declareNameVars(runs [][]Value, typeName string, suffix string) {
	g.Printf("const _%s_name%s = \"", typeName, suffix)
	for _, run := range runs {
		for i := range run {
			g.Printf("%s", run[i].name)
		}
	}
	g.Printf("\"\n")
}

// createIndexAndNameDecl returns the pair of declarations for the run. The caller will add "const" and "var".
func (g *Generator) createIndexAndNameDecl(run []Value, typeName string, suffix string) (string, string) {
	b := new(bytes.Buffer)
	indexes := make([]int, len(run))
	for i := range run {
		b.WriteString(run[i].name)
		indexes[i] = b.Len()
	}
	nameConst := fmt.Sprintf("_%s_name%s = %q", typeName, suffix, b.String())
	nameLen := b.Len()
	b.Reset()
	fmt.Fprintf(b, "_%s_index%s = [...]uint%d{0, ", typeName, suffix, usize(nameLen))
	for i, v := range indexes {
		if i > 0 {
			fmt.Fprintf(b, ", ")
		}
		fmt.Fprintf(b, "%d", v)
	}
	fmt.Fprintf(b, "}")
	return b.String(), nameConst
}

func (g *Generator) buildTextMethods(typeName string) {
	if !g.bits {
		g.Printf(textMethods, typeName, upperFirst(typeName))
	}
}

const textMethods = `
func (i %[1]s) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *%[1]s) UnmarshalText(text []byte) error {
  val, err := Parse%[2]s(string(text))
  if err != nil {
    return err
  }

  *i = val
  return nil
}
`

func (g *Generator) buildIsMethods(runs [][]Value, typeName string) {
	for _, values := range runs {
		for _, value := range values {
			name := value.originalName
			if g.trimPrefix {
				name = strings.TrimPrefix(name, typeName)
			}
			if g.bits {
				g.Printf(hasMethod, typeName, name, value.originalName)
			} else {
				g.Printf(isMethod, typeName, name, value.originalName)
			}
		}
	}
}

const hasMethod = `
func (i %[1]s) Has%[2]s() bool {
	return i&%[3]s != 0
}
`

const isMethod = `
func (i %[1]s) Is%[2]s() bool {
	return i == %[3]s
}
`

func (g *Generator) buildParseMethod(runs [][]Value, typeName string) {
	g.Printf("\n")

	if g.bits {
		g.Printf(`func Parse%[1]s(strs []string) (%[2]s, error) {
  var result %[2]s

  for _, s := range strs {
	  switch s {
`, upperFirst(typeName), typeName)
		for _, values := range runs {
			for _, value := range values {
				g.Printf(`  case "%[1]s":
		result |= %[2]s
`, value.name, value.originalName)
			}
		}
		g.Printf(`  default:
		  return %[1]s(0), fmt.Errorf("invalid %[1]s '%%s'", s)
	  }
  }

  return result, nil
}
`, typeName)
	} else {
		g.Printf(`func Parse%[1]s(s string) (%[2]s, error) {
	switch s {
`, upperFirst(typeName), typeName)
		if g.empty {
			g.Printf(`  case "":
		return %[1]s(0), nil
`, typeName)
		}
		for _, values := range runs {
			for _, value := range values {
				g.Printf(`  case "%[1]s":
		return %[2]s, nil
`, value.name, value.originalName)
			}
		}
		g.Printf(`  default:
		return %[1]s(0), fmt.Errorf("invalid %[1]s '%%s'", s)
	}
}
`, typeName)
	}
}

func (g *Generator) buildValues(runs [][]Value, typeName string) {
	g.Printf("\ntype %[1]sOptions struct {\n", typeName)
	for _, values := range runs {
		for _, value := range values {
			name := value.originalName
			if g.trimPrefix {
				name = strings.TrimPrefix(value.originalName, typeName)
			}

			g.Printf("%[1]s %[2]s\n", name, typeName)
		}
	}
	g.Printf("\n}\n")

	g.Printf("var %[1]sValues = %[1]sOptions{\n", typeName)
	for _, values := range runs {
		for _, value := range values {
			name := value.originalName
			if g.trimPrefix {
				name = strings.TrimPrefix(value.originalName, typeName)
			}

			g.Printf("%[1]s: %[2]s,\n", name, value.originalName)
		}
	}
	g.Printf("\n}\n")
}

func (g *Generator) buildEmpty(typeName string) {
	g.Printf(`
func (i %[1]s) Empty() bool {
	return i == %[1]s(0)
}
`, typeName)
}

// buildString generates the variables and String method for a single run of contiguous values.
func (g *Generator) buildOneRun(runs [][]Value, typeName string) {
	values := runs[0]

	g.Printf("\n")
	g.declareIndexAndNameVar(values, typeName)
	// The generated code is simple enough to write as a Printf format.
	lessThanZero := ""
	if values[0].signed {
		lessThanZero = "i < 0 || "
	}
	emptyEarlyReturn := ""
	if g.empty {
		emptyEarlyReturn = "if i == 0 {\n     return \"\"\n}\n"
	}

	if values[0].value == 0 { // Signed or unsigned, 0 is still 0.
		g.Printf(stringOneRun, typeName, usize(len(values)), lessThanZero, emptyEarlyReturn)
	} else {
		g.Printf(stringOneRunWithOffset, typeName, values[0].String(), usize(len(values)), lessThanZero, emptyEarlyReturn)
	}
}

// Arguments to format are:
//
//	[1]: type name
//	[2]: size of index element (8 for uint8 etc.)
//	[3]: less than zero check (for signed types)
//	[4]: early return when empty set and empty string
const stringOneRun = `func (i %[1]s) String() string {
	%[4]s if %[3]si >= %[1]s(len(_%[1]s_index)-1) {
		return "%[1]s(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _%[1]s_name[_%[1]s_index[i]:_%[1]s_index[i+1]]
}
`

// Arguments to format are:
//
//	[1]: type name
//	[2]: lowest defined value for type, as a string
//	[3]: size of index element (8 for uint8 etc.)
//	[4]: less than zero check (for signed types)
//	[5]: early return when empty set and empty string
const stringOneRunWithOffset = `func (i %[1]s) String() string {
	%[5]s i -= %[2]s
	if %[4]si >= %[1]s(len(_%[1]s_index)-1) {
		return "%[1]s(" + strconv.FormatInt(int64(i + %[2]s), 10) + ")"
	}
	return _%[1]s_name[_%[1]s_index[i] : _%[1]s_index[i+1]]
}
`

// buildMultipleRuns generates the variables and String method for multiple runs of contiguous values.
// For this pattern, a single Printf format won't do.
func (g *Generator) buildMultipleRuns(runs [][]Value, typeName string) {
	g.Printf("\n")
	g.declareIndexAndNameVars(runs, typeName)
	g.Printf("func (i %s) String() string {\n", typeName)
	g.Printf("\tswitch {\n")
	for i, values := range runs {
		if len(values) == 1 {
			g.Printf("\tcase i == %s:\n", &values[0])
			g.Printf("\t\treturn _%s_name_%d\n", typeName, i)
			continue
		}
		if values[0].value == 0 && !values[0].signed {
			// For an unsigned lower bound of 0, "0 <= i" would be redundant.
			g.Printf("\tcase i <= %s:\n", &values[len(values)-1])
		} else {
			g.Printf("\tcase %s <= i && i <= %s:\n", &values[0], &values[len(values)-1])
		}
		if values[0].value != 0 {
			g.Printf("\t\ti -= %s\n", &values[0])
		}
		g.Printf("\t\treturn _%s_name_%d[_%s_index_%d[i]:_%s_index_%d[i+1]]\n",
			typeName, i, typeName, i, typeName, i)
	}
	g.Printf("\tdefault:\n")
	g.Printf("\t\treturn \"%s(\" + strconv.FormatInt(int64(i), 10) + \")\"\n", typeName)
	g.Printf("\t}\n")
	g.Printf("}\n")
}

// buildMap handles the case where the space is so sparse a map is a reasonable fallback.
// It's a rare situation but has simple code.
func (g *Generator) buildMap(runs [][]Value, typeName string) {
	g.Printf("\n")
	g.declareNameVars(runs, typeName, "")
	g.Printf("\nvar _%s_map = map[%s]string{\n", typeName, typeName)
	n := 0
	for _, values := range runs {
		for _, value := range values {
			g.Printf("\t%s: _%s_name[%d:%d],\n", &value, typeName, n, n+len(value.name))
			n += len(value.name)
		}
	}
	g.Printf("}\n\n")
	g.Printf(stringMap, typeName)
}

// Argument to format is the type name.
const stringMap = `func (i %[1]s) String() string {
	if str, ok := _%[1]s_map[i]; ok {
		return str
	}
	return "%[1]s(" + strconv.FormatInt(int64(i), 10) + ")"
}
`

func (g *Generator) buildStrings(runs [][]Value, typeName string) {
	g.Printf("\n")

	g.Printf(`func (i %[1]s) Strings() []string {
  var result []string
`, typeName)
	for _, values := range runs {
		for _, value := range values {
			g.Printf(`if i.Has%[1]s() {
  result = append(result, %[2]s.String())
}
`, value.name, value.originalName)
		}
	}
	g.Printf(`  return result
}
`)
}

func upperFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}
