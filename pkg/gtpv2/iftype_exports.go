package gtpv2

import (
	"github.com/wmnsk/go-gtp/gtpv2"
)

// ifTypeExports returns the GTPv2 FTEID interface-type numeric codes exposed
// to JS as `gtpv2.IFType.*`. Values match the go-gtp constants exactly, which
// in turn match 3GPP TS 29.274 clause 8.22 table 8.22-1.
//
// Exposed as numbers so scripts can pass them into `ie.newFullyQualifiedTEID`
// or into the IfTypeCplane / IfTypeUplane fields on session params without
// having to memorise the codes.
func ifTypeExports() map[string]uint8 {
	return map[string]uint8{
		"S1U_ENODEB_GTPU": gtpv2.IFTypeS1UeNodeBGTPU,
		"S1U_SGW_GTPU":    gtpv2.IFTypeS1USGWGTPU,
		"S12_RNC_GTPU":    gtpv2.IFTypeS12RNCGTPU,
		"S12_SGW_GTPU":    gtpv2.IFTypeS12SGWGTPU,
		"S5S8_SGW_GTPU":   gtpv2.IFTypeS5S8SGWGTPU,
		"S5S8_PGW_GTPU":   gtpv2.IFTypeS5S8PGWGTPU,
		"S5S8_SGW_GTPC":   gtpv2.IFTypeS5S8SGWGTPC,
		"S5S8_PGW_GTPC":   gtpv2.IFTypeS5S8PGWGTPC,
		"S11_MME_GTPC":    gtpv2.IFTypeS11MMEGTPC,
		"S11_S4_SGW_GTPC": gtpv2.IFTypeS11S4SGWGTPC,
		"S10_MME_GTPC":    gtpv2.IFTypeS10MMEGTPC,
		"S3_MME_GTPC":     gtpv2.IFTypeS3MMEGTPC,
		"S3_SGSN_GTPC":    gtpv2.IFTypeS3SGSNGTPC,
		"S4_SGSN_GTPU":    gtpv2.IFTypeS4SGSNGTPU,
		"S4_SGW_GTPU":     gtpv2.IFTypeS4SGWGTPU,
		"S4_SGSN_GTPC":    gtpv2.IFTypeS4SGSNGTPC,
		"S16_SGSN_GTPC":   gtpv2.IFTypeS16SGSNGTPC,
		"S2B_EPDG_GTPC":   gtpv2.IFTypeS2bePDGGTPC,
		"S2B_UEPDG_GTPU":  gtpv2.IFTypeS2bUePDGGTPU,
		"S2B_PGW_GTPC":    gtpv2.IFTypeS2bPGWGTPC,
		"S2B_UPGW_GTPU":   gtpv2.IFTypeS2bUPGWGTPU,
		"S2A_TWAN_GTPU":   gtpv2.IFTypeS2aTWANGTPU,
		"S2A_TWAN_GTPC":   gtpv2.IFTypeS2aTWANGTPC,
		"S2A_PGW_GTPC":    gtpv2.IFTypeS2aPGWGTPC,
		"S2A_PGW_GTPU":    gtpv2.IFTypeS2aPGWGTPU,
	}
}
