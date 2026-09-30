package gtpv2

import (
	"github.com/wmnsk/go-gtp/gtpv2/ie"
)

// ieExports returns the IE builder functions exposed to JS as `gtpv2.ie.*`.
// The set is intentionally focused on the IEs used to compose S5/S8 Create /
// Modify / Delete Session flows, so scripts can drop or mutate any single IE
// to reproduce abnormal cases without touching Go.
//
// Variadic constructors from go-gtp (e.g. NewBearerContext, NewCause) are
// wrapped so JS callers can pass a plain array of children.
func ieExports() map[string]interface{} {
	return map[string]interface{}{
		"newIMSI":                    ie.NewIMSI,
		"newMSISDN":                  ie.NewMSISDN,
		"newMobileEquipmentIdentity": ie.NewMobileEquipmentIdentity,
		"newAccessPointName":         ie.NewAccessPointName,
		"newEPSBearerID":             ie.NewEPSBearerID,
		"newPDNType":                 ie.NewPDNType,
		"newPDNAddressAllocation":    ie.NewPDNAddressAllocation,
		"newRATType":                 ie.NewRATType,
		"newAggregateMaximumBitRate": ie.NewAggregateMaximumBitRate,
		"newServingNetwork":          ie.NewServingNetwork,
		"newFullyQualifiedTEID":      ie.NewFullyQualifiedTEID,
		"newSelectionMode":           ie.NewSelectionMode,
		"newAPNRestriction":          ie.NewAPNRestriction,
		"newRecovery":                ie.NewRecovery,
		"newBearerQoS":               ie.NewBearerQoS,
		"newFullyQualifiedCSID":      ie.NewFullyQualifiedCSID,
		"newTAI":                     ie.NewTAI,
		"newECGI":                    ie.NewECGI,
		"newUETimeZone":              ie.NewUETimeZone,
		"newBearerContext":           newBearerContextJS,
		"newCause":                   newCauseJS,
	}
}

// newBearerContextJS wraps ie.NewBearerContext (variadic) so JS scripts can
// pass a plain array of child IEs.
func newBearerContextJS(children []*ie.IE) *ie.IE {
	return ie.NewBearerContext(children...)
}

// newCauseJS exposes ie.NewCause with the offendingIE parameter optional
// (nil) — JS callers usually only care about the numeric Cause value.
func newCauseJS(cause, pce, bce, cs uint8) *ie.IE {
	return ie.NewCause(cause, pce, bce, cs, nil)
}
