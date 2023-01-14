package utils_test

import (
	"testing"

	"github.com/bohdanch-w/go-tgupload/pkg/utils"
	"github.com/stretchr/testify/require"
)

var testList = []string{ // nolint: gochecknoglobals
	"1000X Radonius Maximus",
	"000050X Radonius",
	"10X Radonius",
	"200X Radonius",
	"20X Radonius",
	"20X Radonius Prime",
	"30X Radonius",
	"40X Radonius",
	"Allegia 50 Clasteron",
	"Allegia 500 Clasteron",
	"Allegia 50B Clasteron",
	"Allegia 51 Clasteron",
	"Allegia 6R Clasteron",
	"Alpha 100",
	"Alpha 2",
	"Alpha 200",
	"Alpha 2A",
	"Alpha 2A-8000",
	"Alpha 2A-900",
	"Callisto Morphamax",
	"Callisto Morphamax 500",
	"Callisto Morphamax 5000",
	"Callisto Morphamax 600",
	"Callisto Morphamax 6000 SE",
	"Callisto Morphamax 6000 SE2",
	"Callisto Morphamax 700",
	"Callisto Morphamax 7000",
	"Xiph Xlater 10000",
	"Xiph Xlater 2000",
	"Xiph Xlater 300",
	"Xiph Xlater 40",
	"Xiph Xlater 5",
	"Xiph Xlater 50",
	"Xiph Xlater 500",
	"Xiph Xlater 5000",
	"Xiph Xlater 58",
}

func TestSortStrings1(t *testing.T) {
	testListSortedOK := []string{
		"10X Radonius",
		"20X Radonius",
		"20X Radonius Prime",
		"30X Radonius",
		"40X Radonius",
		"000050X Radonius",
		"200X Radonius",
		"1000X Radonius Maximus",
		"Allegia 6R Clasteron",
		"Allegia 50 Clasteron",
		"Allegia 50B Clasteron",
		"Allegia 51 Clasteron",
		"Allegia 500 Clasteron",
		"Alpha 2",
		"Alpha 2A",
		"Alpha 2A-900",
		"Alpha 2A-8000",
		"Alpha 100",
		"Alpha 200",
		"Callisto Morphamax",
		"Callisto Morphamax 500",
		"Callisto Morphamax 600",
		"Callisto Morphamax 700",
		"Callisto Morphamax 5000",
		"Callisto Morphamax 6000 SE",
		"Callisto Morphamax 6000 SE2",
		"Callisto Morphamax 7000",
		"Xiph Xlater 5",
		"Xiph Xlater 40",
		"Xiph Xlater 50",
		"Xiph Xlater 58",
		"Xiph Xlater 300",
		"Xiph Xlater 500",
		"Xiph Xlater 2000",
		"Xiph Xlater 5000",
		"Xiph Xlater 10000",
	}

	testListSorted := make([]string, len(testList))
	copy(testListSorted, testList)

	utils.NaturalSort(testListSorted, func(s string) string { return s })
	require.Equal(t, testListSortedOK, testListSorted)
}

func TestSortStrings2(t *testing.T) {
	testList := []string{
		"z1.doc",
		"z10.doc",
		"z100.doc",
		"z101.doc",
		"z102.doc",
		"z11.doc",
		"z12.doc",
		"z13.doc",
		"z14.doc",
		"z15.doc",
		"z16.doc",
		"z17.doc",
		"z18.doc",
		"z19.doc",
		"z2.doc",
		"z20.doc",
		"z3.doc",
		"z4.doc",
		"z5.doc",
		"z6.doc",
		"z7.doc",
		"z8.doc",
		"z9.doc",
	}

	testListSortedOK := []string{
		"z1.doc",
		"z2.doc",
		"z3.doc",
		"z4.doc",
		"z5.doc",
		"z6.doc",
		"z7.doc",
		"z8.doc",
		"z9.doc",
		"z10.doc",
		"z11.doc",
		"z12.doc",
		"z13.doc",
		"z14.doc",
		"z15.doc",
		"z16.doc",
		"z17.doc",
		"z18.doc",
		"z19.doc",
		"z20.doc",
		"z100.doc",
		"z101.doc",
		"z102.doc",
	}

	testListSorted := make([]string, len(testList))
	copy(testListSorted, testList)

	utils.NaturalSort(testListSorted, func(s string) string { return s })
	require.Equal(t, testListSortedOK, testListSorted)
}

func TestSortStructs(t *testing.T) {
	type testStruct struct {
		a string
	}

	testList := []testStruct{
		{"z1.doc"},
		{"z10.doc"},
		{"z100.doc"},
		{"z101.doc"},
		{"z102.doc"},
		{"z11.doc"},
		{"z12.doc"},
		{"z13.doc"},
		{"z14.doc"},
		{"z15.doc"},
		{"z16.doc"},
		{"z17.doc"},
		{"z18.doc"},
		{"z19.doc"},
		{"z2.doc"},
		{"z20.doc"},
		{"z3.doc"},
		{"z4.doc"},
		{"z5.doc"},
		{"z6.doc"},
		{"z7.doc"},
		{"z8.doc"},
		{"z9.doc"},
	}

	testListSortedOK := []testStruct{
		{"z1.doc"},
		{"z2.doc"},
		{"z3.doc"},
		{"z4.doc"},
		{"z5.doc"},
		{"z6.doc"},
		{"z7.doc"},
		{"z8.doc"},
		{"z9.doc"},
		{"z10.doc"},
		{"z11.doc"},
		{"z12.doc"},
		{"z13.doc"},
		{"z14.doc"},
		{"z15.doc"},
		{"z16.doc"},
		{"z17.doc"},
		{"z18.doc"},
		{"z19.doc"},
		{"z20.doc"},
		{"z100.doc"},
		{"z101.doc"},
		{"z102.doc"},
	}

	utils.NaturalSort(testList, func(s testStruct) string { return s.a })

	require.Equal(t, testListSortedOK, testList)
}

func BenchmarkSort1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		utils.NaturalSort(testList, func(s string) string { return s })
	}
}
