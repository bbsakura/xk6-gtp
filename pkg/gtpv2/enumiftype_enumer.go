// Code generated by "enumer -type=EnumIFType"; DO NOT EDIT.

package gtpv2

import (
	"fmt"
	"strings"
)

const _EnumIFTypeName = "IFTypeS1UeNodeBGTPUIFTypeS1USGWGTPUIFTypeS12RNCGTPUIFTypeS12SGWGTPUIFTypeS5S8SGWGTPUIFTypeS5S8PGWGTPUIFTypeS5S8SGWGTPCIFTypeS5S8PGWGTPCIFTypeS5S8SGWPMIPv6IFTypeS5S8PGWPMIPv6IFTypeS11MMEGTPCIFTypeS11S4SGWGTPCIFTypeS10MMEGTPCIFTypeS3MMEGTPCIFTypeS3SGSNGTPCIFTypeS4SGSNGTPUIFTypeS4SGWGTPUIFTypeS4SGSNGTPCIFTypeS16SGSNGTPCIFTypeeNodeBGTPUForDLIFTypeeNodeBGTPUForULIFTypeRNCGTPUForDataIFTypeSGSNGTPUForDataIFTypeSGWUPFGTPUForDLIFTypeSmMBMSGWGTPCIFTypeSnMBMSGWGTPCIFTypeSmMMEGTPCIFTypeSnSGSNGTPCIFTypeSGWGTPUForULIFTypeSnSGSNGTPUIFTypeS2bePDGGTPCIFTypeS2bUePDGGTPUIFTypeS2bPGWGTPCIFTypeS2bUPGWGTPUIFTypeS2aTWANGTPUIFTypeS2aTWANGTPCIFTypeS2aPGWGTPCIFTypeS2aPGWGTPUIFTypeS11MMEGTPUIFTypeS11SGWGTPU"

var _EnumIFTypeIndex = [...]uint16{0, 19, 35, 51, 67, 84, 101, 118, 135, 154, 173, 189, 207, 223, 238, 254, 270, 285, 301, 318, 339, 360, 380, 401, 422, 440, 458, 473, 489, 507, 523, 540, 558, 574, 591, 608, 625, 641, 657, 673, 689}

const _EnumIFTypeLowerName = "iftypes1uenodebgtpuiftypes1usgwgtpuiftypes12rncgtpuiftypes12sgwgtpuiftypes5s8sgwgtpuiftypes5s8pgwgtpuiftypes5s8sgwgtpciftypes5s8pgwgtpciftypes5s8sgwpmipv6iftypes5s8pgwpmipv6iftypes11mmegtpciftypes11s4sgwgtpciftypes10mmegtpciftypes3mmegtpciftypes3sgsngtpciftypes4sgsngtpuiftypes4sgwgtpuiftypes4sgsngtpciftypes16sgsngtpciftypeenodebgtpufordliftypeenodebgtpuforuliftyperncgtpufordataiftypesgsngtpufordataiftypesgwupfgtpufordliftypesmmbmsgwgtpciftypesnmbmsgwgtpciftypesmmmegtpciftypesnsgsngtpciftypesgwgtpuforuliftypesnsgsngtpuiftypes2bepdggtpciftypes2buepdggtpuiftypes2bpgwgtpciftypes2bupgwgtpuiftypes2atwangtpuiftypes2atwangtpciftypes2apgwgtpciftypes2apgwgtpuiftypes11mmegtpuiftypes11sgwgtpu"

