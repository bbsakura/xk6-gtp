package ie

import (
	"fmt"

	"github.com/wmnsk/go-gtp/gtpv2/ie"
)

// New creates new IE.
func New(itype string, ins uint8, data []byte) (*ie.IE, error) {
	iftype, err := EnumIETypeString(itype)
	if err != nil {
		return nil, fmt.Errorf("invalid IFTypeName")
	}

	return ie.New(uint8(iftype), ins, data), nil
}
