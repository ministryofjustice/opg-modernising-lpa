// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains simple golden tests for various examples.
// Besides validating the results when the implementation changes,
// it provides a way to look at the generated code without having
// to execute the print statements in one's head.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Golden represents a test case.
type Golden struct {
	name        string
	trimPrefix  bool
	lineComment bool
	empty       bool
	bits        bool
	input       string // input; the package clause is provided when running the test.
	output      string // expected output.
}

var golden = []Golden{
	{"day", false, false, true, false, day_in, day_out},
	{"offset", false, false, false, false, offset_in, offset_out},
	{"gap", false, false, false, false, gap_in, gap_out},
	{"num", false, false, false, false, num_in, num_out},
	{"unum", false, false, false, false, unum_in, unum_out},
	{"unumpos", false, false, false, false, unumpos_in, unumpos_out},
	{"prime", false, false, false, false, prime_in, prime_out},
	{"prefix", true, false, false, false, prefix_in, prefix_out},
	{"tokens", false, true, false, false, tokens_in, tokens_out},
	{"bits", false, true, false, true, bits_in, bits_out},
}

// Each example starts with "type XXX [u]int", with a single space separating them.

// Simple test: enumeration of type int starting at 0.
const day_in = `type Day int
const (
	Monday Day = iota + 1
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
)
`

const day_out = `func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Monday-1]
	_ = x[Tuesday-2]
	_ = x[Wednesday-3]
	_ = x[Thursday-4]
	_ = x[Friday-5]
	_ = x[Saturday-6]
	_ = x[Sunday-7]
}

const _Day_name = "MondayTuesdayWednesdayThursdayFridaySaturdaySunday"

var _Day_index = [...]uint8{0, 6, 13, 22, 30, 36, 44, 50}

func (i Day) String() string {
	if i == 0 {
		return ""
	}
	i -= 1
	if i < 0 || i >= Day(len(_Day_index)-1) {
		return "Day(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _Day_name[_Day_index[i]:_Day_index[i+1]]
}

func (i Day) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *Day) UnmarshalText(text []byte) error {
	val, err := ParseDay(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i Day) IsMonday() bool {
	return i == Monday
}

func (i Day) IsTuesday() bool {
	return i == Tuesday
}

func (i Day) IsWednesday() bool {
	return i == Wednesday
}

func (i Day) IsThursday() bool {
	return i == Thursday
}

func (i Day) IsFriday() bool {
	return i == Friday
}

func (i Day) IsSaturday() bool {
	return i == Saturday
}

func (i Day) IsSunday() bool {
	return i == Sunday
}

func ParseDay(s string) (Day, error) {
	switch s {
	case "":
		return Day(0), nil
	case "Monday":
		return Monday, nil
	case "Tuesday":
		return Tuesday, nil
	case "Wednesday":
		return Wednesday, nil
	case "Thursday":
		return Thursday, nil
	case "Friday":
		return Friday, nil
	case "Saturday":
		return Saturday, nil
	case "Sunday":
		return Sunday, nil
	default:
		return Day(0), fmt.Errorf("invalid Day '%s'", s)
	}
}

type DayOptions struct {
	Monday    Day
	Tuesday   Day
	Wednesday Day
	Thursday  Day
	Friday    Day
	Saturday  Day
	Sunday    Day
}

var DayValues = DayOptions{
	Monday:    Monday,
	Tuesday:   Tuesday,
	Wednesday: Wednesday,
	Thursday:  Thursday,
	Friday:    Friday,
	Saturday:  Saturday,
	Sunday:    Sunday,
}

func (i Day) Empty() bool {
	return i == Day(0)
}
`

// Enumeration with an offset.
// Also includes a duplicate.
const offset_in = `type Number int
const (
	_ Number = iota
	One
	Two
	Three
	AnotherOne = One  // Duplicate; note that AnotherOne doesn't appear below.
)
`

