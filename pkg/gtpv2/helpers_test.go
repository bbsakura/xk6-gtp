package gtpv2

import "testing"

func TestGenerateDummyIMSI(t *testing.T) {
	type testCase struct {
		input int
		want  string
	}
	for _, c := range []testCase{
		{input: 1, want: "454060000000001"},
		{input: 1111111111, want: "454061111111111"},
		{input: 10000000001, want: "454060000000001"},
		{input: 11111111111, want: "454061111111111"},
		{input: -1, want: "454060000000001"},
	} {
		got := GenerateDummyIMSI(c.input)
		if got != c.want {
			t.Error(got, "!=", c.want)
		}
	}
}
