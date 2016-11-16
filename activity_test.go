package main

import (
	"testing"
)

var testTypes = []struct {
	integerType string
	byteNumber  uint
}{
	{"bigint", uint(8)},
	{"int", uint(4)},
	{"mediumint", uint(3)},
	{"smallint", uint(2)},
	{"tinyint", uint(1)},
}

func TestTypeToByte(t *testing.T) {
	var column AutoIncrementColumn
	for _, tt := range testTypes {
		//pp.Println(tt.integerType)
		var bt uint = column.typeToByte(tt.integerType)
		if bt == 0 {
			t.Error("testTypeToByte: missiing integer type")
			continue
		} else if bt != tt.byteNumber {
			t.Errorf("testTypeToByte: response expected %v got %v", tt.byteNumber, bt)
			continue
		}
	}
}

var limitTests = []struct {
	byte       uint
	unsigned   bool
	upperLimit uint
}{
	{uint(1), false, uint(127)},
	{uint(1), true, uint(255)},
	{uint(2), false, uint(32767)},
	{uint(2), true, uint(65535)},
	{uint(3), false, uint(8388607)},
	{uint(3), true, uint(16777215)},
	{uint(4), false, uint(2147483647)},
	{uint(4), true, uint(4294967295)},
	{uint(8), false, uint(9223372036854775807)},
	{uint(8), true, uint(18446744073709551615)},
}

func TestCalculateLimit(t *testing.T) {
	var column AutoIncrementColumn

	for _, lt := range limitTests {
		column.Unsigned = lt.unsigned
		var limit uint = column.calculateLimit(lt.byte)
		if limit == 0 {
			t.Error("TestCalculateLimit: received byte is 0")
			continue
		} else if limit != lt.upperLimit {
			t.Errorf("TestCalculateLimit: expected %v got %v", lt.upperLimit, limit)
			continue
		}
	}
}

var testColumnTypes = []struct {
	columnType  string
	integerType string
}{
	{"bigint(20) unsigned", "bigint"},
	{"bigint(20)", "bigint"},
	{"int(11)", "int"},
	{"mediumint(8)", "mediumint"},
	{"smallint(5)", "smallint"},
	{"tinyint(3)", "tinyint"},
	{"int", "int"},
}

func TestColumnTypeToType(t *testing.T) {

	for _, ct := range testColumnTypes {
		var column AutoIncrementColumn
		it := column.columnTypeToType(ct.columnType)
		if it == "" {
			t.Error("TestColumnTypeToType: illegal format(blank? or started by bracket?)")
			continue
		} else if it != ct.integerType {
			t.Errorf("TestColumnTypeToType: expected %v got %v", ct.integerType, it)
			continue
		}
	}
}

var testColumnTypesForUnsigned = []struct {
	columnType string
	unsigned   bool
}{
	{"bigint(20) unsigned", true},
	{"bigint(20)", false},
	{"int(11) unsigned", true},
	{"mediumint(8)", false},
	{"smallint(5) unsigned", true},
	{"tinyint(3)", false},
	{"int", false},
}

func TestColumnTypeToUnsigned(t *testing.T) {

	for _, ut := range testColumnTypesForUnsigned {
		var column AutoIncrementColumn
		unsigned := column.columnTypeToUnsigned(ut.columnType)
		if unsigned != ut.unsigned {
			t.Errorf("TestColumnTypeToUnsinged: expected %v got %v", ut.unsigned, unsigned)
			continue
		}
	}
}