const offset_out = `func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[One-1]
	_ = x[Two-2]
	_ = x[Three-3]
}

const _Number_name = "OneTwoThree"

var _Number_index = [...]uint8{0, 3, 6, 11}

func (i Number) String() string {
	i -= 1
	if i < 0 || i >= Number(len(_Number_index)-1) {
		return "Number(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _Number_name[_Number_index[i]:_Number_index[i+1]]
}

func (i Number) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *Number) UnmarshalText(text []byte) error {
	val, err := ParseNumber(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i Number) IsOne() bool {
	return i == One
}

func (i Number) IsTwo() bool {
	return i == Two
}

func (i Number) IsThree() bool {
	return i == Three
}

func ParseNumber(s string) (Number, error) {
	switch s {
	case "One":
		return One, nil
	case "Two":
		return Two, nil
	case "Three":
		return Three, nil
	default:
		return Number(0), fmt.Errorf("invalid Number '%s'", s)
	}
}

type NumberOptions struct {
	One   Number
	Two   Number
	Three Number
}

var NumberValues = NumberOptions{
	One:   One,
	Two:   Two,
	Three: Three,
}
`

// Gaps and an offset.
const gap_in = `type Gap int
const (
	Two Gap = 2
	Three Gap = 3
	Five Gap = 5
	Six Gap = 6
	Seven Gap = 7
	Eight Gap = 8
	Nine Gap = 9
	Eleven Gap = 11
)
`

const gap_out = `func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Two-2]
	_ = x[Three-3]
	_ = x[Five-5]
	_ = x[Six-6]
	_ = x[Seven-7]
	_ = x[Eight-8]
	_ = x[Nine-9]
	_ = x[Eleven-11]
}

const (
	_Gap_name_0 = "TwoThree"
	_Gap_name_1 = "FiveSixSevenEightNine"
	_Gap_name_2 = "Eleven"
)

var (
	_Gap_index_0 = [...]uint8{0, 3, 8}
	_Gap_index_1 = [...]uint8{0, 4, 7, 12, 17, 21}
)

func (i Gap) String() string {
	switch {
	case 2 <= i && i <= 3:
		i -= 2
		return _Gap_name_0[_Gap_index_0[i]:_Gap_index_0[i+1]]
	case 5 <= i && i <= 9:
		i -= 5
		return _Gap_name_1[_Gap_index_1[i]:_Gap_index_1[i+1]]
	case i == 11:
		return _Gap_name_2
	default:
		return "Gap(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}

func (i Gap) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *Gap) UnmarshalText(text []byte) error {
	val, err := ParseGap(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i Gap) IsTwo() bool {
	return i == Two
}

func (i Gap) IsThree() bool {
	return i == Three
}

func (i Gap) IsFive() bool {
	return i == Five
}

func (i Gap) IsSix() bool {
	return i == Six
}

func (i Gap) IsSeven() bool {
	return i == Seven
}

func (i Gap) IsEight() bool {
	return i == Eight
}

func (i Gap) IsNine() bool {
	return i == Nine
}

func (i Gap) IsEleven() bool {
	return i == Eleven
}

func ParseGap(s string) (Gap, error) {
	switch s {
	case "Two":
		return Two, nil
	case "Three":
		return Three, nil
	case "Five":
		return Five, nil
	case "Six":
		return Six, nil
	case "Seven":
		return Seven, nil
	case "Eight":
		return Eight, nil
	case "Nine":
		return Nine, nil
	case "Eleven":
		return Eleven, nil
	default:
		return Gap(0), fmt.Errorf("invalid Gap '%s'", s)
	}
}

type GapOptions struct {
	Two    Gap
	Three  Gap
	Five   Gap
	Six    Gap
	Seven  Gap
	Eight  Gap
	Nine   Gap
	Eleven Gap
}

var GapValues = GapOptions{
	Two:    Two,
	Three:  Three,
	Five:   Five,
	Six:    Six,
	Seven:  Seven,
	Eight:  Eight,
	Nine:   Nine,
	Eleven: Eleven,
}
`

// Signed integers spanning zero.
const num_in = `type Num int
const (
	m_2 Num = -2 + iota
	m_1
	m0
	m1
	m2
)
`

