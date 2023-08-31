package gtpv2

import "fmt"

func GenerateDummyIMSI(n int) string {
	if n < 0 {
		n *= -1
	}
	return fmt.Sprintf("45406%010d", n%10000000000)
}