func (i EnumIFType) String() string {
	if i >= EnumIFType(len(_EnumIFTypeIndex)-1) {
		return fmt.Sprintf("EnumIFType(%d)", i)
	}
	return _EnumIFTypeName[_EnumIFTypeIndex[i]:_EnumIFTypeIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _EnumIFTypeNoOp() {
	var x [1]struct{}
	_ = x[IFTypeS1UeNodeBGTPU-(0)]
	_ = x[IFTypeS1USGWGTPU-(1)]
	_ = x[IFTypeS12RNCGTPU-(2)]
	_ = x[IFTypeS12SGWGTPU-(3)]
	_ = x[IFTypeS5S8SGWGTPU-(4)]
	_ = x[IFTypeS5S8PGWGTPU-(5)]
	_ = x[IFTypeS5S8SGWGTPC-(6)]
	_ = x[IFTypeS5S8PGWGTPC-(7)]
	_ = x[IFTypeS5S8SGWPMIPv6-(8)]
	_ = x[IFTypeS5S8PGWPMIPv6-(9)]
	_ = x[IFTypeS11MMEGTPC-(10)]
	_ = x[IFTypeS11S4SGWGTPC-(11)]
	_ = x[IFTypeS10MMEGTPC-(12)]
	_ = x[IFTypeS3MMEGTPC-(13)]
	_ = x[IFTypeS3SGSNGTPC-(14)]
	_ = x[IFTypeS4SGSNGTPU-(15)]
	_ = x[IFTypeS4SGWGTPU-(16)]
	_ = x[IFTypeS4SGSNGTPC-(17)]
	_ = x[IFTypeS16SGSNGTPC-(18)]
	_ = x[IFTypeeNodeBGTPUForDL-(19)]
	_ = x[IFTypeeNodeBGTPUForUL-(20)]
	_ = x[IFTypeRNCGTPUForData-(21)]
	_ = x[IFTypeSGSNGTPUForData-(22)]
	_ = x[IFTypeSGWUPFGTPUForDL-(23)]
	_ = x[IFTypeSmMBMSGWGTPC-(24)]
	_ = x[IFTypeSnMBMSGWGTPC-(25)]
	_ = x[IFTypeSmMMEGTPC-(26)]
	_ = x[IFTypeSnSGSNGTPC-(27)]
	_ = x[IFTypeSGWGTPUForUL-(28)]
	_ = x[IFTypeSnSGSNGTPU-(29)]
	_ = x[IFTypeS2bePDGGTPC-(30)]
	_ = x[IFTypeS2bUePDGGTPU-(31)]
	_ = x[IFTypeS2bPGWGTPC-(32)]
	_ = x[IFTypeS2bUPGWGTPU-(33)]
	_ = x[IFTypeS2aTWANGTPU-(34)]
	_ = x[IFTypeS2aTWANGTPC-(35)]
	_ = x[IFTypeS2aPGWGTPC-(36)]
	_ = x[IFTypeS2aPGWGTPU-(37)]
	_ = x[IFTypeS11MMEGTPU-(38)]
	_ = x[IFTypeS11SGWGTPU-(39)]
}

var _EnumIFTypeValues = []EnumIFType{IFTypeS1UeNodeBGTPU, IFTypeS1USGWGTPU, IFTypeS12RNCGTPU, IFTypeS12SGWGTPU, IFTypeS5S8SGWGTPU, IFTypeS5S8PGWGTPU, IFTypeS5S8SGWGTPC, IFTypeS5S8PGWGTPC, IFTypeS5S8SGWPMIPv6, IFTypeS5S8PGWPMIPv6, IFTypeS11MMEGTPC, IFTypeS11S4SGWGTPC, IFTypeS10MMEGTPC, IFTypeS3MMEGTPC, IFTypeS3SGSNGTPC, IFTypeS4SGSNGTPU, IFTypeS4SGWGTPU, IFTypeS4SGSNGTPC, IFTypeS16SGSNGTPC, IFTypeeNodeBGTPUForDL, IFTypeeNodeBGTPUForUL, IFTypeRNCGTPUForData, IFTypeSGSNGTPUForData, IFTypeSGWUPFGTPUForDL, IFTypeSmMBMSGWGTPC, IFTypeSnMBMSGWGTPC, IFTypeSmMMEGTPC, IFTypeSnSGSNGTPC, IFTypeSGWGTPUForUL, IFTypeSnSGSNGTPU, IFTypeS2bePDGGTPC, IFTypeS2bUePDGGTPU, IFTypeS2bPGWGTPC, IFTypeS2bUPGWGTPU, IFTypeS2aTWANGTPU, IFTypeS2aTWANGTPC, IFTypeS2aPGWGTPC, IFTypeS2aPGWGTPU, IFTypeS11MMEGTPU, IFTypeS11SGWGTPU}

var _EnumIFTypeNameToValueMap = map[string]EnumIFType{
	_EnumIFTypeName[0:19]:         IFTypeS1UeNodeBGTPU,
	_EnumIFTypeLowerName[0:19]:    IFTypeS1UeNodeBGTPU,
	_EnumIFTypeName[19:35]:        IFTypeS1USGWGTPU,
	_EnumIFTypeLowerName[19:35]:   IFTypeS1USGWGTPU,
	_EnumIFTypeName[35:51]:        IFTypeS12RNCGTPU,
	_EnumIFTypeLowerName[35:51]:   IFTypeS12RNCGTPU,
	_EnumIFTypeName[51:67]:        IFTypeS12SGWGTPU,
	_EnumIFTypeLowerName[51:67]:   IFTypeS12SGWGTPU,
	_EnumIFTypeName[67:84]:        IFTypeS5S8SGWGTPU,
	_EnumIFTypeLowerName[67:84]:   IFTypeS5S8SGWGTPU,
	_EnumIFTypeName[84:101]:       IFTypeS5S8PGWGTPU,
	_EnumIFTypeLowerName[84:101]:  IFTypeS5S8PGWGTPU,
	_EnumIFTypeName[101:118]:      IFTypeS5S8SGWGTPC,
	_EnumIFTypeLowerName[101:118]: IFTypeS5S8SGWGTPC,
	_EnumIFTypeName[118:135]:      IFTypeS5S8PGWGTPC,
	_EnumIFTypeLowerName[118:135]: IFTypeS5S8PGWGTPC,
	_EnumIFTypeName[135:154]:      IFTypeS5S8SGWPMIPv6,
	_EnumIFTypeLowerName[135:154]: IFTypeS5S8SGWPMIPv6,
	_EnumIFTypeName[154:173]:      IFTypeS5S8PGWPMIPv6,
	_EnumIFTypeLowerName[154:173]: IFTypeS5S8PGWPMIPv6,
	_EnumIFTypeName[173:189]:      IFTypeS11MMEGTPC,
	_EnumIFTypeLowerName[173:189]: IFTypeS11MMEGTPC,
	_EnumIFTypeName[189:207]:      IFTypeS11S4SGWGTPC,
	_EnumIFTypeLowerName[189:207]: IFTypeS11S4SGWGTPC,
	_EnumIFTypeName[207:223]:      IFTypeS10MMEGTPC,
	_EnumIFTypeLowerName[207:223]: IFTypeS10MMEGTPC,
	_EnumIFTypeName[223:238]:      IFTypeS3MMEGTPC,
	_EnumIFTypeLowerName[223:238]: IFTypeS3MMEGTPC,
	_EnumIFTypeName[238:254]:      IFTypeS3SGSNGTPC,
	_EnumIFTypeLowerName[238:254]: IFTypeS3SGSNGTPC,
	_EnumIFTypeName[254:270]:      IFTypeS4SGSNGTPU,
	_EnumIFTypeLowerName[254:270]: IFTypeS4SGSNGTPU,
	_EnumIFTypeName[270:285]:      IFTypeS4SGWGTPU,
	_EnumIFTypeLowerName[270:285]: IFTypeS4SGWGTPU,
	_EnumIFTypeName[285:301]:      IFTypeS4SGSNGTPC,
	_EnumIFTypeLowerName[285:301]: IFTypeS4SGSNGTPC,
	_EnumIFTypeName[301:318]:      IFTypeS16SGSNGTPC,
	_EnumIFTypeLowerName[301:318]: IFTypeS16SGSNGTPC,
	_EnumIFTypeName[318:339]:      IFTypeeNodeBGTPUForDL,
	_EnumIFTypeLowerName[318:339]: IFTypeeNodeBGTPUForDL,
	_EnumIFTypeName[339:360]:      IFTypeeNodeBGTPUForUL,
	_EnumIFTypeLowerName[339:360]: IFTypeeNodeBGTPUForUL,
	_EnumIFTypeName[360:380]:      IFTypeRNCGTPUForData,
	_EnumIFTypeLowerName[360:380]: IFTypeRNCGTPUForData,
	_EnumIFTypeName[380:401]:      IFTypeSGSNGTPUForData,
	_EnumIFTypeLowerName[380:401]: IFTypeSGSNGTPUForData,
	_EnumIFTypeName[401:422]:      IFTypeSGWUPFGTPUForDL,
	_EnumIFTypeLowerName[401:422]: IFTypeSGWUPFGTPUForDL,
	_EnumIFTypeName[422:440]:      IFTypeSmMBMSGWGTPC,
	_EnumIFTypeLowerName[422:440]: IFTypeSmMBMSGWGTPC,
	_EnumIFTypeName[440:458]:      IFTypeSnMBMSGWGTPC,
	_EnumIFTypeLowerName[440:458]: IFTypeSnMBMSGWGTPC,
	_EnumIFTypeName[458:473]:      IFTypeSmMMEGTPC,
	_EnumIFTypeLowerName[458:473]: IFTypeSmMMEGTPC,
	_EnumIFTypeName[473:489]:      IFTypeSnSGSNGTPC,
	_EnumIFTypeLowerName[473:489]: IFTypeSnSGSNGTPC,
	_EnumIFTypeName[489:507]:      IFTypeSGWGTPUForUL,
	_EnumIFTypeLowerName[489:507]: IFTypeSGWGTPUForUL,
	_EnumIFTypeName[507:523]:      IFTypeSnSGSNGTPU,
	_EnumIFTypeLowerName[507:523]: IFTypeSnSGSNGTPU,
	_EnumIFTypeName[523:540]:      IFTypeS2bePDGGTPC,
	_EnumIFTypeLowerName[523:540]: IFTypeS2bePDGGTPC,
	_EnumIFTypeName[540:558]:      IFTypeS2bUePDGGTPU,
	_EnumIFTypeLowerName[540:558]: IFTypeS2bUePDGGTPU,
	_EnumIFTypeName[558:574]:      IFTypeS2bPGWGTPC,
	_EnumIFTypeLowerName[558:574]: IFTypeS2bPGWGTPC,
	_EnumIFTypeName[574:591]:      IFTypeS2bUPGWGTPU,
	_EnumIFTypeLowerName[574:591]: IFTypeS2bUPGWGTPU,
	_EnumIFTypeName[591:608]:      IFTypeS2aTWANGTPU,
	_EnumIFTypeLowerName[591:608]: IFTypeS2aTWANGTPU,
	_EnumIFTypeName[608:625]:      IFTypeS2aTWANGTPC,
	_EnumIFTypeLowerName[608:625]: IFTypeS2aTWANGTPC,
	_EnumIFTypeName[625:641]:      IFTypeS2aPGWGTPC,
	_EnumIFTypeLowerName[625:641]: IFTypeS2aPGWGTPC,
	_EnumIFTypeName[641:657]:      IFTypeS2aPGWGTPU,
	_EnumIFTypeLowerName[641:657]: IFTypeS2aPGWGTPU,
	_EnumIFTypeName[657:673]:      IFTypeS11MMEGTPU,
	_EnumIFTypeLowerName[657:673]: IFTypeS11MMEGTPU,
	_EnumIFTypeName[673:689]:      IFTypeS11SGWGTPU,
	_EnumIFTypeLowerName[673:689]: IFTypeS11SGWGTPU,
}

var _EnumIFTypeNames = []string{
	_EnumIFTypeName[0:19],
	_EnumIFTypeName[19:35],
	_EnumIFTypeName[35:51],
	_EnumIFTypeName[51:67],
	_EnumIFTypeName[67:84],
	_EnumIFTypeName[84:101],
	_EnumIFTypeName[101:118],
	_EnumIFTypeName[118:135],
	_EnumIFTypeName[135:154],
	_EnumIFTypeName[154:173],
	_EnumIFTypeName[173:189],
	_EnumIFTypeName[189:207],
	_EnumIFTypeName[207:223],
	_EnumIFTypeName[223:238],
	_EnumIFTypeName[238:254],
	_EnumIFTypeName[254:270],
	_EnumIFTypeName[270:285],
	_EnumIFTypeName[285:301],
	_EnumIFTypeName[301:318],
	_EnumIFTypeName[318:339],
	_EnumIFTypeName[339:360],
	_EnumIFTypeName[360:380],
	_EnumIFTypeName[380:401],
	_EnumIFTypeName[401:422],
	_EnumIFTypeName[422:440],
	_EnumIFTypeName[440:458],
	_EnumIFTypeName[458:473],
	_EnumIFTypeName[473:489],
	_EnumIFTypeName[489:507],
	_EnumIFTypeName[507:523],
	_EnumIFTypeName[523:540],
	_EnumIFTypeName[540:558],
	_EnumIFTypeName[558:574],
	_EnumIFTypeName[574:591],
	_EnumIFTypeName[591:608],
	_EnumIFTypeName[608:625],
	_EnumIFTypeName[625:641],
	_EnumIFTypeName[641:657],
	_EnumIFTypeName[657:673],
	_EnumIFTypeName[673:689],
}

// EnumIFTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func EnumIFTypeString(s string) (EnumIFType, error) {
	if val, ok := _EnumIFTypeNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _EnumIFTypeNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to EnumIFType values", s)
}

// EnumIFTypeValues returns all values of the enum
func EnumIFTypeValues() []EnumIFType {
	return _EnumIFTypeValues
}

// EnumIFTypeStrings returns a slice of all String values of the enum
func EnumIFTypeStrings() []string {
	strs := make([]string, len(_EnumIFTypeNames))
	copy(strs, _EnumIFTypeNames)
	return strs
}

// IsAEnumIFType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i EnumIFType) IsAEnumIFType() bool {
	for _, v := range _EnumIFTypeValues {
		if i == v {
			return true
		}
	}
	return false
}