const num_out = `func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[m_2 - -2]
	_ = x[m_1 - -1]
	_ = x[m0-0]
	_ = x[m1-1]
	_ = x[m2-2]
}

const _Num_name = "m_2m_1m0m1m2"

var _Num_index = [...]uint8{0, 3, 6, 8, 10, 12}

func (i Num) String() string {
	i -= -2
	if i < 0 || i >= Num(len(_Num_index)-1) {
		return "Num(" + strconv.FormatInt(int64(i+-2), 10) + ")"
	}
	return _Num_name[_Num_index[i]:_Num_index[i+1]]
}

func (i Num) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *Num) UnmarshalText(text []byte) error {
	val, err := ParseNum(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i Num) Ism_2() bool {
	return i == m_2
}

func (i Num) Ism_1() bool {
	return i == m_1
}

func (i Num) Ism0() bool {
	return i == m0
}

func (i Num) Ism1() bool {
	return i == m1
}

func (i Num) Ism2() bool {
	return i == m2
}

func ParseNum(s string) (Num, error) {
	switch s {
	case "m_2":
		return m_2, nil
	case "m_1":
		return m_1, nil
	case "m0":
		return m0, nil
	case "m1":
		return m1, nil
	case "m2":
		return m2, nil
	default:
		return Num(0), fmt.Errorf("invalid Num '%s'", s)
	}
}

type NumOptions struct {
	m_2 Num
	m_1 Num
	m0  Num
	m1  Num
	m2  Num
}

var NumValues = NumOptions{
	m_2: m_2,
	m_1: m_1,
	m0:  m0,
	m1:  m1,
	m2:  m2,
}
`

// Unsigned integers spanning zero.
const unum_in = `type Unum uint
const (
	m_2 Unum = iota + 253
	m_1
)

const (
	m0 Unum = iota
	m1
	m2
)
`

const unum_out = `func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[m_2-253]
	_ = x[m_1-254]
	_ = x[m0-0]
	_ = x[m1-1]
	_ = x[m2-2]
}

const (
	_Unum_name_0 = "m0m1m2"
	_Unum_name_1 = "m_2m_1"
)

var (
	_Unum_index_0 = [...]uint8{0, 2, 4, 6}
	_Unum_index_1 = [...]uint8{0, 3, 6}
)

func (i Unum) String() string {
	switch {
	case i <= 2:
		return _Unum_name_0[_Unum_index_0[i]:_Unum_index_0[i+1]]
	case 253 <= i && i <= 254:
		i -= 253
		return _Unum_name_1[_Unum_index_1[i]:_Unum_index_1[i+1]]
	default:
		return "Unum(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}

func (i Unum) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *Unum) UnmarshalText(text []byte) error {
	val, err := ParseUnum(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i Unum) Ism0() bool {
	return i == m0
}

func (i Unum) Ism1() bool {
	return i == m1
}

func (i Unum) Ism2() bool {
	return i == m2
}

func (i Unum) Ism_2() bool {
	return i == m_2
}

func (i Unum) Ism_1() bool {
	return i == m_1
}

func ParseUnum(s string) (Unum, error) {
	switch s {
	case "m0":
		return m0, nil
	case "m1":
		return m1, nil
	case "m2":
		return m2, nil
	case "m_2":
		return m_2, nil
	case "m_1":
		return m_1, nil
	default:
		return Unum(0), fmt.Errorf("invalid Unum '%s'", s)
	}
}

type UnumOptions struct {
	m0  Unum
	m1  Unum
	m2  Unum
	m_2 Unum
	m_1 Unum
}

var UnumValues = UnumOptions{
	m0:  m0,
	m1:  m1,
	m2:  m2,
	m_2: m_2,
	m_1: m_1,
}
`

// Unsigned positive integers.
const unumpos_in = `type Unumpos uint
const (
	m253 Unumpos = iota + 253
	m254
)

const (
	m1 Unumpos = iota + 1
	m2
	m3
)
`

const unumpos_out = `func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[m253-253]
	_ = x[m254-254]
	_ = x[m1-1]
	_ = x[m2-2]
	_ = x[m3-3]
}

const (
	_Unumpos_name_0 = "m1m2m3"
	_Unumpos_name_1 = "m253m254"
)

var (
	_Unumpos_index_0 = [...]uint8{0, 2, 4, 6}
	_Unumpos_index_1 = [...]uint8{0, 4, 8}
)

func (i Unumpos) String() string {
	switch {
	case 1 <= i && i <= 3:
		i -= 1
		return _Unumpos_name_0[_Unumpos_index_0[i]:_Unumpos_index_0[i+1]]
	case 253 <= i && i <= 254:
		i -= 253
		return _Unumpos_name_1[_Unumpos_index_1[i]:_Unumpos_index_1[i+1]]
	default:
		return "Unumpos(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}

func (i Unumpos) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *Unumpos) UnmarshalText(text []byte) error {
	val, err := ParseUnumpos(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i Unumpos) Ism1() bool {
	return i == m1
}

func (i Unumpos) Ism2() bool {
	return i == m2
}

func (i Unumpos) Ism3() bool {
	return i == m3
}

func (i Unumpos) Ism253() bool {
	return i == m253
}

func (i Unumpos) Ism254() bool {
	return i == m254
}

func ParseUnumpos(s string) (Unumpos, error) {
	switch s {
	case "m1":
		return m1, nil
	case "m2":
		return m2, nil
	case "m3":
		return m3, nil
	case "m253":
		return m253, nil
	case "m254":
		return m254, nil
	default:
		return Unumpos(0), fmt.Errorf("invalid Unumpos '%s'", s)
	}
}

type UnumposOptions struct {
	m1   Unumpos
	m2   Unumpos
	m3   Unumpos
	m253 Unumpos
	m254 Unumpos
}

var UnumposValues = UnumposOptions{
	m1:   m1,
	m2:   m2,
	m3:   m3,
	m253: m253,
	m254: m254,
}
`

// Enough gaps to trigger a map implementation of the method.
// Also includes a duplicate to test that it doesn't cause problems
const prime_in = `type Prime int
const (
	p2 Prime = 2
	p3 Prime = 3
	p5 Prime = 5
	p7 Prime = 7
	p77 Prime = 7 // Duplicate; note that p77 doesn't appear below.
	p11 Prime = 11
	p13 Prime = 13
	p17 Prime = 17
	p19 Prime = 19
	p23 Prime = 23
	p29 Prime = 29
	p37 Prime = 31
	p41 Prime = 41
	p43 Prime = 43
)
`

const prime_out = `func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[p2-2]
	_ = x[p3-3]
	_ = x[p5-5]
	_ = x[p7-7]
	_ = x[p77-7]
	_ = x[p11-11]
	_ = x[p13-13]
	_ = x[p17-17]
	_ = x[p19-19]
	_ = x[p23-23]
	_ = x[p29-29]
	_ = x[p37-31]
	_ = x[p41-41]
	_ = x[p43-43]
}

const _Prime_name = "p2p3p5p7p11p13p17p19p23p29p37p41p43"

var _Prime_map = map[Prime]string{
	2:  _Prime_name[0:2],
	3:  _Prime_name[2:4],
	5:  _Prime_name[4:6],
	7:  _Prime_name[6:8],
	11: _Prime_name[8:11],
	13: _Prime_name[11:14],
	17: _Prime_name[14:17],
	19: _Prime_name[17:20],
	23: _Prime_name[20:23],
	29: _Prime_name[23:26],
	31: _Prime_name[26:29],
	41: _Prime_name[29:32],
	43: _Prime_name[32:35],
}

func (i Prime) String() string {
	if str, ok := _Prime_map[i]; ok {
		return str
	}
	return "Prime(" + strconv.FormatInt(int64(i), 10) + ")"
}

func (i Prime) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *Prime) UnmarshalText(text []byte) error {
	val, err := ParsePrime(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i Prime) Isp2() bool {
	return i == p2
}

func (i Prime) Isp3() bool {
	return i == p3
}

func (i Prime) Isp5() bool {
	return i == p5
}

func (i Prime) Isp7() bool {
	return i == p7
}

func (i Prime) Isp11() bool {
	return i == p11
}

func (i Prime) Isp13() bool {
	return i == p13
}

func (i Prime) Isp17() bool {
	return i == p17
}

func (i Prime) Isp19() bool {
	return i == p19
}

func (i Prime) Isp23() bool {
	return i == p23
}

func (i Prime) Isp29() bool {
	return i == p29
}

func (i Prime) Isp37() bool {
	return i == p37
}

func (i Prime) Isp41() bool {
	return i == p41
}

func (i Prime) Isp43() bool {
	return i == p43
}

func ParsePrime(s string) (Prime, error) {
	switch s {
	case "p2":
		return p2, nil
	case "p3":
		return p3, nil
	case "p5":
		return p5, nil
	case "p7":
		return p7, nil
	case "p11":
		return p11, nil
	case "p13":
		return p13, nil
	case "p17":
		return p17, nil
	case "p19":
		return p19, nil
	case "p23":
		return p23, nil
	case "p29":
		return p29, nil
	case "p37":
		return p37, nil
	case "p41":
		return p41, nil
	case "p43":
		return p43, nil
	default:
		return Prime(0), fmt.Errorf("invalid Prime '%s'", s)
	}
}

type PrimeOptions struct {
	p2  Prime
	p3  Prime
	p5  Prime
	p7  Prime
	p11 Prime
	p13 Prime
	p17 Prime
	p19 Prime
	p23 Prime
	p29 Prime
	p37 Prime
	p41 Prime
	p43 Prime
}

var PrimeValues = PrimeOptions{
	p2:  p2,
	p3:  p3,
	p5:  p5,
	p7:  p7,
	p11: p11,
	p13: p13,
	p17: p17,
	p19: p19,
	p23: p23,
	p29: p29,
	p37: p37,
	p41: p41,
	p43: p43,
}
`

const prefix_in = `type Type int
const (
	TypeInt Type = iota
	TypeString
	TypeFloat
	TypeRune
	TypeByte
	TypeStruct
	TypeSlice
)
`

const prefix_out = `func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TypeInt-0]
	_ = x[TypeString-1]
	_ = x[TypeFloat-2]
	_ = x[TypeRune-3]
	_ = x[TypeByte-4]
	_ = x[TypeStruct-5]
	_ = x[TypeSlice-6]
}

const _Type_name = "IntStringFloatRuneByteStructSlice"

var _Type_index = [...]uint8{0, 3, 9, 14, 18, 22, 28, 33}

func (i Type) String() string {
	if i < 0 || i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}

func (i Type) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *Type) UnmarshalText(text []byte) error {
	val, err := ParseType(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i Type) IsInt() bool {
	return i == TypeInt
}

func (i Type) IsString() bool {
	return i == TypeString
}

func (i Type) IsFloat() bool {
	return i == TypeFloat
}

func (i Type) IsRune() bool {
	return i == TypeRune
}

func (i Type) IsByte() bool {
	return i == TypeByte
}

func (i Type) IsStruct() bool {
	return i == TypeStruct
}

func (i Type) IsSlice() bool {
	return i == TypeSlice
}

func ParseType(s string) (Type, error) {
	switch s {
	case "Int":
		return TypeInt, nil
	case "String":
		return TypeString, nil
	case "Float":
		return TypeFloat, nil
	case "Rune":
		return TypeRune, nil
	case "Byte":
		return TypeByte, nil
	case "Struct":
		return TypeStruct, nil
	case "Slice":
		return TypeSlice, nil
	default:
		return Type(0), fmt.Errorf("invalid Type '%s'", s)
	}
}

type TypeOptions struct {
	Int    Type
	String Type
	Float  Type
	Rune   Type
	Byte   Type
	Struct Type
	Slice  Type
}

var TypeValues = TypeOptions{
	Int:    TypeInt,
	String: TypeString,
	Float:  TypeFloat,
	Rune:   TypeRune,
	Byte:   TypeByte,
	Struct: TypeStruct,
	Slice:  TypeSlice,
}
`

const tokens_in = `type Token int
const (
	And Token = iota // &
	Or               // |
	Add              // +
	Sub              // -
	Ident
	Period // .

	// not to be used
	SingleBefore
	// not to be used
	BeforeAndInline // inline
	InlineGeneral /* inline general */
)
`

const tokens_out = `func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[And-0]
	_ = x[Or-1]
	_ = x[Add-2]
	_ = x[Sub-3]
	_ = x[Ident-4]
	_ = x[Period-5]
	_ = x[SingleBefore-6]
	_ = x[BeforeAndInline-7]
	_ = x[InlineGeneral-8]
}

const _Token_name = "&|+-Ident.SingleBeforeinlineinline general"

var _Token_index = [...]uint8{0, 1, 2, 3, 4, 9, 10, 22, 28, 42}

func (i Token) String() string {
	if i < 0 || i >= Token(len(_Token_index)-1) {
		return "Token(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Token_name[_Token_index[i]:_Token_index[i+1]]
}

func (i Token) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *Token) UnmarshalText(text []byte) error {
	val, err := ParseToken(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i Token) IsAnd() bool {
	return i == And
}

func (i Token) IsOr() bool {
	return i == Or
}

func (i Token) IsAdd() bool {
	return i == Add
}

func (i Token) IsSub() bool {
	return i == Sub
}

func (i Token) IsIdent() bool {
	return i == Ident
}

func (i Token) IsPeriod() bool {
	return i == Period
}

func (i Token) IsSingleBefore() bool {
	return i == SingleBefore
}

func (i Token) IsBeforeAndInline() bool {
	return i == BeforeAndInline
}

func (i Token) IsInlineGeneral() bool {
	return i == InlineGeneral
}

func ParseToken(s string) (Token, error) {
	switch s {
	case "&":
		return And, nil
	case "|":
		return Or, nil
	case "+":
		return Add, nil
	case "-":
		return Sub, nil
	case "Ident":
		return Ident, nil
	case ".":
		return Period, nil
	case "SingleBefore":
		return SingleBefore, nil
	case "inline":
		return BeforeAndInline, nil
	case "inline general":
		return InlineGeneral, nil
	default:
		return Token(0), fmt.Errorf("invalid Token '%s'", s)
	}
}

type TokenOptions struct {
	And             Token
	Or              Token
	Add             Token
	Sub             Token
	Ident           Token
	Period          Token
	SingleBefore    Token
	BeforeAndInline Token
	InlineGeneral   Token
}

var TokenValues = TokenOptions{
	And:             And,
	Or:              Or,
	Add:             Add,
	Sub:             Sub,
	Ident:           Ident,
	Period:          Period,
	SingleBefore:    SingleBefore,
	BeforeAndInline: BeforeAndInline,
	InlineGeneral:   InlineGeneral,
}
`

const bits_in = `type Bits uint8
const (
  X Bits = 1 << iota
  Y
  Z
`

const bits_out = `func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[X-1]
	_ = x[Y-2]
	_ = x[Z-4]
}

const (
	_Bits_name_0 = "XY"
	_Bits_name_1 = "Z"
)

var (
	_Bits_index_0 = [...]uint8{0, 1, 2}
)

func (i Bits) String() string {
	switch {
	case 1 <= i && i <= 2:
		i -= 1
		return _Bits_name_0[_Bits_index_0[i]:_Bits_index_0[i+1]]
	case i == 4:
		return _Bits_name_1
	default:
		return "Bits(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}

func (i Bits) Strings() []string {
	var result []string
	if i.HasX() {
		result = append(result, X.String())
	}
	if i.HasY() {
		result = append(result, Y.String())
	}
	if i.HasZ() {
		result = append(result, Z.String())
	}
	return result
}

func (i Bits) HasX() bool {
	return i&X != 0
}

func (i Bits) HasY() bool {
	return i&Y != 0
}

func (i Bits) HasZ() bool {
	return i&Z != 0
}

func ParseBits(strs []string) (Bits, error) {
	var result Bits

	for _, s := range strs {
		switch s {
		case "X":
			result |= X
		case "Y":
			result |= Y
		case "Z":
			result |= Z
		default:
			return Bits(0), fmt.Errorf("invalid Bits '%s'", s)
		}
	}

	return result, nil
}

type BitsOptions struct {
	X Bits
	Y Bits
	Z Bits
}

var BitsValues = BitsOptions{
	X: X,
	Y: Y,
	Z: Z,
}
`

func TestGolden(t *testing.T) {
	dir := t.TempDir()
	for _, test := range golden {
		t.Run(test.name, func(t *testing.T) {
			g := Generator{
				trimPrefix:  test.trimPrefix,
				lineComment: test.lineComment,
				empty:       test.empty,
				bits:        test.bits,
			}
			input := "package test\n" + test.input
			file := test.name + ".go"
			absFile := filepath.Join(dir, file)
			err := os.WriteFile(absFile, []byte(input), 0644)
			if err != nil {
				t.Error(err)
			}

			g.parsePackage(absFile)
			// Extract the name and type of the constant from the first line.
			tokens := strings.SplitN(test.input, " ", 3)
			if len(tokens) != 3 {
				t.Fatalf("%s: need type declaration on first line", test.name)
			}
			g.generate(tokens[1])
			got := string(g.format())

			assert.Equal(t, test.output, got)
		})
	}
}
