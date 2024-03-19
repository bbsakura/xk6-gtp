// Code generated by "enumer -type=EnumIFCause"; DO NOT EDIT.

package gtpv2

import (
	"fmt"
	"strings"
)

const (
	_EnumIFCauseName_0      = "CauseLocalDetachCauseCompleteDetachCauseRATChangedFrom3GPPToNon3GPPCauseISRDeactivationCauseErrorIndicationReceivedFromRNCeNodeBS4SGSNMMECauseIMSIDetachOnlyCauseReactivationRequestedCausePDNReconnectionToThisAPNDisallowedCauseAccessChangedFromNon3GPPTo3GPPCausePDNConnectionInactivityTimerExpiresCausePGWNotRespondingCauseNetworkFailureCauseQoSParameterMismatch"
	_EnumIFCauseLowerName_0 = "causelocaldetachcausecompletedetachcauseratchangedfrom3gpptonon3gppcauseisrdeactivationcauseerrorindicationreceivedfromrncenodebs4sgsnmmecauseimsidetachonlycausereactivationrequestedcausepdnreconnectiontothisapndisallowedcauseaccesschangedfromnon3gppto3gppcausepdnconnectioninactivitytimerexpirescausepgwnotrespondingcausenetworkfailurecauseqosparametermismatch"
	_EnumIFCauseName_1      = "CauseRequestAcceptedCauseRequestAcceptedPartiallyCauseNewPDNTypeDueToNetworkPreferenceCauseNewPDNTypeDueToSingleAddressBearerOnly"
	_EnumIFCauseLowerName_1 = "causerequestacceptedcauserequestacceptedpartiallycausenewpdntypeduetonetworkpreferencecausenewpdntypeduetosingleaddressbeareronly"
	_EnumIFCauseName_2      = "CauseContextNotFoundCauseInvalidMessageFormatCauseVersionNotSupportedByNextPeerCauseInvalidLengthCauseServiceNotSupportedCauseMandatoryIEIncorrectCauseMandatoryIEMissing"
	_EnumIFCauseLowerName_2 = "causecontextnotfoundcauseinvalidmessageformatcauseversionnotsupportedbynextpeercauseinvalidlengthcauseservicenotsupportedcausemandatoryieincorrectcausemandatoryiemissing"
	_EnumIFCauseName_3      = "CauseSystemFailureCauseNoResourcesAvailableCauseSemanticErrorInTheTFTOperationCauseSyntacticErrorInTheTFTOperationCauseSemanticErrorsInPacketFiltersCauseSyntacticErrorsInPacketFiltersCauseMissingOrUnknownAPN"
	_EnumIFCauseLowerName_3 = "causesystemfailurecausenoresourcesavailablecausesemanticerrorinthetftoperationcausesyntacticerrorinthetftoperationcausesemanticerrorsinpacketfilterscausesyntacticerrorsinpacketfilterscausemissingorunknownapn"
	_EnumIFCauseName_4      = "CauseGREKeyNotFoundCauseRelocationFailureCauseDeniedInRATCausePreferredPDNTypeNotSupportedCauseAllDynamicAddressesAreOccupiedCauseUEContextWithoutTFTAlreadyActivatedCauseProtocolTypeNotSupportedCauseUENotRespondingCauseUERefusesCauseServiceDeniedCauseUnableToPageUECauseNoMemoryAvailableCauseUserAuthenticationFailedCauseAPNAccessDeniedNoSubscriptionCauseRequestRejectedReasonNotSpecifiedCausePTMSISignatureMismatchCauseIMSIIMEINotKnownCauseSemanticErrorInTheTADOperationCauseSyntacticErrorInTheTADOperation"
	_EnumIFCauseLowerName_4 = "causegrekeynotfoundcauserelocationfailurecausedeniedinratcausepreferredpdntypenotsupportedcausealldynamicaddressesareoccupiedcauseuecontextwithouttftalreadyactivatedcauseprotocoltypenotsupportedcauseuenotrespondingcauseuerefusescauseservicedeniedcauseunabletopageuecausenomemoryavailablecauseuserauthenticationfailedcauseapnaccessdeniednosubscriptioncauserequestrejectedreasonnotspecifiedcauseptmsisignaturemismatchcauseimsiimeinotknowncausesemanticerrorinthetadoperationcausesyntacticerrorinthetadoperation"
	_EnumIFCauseName_5      = "CauseRemotePeerNotRespondingCauseCollisionWithNetworkInitiatedRequestCauseUnableToPageUEDueToSuspensionCauseConditionalIEMissingCauseAPNRestrictionTypeIncompatibleWithCurrentlyActivePDNConnectionCauseInvalidOverallLengthOfTheTriggeredResponseMessageAndAPiggybackedInitialMessageCauseDataForwardingNotSupportedCauseInvalidReplyFromRemotePeerCauseFallbackToGTPv1CauseInvalidPeerCauseTemporarilyRejectedDueToHandoverTAURAUProcedureInProgressCauseModificationsNotLimitedToS1UBearersCauseRequestRejectedForAPMIPv6ReasonCauseAPNCongestionCauseBearerHandlingNotSupportedCauseUEAlreadyReattachedCauseMultiplePDNConnectionsForAGivenAPNNotAllowedCauseTargetAccessRestrictedForTheSubscriber"
	_EnumIFCauseLowerName_5 = "causeremotepeernotrespondingcausecollisionwithnetworkinitiatedrequestcauseunabletopageueduetosuspensioncauseconditionaliemissingcauseapnrestrictiontypeincompatiblewithcurrentlyactivepdnconnectioncauseinvalidoveralllengthofthetriggeredresponsemessageandapiggybackedinitialmessagecausedataforwardingnotsupportedcauseinvalidreplyfromremotepeercausefallbacktogtpv1causeinvalidpeercausetemporarilyrejectedduetohandovertaurauprocedureinprogresscausemodificationsnotlimitedtos1ubearerscauserequestrejectedforapmipv6reasoncauseapncongestioncausebearerhandlingnotsupportedcauseuealreadyreattachedcausemultiplepdnconnectionsforagivenapnnotallowedcausetargetaccessrestrictedforthesubscriber"
	_EnumIFCauseName_6      = "CauseMMESGSNRefusesDueToVPLMNPolicyCauseGTPCEntityCongestionCauseLateOverlappingRequestCauseTimedOutRequestCauseUEIsTemporarilyNotReachableDueToPowerSavingCauseRelocationFailureDueToNASMessageRedirectionCauseUENotAuthorisedByOCSOrExternalAAAServerCauseMultipleAccessesToAPDNConnectionNotAllowedCauseRequestRejectedDueToUECapabilityCauseS1UPathFailure"
	_EnumIFCauseLowerName_6 = "causemmesgsnrefusesduetovplmnpolicycausegtpcentitycongestioncauselateoverlappingrequestcausetimedoutrequestcauseueistemporarilynotreachableduetopowersavingcauserelocationfailureduetonasmessageredirectioncauseuenotauthorisedbyocsorexternalaaaservercausemultipleaccessestoapdnconnectionnotallowedcauserequestrejectedduetouecapabilitycauses1upathfailure"
)

var (
	_EnumIFCauseIndex_0 = [...]uint16{0, 16, 35, 67, 87, 137, 156, 182, 221, 256, 296, 317, 336, 361}
	_EnumIFCauseIndex_1 = [...]uint8{0, 20, 49, 86, 129}
	_EnumIFCauseIndex_2 = [...]uint8{0, 20, 45, 79, 97, 121, 146, 169}
	_EnumIFCauseIndex_3 = [...]uint8{0, 18, 43, 78, 114, 148, 183, 207}
	_EnumIFCauseIndex_4 = [...]uint16{0, 19, 41, 57, 90, 125, 165, 194, 214, 228, 246, 265, 287, 316, 350, 388, 415, 436, 471, 507}
	_EnumIFCauseIndex_5 = [...]uint16{0, 28, 69, 103, 128, 195, 278, 309, 340, 360, 376, 438, 478, 514, 532, 563, 587, 636, 679}
	_EnumIFCauseIndex_6 = [...]uint16{0, 35, 60, 87, 107, 155, 203, 247, 294, 331, 350}
)

func (i EnumIFCause) String() string {
	switch {
	case 2 <= i && i <= 14:
		i -= 2
		return _EnumIFCauseName_0[_EnumIFCauseIndex_0[i]:_EnumIFCauseIndex_0[i+1]]
	case 16 <= i && i <= 19:
		i -= 16
		return _EnumIFCauseName_1[_EnumIFCauseIndex_1[i]:_EnumIFCauseIndex_1[i+1]]
	case 64 <= i && i <= 70:
		i -= 64
		return _EnumIFCauseName_2[_EnumIFCauseIndex_2[i]:_EnumIFCauseIndex_2[i+1]]
	case 72 <= i && i <= 78:
		i -= 72
		return _EnumIFCauseName_3[_EnumIFCauseIndex_3[i]:_EnumIFCauseIndex_3[i+1]]
	case 80 <= i && i <= 98:
		i -= 80
		return _EnumIFCauseName_4[_EnumIFCauseIndex_4[i]:_EnumIFCauseIndex_4[i+1]]
	case 100 <= i && i <= 117:
		i -= 100
		return _EnumIFCauseName_5[_EnumIFCauseIndex_5[i]:_EnumIFCauseIndex_5[i+1]]
	case 119 <= i && i <= 128:
		i -= 119
		return _EnumIFCauseName_6[_EnumIFCauseIndex_6[i]:_EnumIFCauseIndex_6[i+1]]
	default:
		return fmt.Sprintf("EnumIFCause(%d)", i)
	}
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _EnumIFCauseNoOp() {
	var x [1]struct{}
	_ = x[CauseLocalDetach-(2)]
	_ = x[CauseCompleteDetach-(3)]
	_ = x[CauseRATChangedFrom3GPPToNon3GPP-(4)]
	_ = x[CauseISRDeactivation-(5)]
	_ = x[CauseErrorIndicationReceivedFromRNCeNodeBS4SGSNMME-(6)]
	_ = x[CauseIMSIDetachOnly-(7)]
	_ = x[CauseReactivationRequested-(8)]
	_ = x[CausePDNReconnectionToThisAPNDisallowed-(9)]
	_ = x[CauseAccessChangedFromNon3GPPTo3GPP-(10)]
	_ = x[CausePDNConnectionInactivityTimerExpires-(11)]
	_ = x[CausePGWNotResponding-(12)]
	_ = x[CauseNetworkFailure-(13)]
	_ = x[CauseQoSParameterMismatch-(14)]
	_ = x[CauseRequestAccepted-(16)]
	_ = x[CauseRequestAcceptedPartially-(17)]
	_ = x[CauseNewPDNTypeDueToNetworkPreference-(18)]
	_ = x[CauseNewPDNTypeDueToSingleAddressBearerOnly-(19)]
	_ = x[CauseContextNotFound-(64)]
	_ = x[CauseInvalidMessageFormat-(65)]
	_ = x[CauseVersionNotSupportedByNextPeer-(66)]
	_ = x[CauseInvalidLength-(67)]
	_ = x[CauseServiceNotSupported-(68)]
	_ = x[CauseMandatoryIEIncorrect-(69)]
	_ = x[CauseMandatoryIEMissing-(70)]
	_ = x[CauseSystemFailure-(72)]
	_ = x[CauseNoResourcesAvailable-(73)]
	_ = x[CauseSemanticErrorInTheTFTOperation-(74)]
	_ = x[CauseSyntacticErrorInTheTFTOperation-(75)]
	_ = x[CauseSemanticErrorsInPacketFilters-(76)]
	_ = x[CauseSyntacticErrorsInPacketFilters-(77)]
	_ = x[CauseMissingOrUnknownAPN-(78)]
	_ = x[CauseGREKeyNotFound-(80)]
	_ = x[CauseRelocationFailure-(81)]
	_ = x[CauseDeniedInRAT-(82)]
	_ = x[CausePreferredPDNTypeNotSupported-(83)]
	_ = x[CauseAllDynamicAddressesAreOccupied-(84)]
	_ = x[CauseUEContextWithoutTFTAlreadyActivated-(85)]
	_ = x[CauseProtocolTypeNotSupported-(86)]
	_ = x[CauseUENotResponding-(87)]
	_ = x[CauseUERefuses-(88)]
	_ = x[CauseServiceDenied-(89)]
	_ = x[CauseUnableToPageUE-(90)]
	_ = x[CauseNoMemoryAvailable-(91)]
	_ = x[CauseUserAuthenticationFailed-(92)]
	_ = x[CauseAPNAccessDeniedNoSubscription-(93)]
	_ = x[CauseRequestRejectedReasonNotSpecified-(94)]
	_ = x[CausePTMSISignatureMismatch-(95)]
	_ = x[CauseIMSIIMEINotKnown-(96)]
	_ = x[CauseSemanticErrorInTheTADOperation-(97)]
	_ = x[CauseSyntacticErrorInTheTADOperation-(98)]
	_ = x[CauseRemotePeerNotResponding-(100)]
	_ = x[CauseCollisionWithNetworkInitiatedRequest-(101)]
	_ = x[CauseUnableToPageUEDueToSuspension-(102)]
	_ = x[CauseConditionalIEMissing-(103)]
	_ = x[CauseAPNRestrictionTypeIncompatibleWithCurrentlyActivePDNConnection-(104)]
	_ = x[CauseInvalidOverallLengthOfTheTriggeredResponseMessageAndAPiggybackedInitialMessage-(105)]
	_ = x[CauseDataForwardingNotSupported-(106)]
	_ = x[CauseInvalidReplyFromRemotePeer-(107)]
	_ = x[CauseFallbackToGTPv1-(108)]
	_ = x[CauseInvalidPeer-(109)]
	_ = x[CauseTemporarilyRejectedDueToHandoverTAURAUProcedureInProgress-(110)]
	_ = x[CauseModificationsNotLimitedToS1UBearers-(111)]
	_ = x[CauseRequestRejectedForAPMIPv6Reason-(112)]
	_ = x[CauseAPNCongestion-(113)]
	_ = x[CauseBearerHandlingNotSupported-(114)]
	_ = x[CauseUEAlreadyReattached-(115)]
	_ = x[CauseMultiplePDNConnectionsForAGivenAPNNotAllowed-(116)]
	_ = x[CauseTargetAccessRestrictedForTheSubscriber-(117)]
	_ = x[CauseMMESGSNRefusesDueToVPLMNPolicy-(119)]
	_ = x[CauseGTPCEntityCongestion-(120)]
	_ = x[CauseLateOverlappingRequest-(121)]
	_ = x[CauseTimedOutRequest-(122)]
	_ = x[CauseUEIsTemporarilyNotReachableDueToPowerSaving-(123)]
	_ = x[CauseRelocationFailureDueToNASMessageRedirection-(124)]
	_ = x[CauseUENotAuthorisedByOCSOrExternalAAAServer-(125)]
	_ = x[CauseMultipleAccessesToAPDNConnectionNotAllowed-(126)]
	_ = x[CauseRequestRejectedDueToUECapability-(127)]
	_ = x[CauseS1UPathFailure-(128)]
}

var _EnumIFCauseValues = []EnumIFCause{CauseLocalDetach, CauseCompleteDetach, CauseRATChangedFrom3GPPToNon3GPP, CauseISRDeactivation, CauseErrorIndicationReceivedFromRNCeNodeBS4SGSNMME, CauseIMSIDetachOnly, CauseReactivationRequested, CausePDNReconnectionToThisAPNDisallowed, CauseAccessChangedFromNon3GPPTo3GPP, CausePDNConnectionInactivityTimerExpires, CausePGWNotResponding, CauseNetworkFailure, CauseQoSParameterMismatch, CauseRequestAccepted, CauseRequestAcceptedPartially, CauseNewPDNTypeDueToNetworkPreference, CauseNewPDNTypeDueToSingleAddressBearerOnly, CauseContextNotFound, CauseInvalidMessageFormat, CauseVersionNotSupportedByNextPeer, CauseInvalidLength, CauseServiceNotSupported, CauseMandatoryIEIncorrect, CauseMandatoryIEMissing, CauseSystemFailure, CauseNoResourcesAvailable, CauseSemanticErrorInTheTFTOperation, CauseSyntacticErrorInTheTFTOperation, CauseSemanticErrorsInPacketFilters, CauseSyntacticErrorsInPacketFilters, CauseMissingOrUnknownAPN, CauseGREKeyNotFound, CauseRelocationFailure, CauseDeniedInRAT, CausePreferredPDNTypeNotSupported, CauseAllDynamicAddressesAreOccupied, CauseUEContextWithoutTFTAlreadyActivated, CauseProtocolTypeNotSupported, CauseUENotResponding, CauseUERefuses, CauseServiceDenied, CauseUnableToPageUE, CauseNoMemoryAvailable, CauseUserAuthenticationFailed, CauseAPNAccessDeniedNoSubscription, CauseRequestRejectedReasonNotSpecified, CausePTMSISignatureMismatch, CauseIMSIIMEINotKnown, CauseSemanticErrorInTheTADOperation, CauseSyntacticErrorInTheTADOperation, CauseRemotePeerNotResponding, CauseCollisionWithNetworkInitiatedRequest, CauseUnableToPageUEDueToSuspension, CauseConditionalIEMissing, CauseAPNRestrictionTypeIncompatibleWithCurrentlyActivePDNConnection, CauseInvalidOverallLengthOfTheTriggeredResponseMessageAndAPiggybackedInitialMessage, CauseDataForwardingNotSupported, CauseInvalidReplyFromRemotePeer, CauseFallbackToGTPv1, CauseInvalidPeer, CauseTemporarilyRejectedDueToHandoverTAURAUProcedureInProgress, CauseModificationsNotLimitedToS1UBearers, CauseRequestRejectedForAPMIPv6Reason, CauseAPNCongestion, CauseBearerHandlingNotSupported, CauseUEAlreadyReattached, CauseMultiplePDNConnectionsForAGivenAPNNotAllowed, CauseTargetAccessRestrictedForTheSubscriber, CauseMMESGSNRefusesDueToVPLMNPolicy, CauseGTPCEntityCongestion, CauseLateOverlappingRequest, CauseTimedOutRequest, CauseUEIsTemporarilyNotReachableDueToPowerSaving, CauseRelocationFailureDueToNASMessageRedirection, CauseUENotAuthorisedByOCSOrExternalAAAServer, CauseMultipleAccessesToAPDNConnectionNotAllowed, CauseRequestRejectedDueToUECapability, CauseS1UPathFailure}

var _EnumIFCauseNameToValueMap = map[string]EnumIFCause{
	_EnumIFCauseName_0[0:16]:         CauseLocalDetach,
	_EnumIFCauseLowerName_0[0:16]:    CauseLocalDetach,
	_EnumIFCauseName_0[16:35]:        CauseCompleteDetach,
	_EnumIFCauseLowerName_0[16:35]:   CauseCompleteDetach,
	_EnumIFCauseName_0[35:67]:        CauseRATChangedFrom3GPPToNon3GPP,
	_EnumIFCauseLowerName_0[35:67]:   CauseRATChangedFrom3GPPToNon3GPP,
	_EnumIFCauseName_0[67:87]:        CauseISRDeactivation,
	_EnumIFCauseLowerName_0[67:87]:   CauseISRDeactivation,
	_EnumIFCauseName_0[87:137]:       CauseErrorIndicationReceivedFromRNCeNodeBS4SGSNMME,
	_EnumIFCauseLowerName_0[87:137]:  CauseErrorIndicationReceivedFromRNCeNodeBS4SGSNMME,
	_EnumIFCauseName_0[137:156]:      CauseIMSIDetachOnly,
	_EnumIFCauseLowerName_0[137:156]: CauseIMSIDetachOnly,
	_EnumIFCauseName_0[156:182]:      CauseReactivationRequested,
	_EnumIFCauseLowerName_0[156:182]: CauseReactivationRequested,
	_EnumIFCauseName_0[182:221]:      CausePDNReconnectionToThisAPNDisallowed,
	_EnumIFCauseLowerName_0[182:221]: CausePDNReconnectionToThisAPNDisallowed,
	_EnumIFCauseName_0[221:256]:      CauseAccessChangedFromNon3GPPTo3GPP,
	_EnumIFCauseLowerName_0[221:256]: CauseAccessChangedFromNon3GPPTo3GPP,
	_EnumIFCauseName_0[256:296]:      CausePDNConnectionInactivityTimerExpires,
	_EnumIFCauseLowerName_0[256:296]: CausePDNConnectionInactivityTimerExpires,
	_EnumIFCauseName_0[296:317]:      CausePGWNotResponding,
	_EnumIFCauseLowerName_0[296:317]: CausePGWNotResponding,
	_EnumIFCauseName_0[317:336]:      CauseNetworkFailure,
	_EnumIFCauseLowerName_0[317:336]: CauseNetworkFailure,
	_EnumIFCauseName_0[336:361]:      CauseQoSParameterMismatch,
	_EnumIFCauseLowerName_0[336:361]: CauseQoSParameterMismatch,
	_EnumIFCauseName_1[0:20]:         CauseRequestAccepted,
	_EnumIFCauseLowerName_1[0:20]:    CauseRequestAccepted,
	_EnumIFCauseName_1[20:49]:        CauseRequestAcceptedPartially,
	_EnumIFCauseLowerName_1[20:49]:   CauseRequestAcceptedPartially,
	_EnumIFCauseName_1[49:86]:        CauseNewPDNTypeDueToNetworkPreference,
	_EnumIFCauseLowerName_1[49:86]:   CauseNewPDNTypeDueToNetworkPreference,
	_EnumIFCauseName_1[86:129]:       CauseNewPDNTypeDueToSingleAddressBearerOnly,
	_EnumIFCauseLowerName_1[86:129]:  CauseNewPDNTypeDueToSingleAddressBearerOnly,
	_EnumIFCauseName_2[0:20]:         CauseContextNotFound,
	_EnumIFCauseLowerName_2[0:20]:    CauseContextNotFound,
	_EnumIFCauseName_2[20:45]:        CauseInvalidMessageFormat,
	_EnumIFCauseLowerName_2[20:45]:   CauseInvalidMessageFormat,
	_EnumIFCauseName_2[45:79]:        CauseVersionNotSupportedByNextPeer,
	_EnumIFCauseLowerName_2[45:79]:   CauseVersionNotSupportedByNextPeer,
	_EnumIFCauseName_2[79:97]:        CauseInvalidLength,
	_EnumIFCauseLowerName_2[79:97]:   CauseInvalidLength,
	_EnumIFCauseName_2[97:121]:       CauseServiceNotSupported,
	_EnumIFCauseLowerName_2[97:121]:  CauseServiceNotSupported,
	_EnumIFCauseName_2[121:146]:      CauseMandatoryIEIncorrect,
	_EnumIFCauseLowerName_2[121:146]: CauseMandatoryIEIncorrect,
	_EnumIFCauseName_2[146:169]:      CauseMandatoryIEMissing,
	_EnumIFCauseLowerName_2[146:169]: CauseMandatoryIEMissing,
	_EnumIFCauseName_3[0:18]:         CauseSystemFailure,
	_EnumIFCauseLowerName_3[0:18]:    CauseSystemFailure,
	_EnumIFCauseName_3[18:43]:        CauseNoResourcesAvailable,
	_EnumIFCauseLowerName_3[18:43]:   CauseNoResourcesAvailable,
	_EnumIFCauseName_3[43:78]:        CauseSemanticErrorInTheTFTOperation,
	_EnumIFCauseLowerName_3[43:78]:   CauseSemanticErrorInTheTFTOperation,
	_EnumIFCauseName_3[78:114]:       CauseSyntacticErrorInTheTFTOperation,
	_EnumIFCauseLowerName_3[78:114]:  CauseSyntacticErrorInTheTFTOperation,
	_EnumIFCauseName_3[114:148]:      CauseSemanticErrorsInPacketFilters,
	_EnumIFCauseLowerName_3[114:148]: CauseSemanticErrorsInPacketFilters,
	_EnumIFCauseName_3[148:183]:      CauseSyntacticErrorsInPacketFilters,
	_EnumIFCauseLowerName_3[148:183]: CauseSyntacticErrorsInPacketFilters,
	_EnumIFCauseName_3[183:207]:      CauseMissingOrUnknownAPN,
	_EnumIFCauseLowerName_3[183:207]: CauseMissingOrUnknownAPN,
	_EnumIFCauseName_4[0:19]:         CauseGREKeyNotFound,
	_EnumIFCauseLowerName_4[0:19]:    CauseGREKeyNotFound,
	_EnumIFCauseName_4[19:41]:        CauseRelocationFailure,
	_EnumIFCauseLowerName_4[19:41]:   CauseRelocationFailure,
	_EnumIFCauseName_4[41:57]:        CauseDeniedInRAT,
	_EnumIFCauseLowerName_4[41:57]:   CauseDeniedInRAT,
	_EnumIFCauseName_4[57:90]:        CausePreferredPDNTypeNotSupported,
	_EnumIFCauseLowerName_4[57:90]:   CausePreferredPDNTypeNotSupported,
	_EnumIFCauseName_4[90:125]:       CauseAllDynamicAddressesAreOccupied,
	_EnumIFCauseLowerName_4[90:125]:  CauseAllDynamicAddressesAreOccupied,
	_EnumIFCauseName_4[125:165]:      CauseUEContextWithoutTFTAlreadyActivated,
	_EnumIFCauseLowerName_4[125:165]: CauseUEContextWithoutTFTAlreadyActivated,
	_EnumIFCauseName_4[165:194]:      CauseProtocolTypeNotSupported,
	_EnumIFCauseLowerName_4[165:194]: CauseProtocolTypeNotSupported,
	_EnumIFCauseName_4[194:214]:      CauseUENotResponding,
	_EnumIFCauseLowerName_4[194:214]: CauseUENotResponding,
	_EnumIFCauseName_4[214:228]:      CauseUERefuses,
	_EnumIFCauseLowerName_4[214:228]: CauseUERefuses,
	_EnumIFCauseName_4[228:246]:      CauseServiceDenied,
	_EnumIFCauseLowerName_4[228:246]: CauseServiceDenied,
	_EnumIFCauseName_4[246:265]:      CauseUnableToPageUE,
	_EnumIFCauseLowerName_4[246:265]: CauseUnableToPageUE,
	_EnumIFCauseName_4[265:287]:      CauseNoMemoryAvailable,
	_EnumIFCauseLowerName_4[265:287]: CauseNoMemoryAvailable,
	_EnumIFCauseName_4[287:316]:      CauseUserAuthenticationFailed,
	_EnumIFCauseLowerName_4[287:316]: CauseUserAuthenticationFailed,
	_EnumIFCauseName_4[316:350]:      CauseAPNAccessDeniedNoSubscription,
	_EnumIFCauseLowerName_4[316:350]: CauseAPNAccessDeniedNoSubscription,
	_EnumIFCauseName_4[350:388]:      CauseRequestRejectedReasonNotSpecified,
	_EnumIFCauseLowerName_4[350:388]: CauseRequestRejectedReasonNotSpecified,
	_EnumIFCauseName_4[388:415]:      CausePTMSISignatureMismatch,
	_EnumIFCauseLowerName_4[388:415]: CausePTMSISignatureMismatch,
	_EnumIFCauseName_4[415:436]:      CauseIMSIIMEINotKnown,
	_EnumIFCauseLowerName_4[415:436]: CauseIMSIIMEINotKnown,
	_EnumIFCauseName_4[436:471]:      CauseSemanticErrorInTheTADOperation,
	_EnumIFCauseLowerName_4[436:471]: CauseSemanticErrorInTheTADOperation,
	_EnumIFCauseName_4[471:507]:      CauseSyntacticErrorInTheTADOperation,
	_EnumIFCauseLowerName_4[471:507]: CauseSyntacticErrorInTheTADOperation,
	_EnumIFCauseName_5[0:28]:         CauseRemotePeerNotResponding,
	_EnumIFCauseLowerName_5[0:28]:    CauseRemotePeerNotResponding,
	_EnumIFCauseName_5[28:69]:        CauseCollisionWithNetworkInitiatedRequest,
	_EnumIFCauseLowerName_5[28:69]:   CauseCollisionWithNetworkInitiatedRequest,
	_EnumIFCauseName_5[69:103]:       CauseUnableToPageUEDueToSuspension,
	_EnumIFCauseLowerName_5[69:103]:  CauseUnableToPageUEDueToSuspension,
	_EnumIFCauseName_5[103:128]:      CauseConditionalIEMissing,
	_EnumIFCauseLowerName_5[103:128]: CauseConditionalIEMissing,
	_EnumIFCauseName_5[128:195]:      CauseAPNRestrictionTypeIncompatibleWithCurrentlyActivePDNConnection,
	_EnumIFCauseLowerName_5[128:195]: CauseAPNRestrictionTypeIncompatibleWithCurrentlyActivePDNConnection,
	_EnumIFCauseName_5[195:278]:      CauseInvalidOverallLengthOfTheTriggeredResponseMessageAndAPiggybackedInitialMessage,
	_EnumIFCauseLowerName_5[195:278]: CauseInvalidOverallLengthOfTheTriggeredResponseMessageAndAPiggybackedInitialMessage,
	_EnumIFCauseName_5[278:309]:      CauseDataForwardingNotSupported,
	_EnumIFCauseLowerName_5[278:309]: CauseDataForwardingNotSupported,
	_EnumIFCauseName_5[309:340]:      CauseInvalidReplyFromRemotePeer,
	_EnumIFCauseLowerName_5[309:340]: CauseInvalidReplyFromRemotePeer,
	_EnumIFCauseName_5[340:360]:      CauseFallbackToGTPv1,
	_EnumIFCauseLowerName_5[340:360]: CauseFallbackToGTPv1,
	_EnumIFCauseName_5[360:376]:      CauseInvalidPeer,
	_EnumIFCauseLowerName_5[360:376]: CauseInvalidPeer,
	_EnumIFCauseName_5[376:438]:      CauseTemporarilyRejectedDueToHandoverTAURAUProcedureInProgress,
	_EnumIFCauseLowerName_5[376:438]: CauseTemporarilyRejectedDueToHandoverTAURAUProcedureInProgress,
	_EnumIFCauseName_5[438:478]:      CauseModificationsNotLimitedToS1UBearers,
	_EnumIFCauseLowerName_5[438:478]: CauseModificationsNotLimitedToS1UBearers,
	_EnumIFCauseName_5[478:514]:      CauseRequestRejectedForAPMIPv6Reason,
	_EnumIFCauseLowerName_5[478:514]: CauseRequestRejectedForAPMIPv6Reason,
	_EnumIFCauseName_5[514:532]:      CauseAPNCongestion,
	_EnumIFCauseLowerName_5[514:532]: CauseAPNCongestion,
	_EnumIFCauseName_5[532:563]:      CauseBearerHandlingNotSupported,
	_EnumIFCauseLowerName_5[532:563]: CauseBearerHandlingNotSupported,
	_EnumIFCauseName_5[563:587]:      CauseUEAlreadyReattached,
	_EnumIFCauseLowerName_5[563:587]: CauseUEAlreadyReattached,
	_EnumIFCauseName_5[587:636]:      CauseMultiplePDNConnectionsForAGivenAPNNotAllowed,
	_EnumIFCauseLowerName_5[587:636]: CauseMultiplePDNConnectionsForAGivenAPNNotAllowed,
	_EnumIFCauseName_5[636:679]:      CauseTargetAccessRestrictedForTheSubscriber,
	_EnumIFCauseLowerName_5[636:679]: CauseTargetAccessRestrictedForTheSubscriber,
	_EnumIFCauseName_6[0:35]:         CauseMMESGSNRefusesDueToVPLMNPolicy,
	_EnumIFCauseLowerName_6[0:35]:    CauseMMESGSNRefusesDueToVPLMNPolicy,
	_EnumIFCauseName_6[35:60]:        CauseGTPCEntityCongestion,
	_EnumIFCauseLowerName_6[35:60]:   CauseGTPCEntityCongestion,
	_EnumIFCauseName_6[60:87]:        CauseLateOverlappingRequest,
	_EnumIFCauseLowerName_6[60:87]:   CauseLateOverlappingRequest,
	_EnumIFCauseName_6[87:107]:       CauseTimedOutRequest,
	_EnumIFCauseLowerName_6[87:107]:  CauseTimedOutRequest,
	_EnumIFCauseName_6[107:155]:      CauseUEIsTemporarilyNotReachableDueToPowerSaving,
	_EnumIFCauseLowerName_6[107:155]: CauseUEIsTemporarilyNotReachableDueToPowerSaving,
	_EnumIFCauseName_6[155:203]:      CauseRelocationFailureDueToNASMessageRedirection,
	_EnumIFCauseLowerName_6[155:203]: CauseRelocationFailureDueToNASMessageRedirection,
	_EnumIFCauseName_6[203:247]:      CauseUENotAuthorisedByOCSOrExternalAAAServer,
	_EnumIFCauseLowerName_6[203:247]: CauseUENotAuthorisedByOCSOrExternalAAAServer,
	_EnumIFCauseName_6[247:294]:      CauseMultipleAccessesToAPDNConnectionNotAllowed,
	_EnumIFCauseLowerName_6[247:294]: CauseMultipleAccessesToAPDNConnectionNotAllowed,
	_EnumIFCauseName_6[294:331]:      CauseRequestRejectedDueToUECapability,
	_EnumIFCauseLowerName_6[294:331]: CauseRequestRejectedDueToUECapability,
	_EnumIFCauseName_6[331:350]:      CauseS1UPathFailure,
	_EnumIFCauseLowerName_6[331:350]: CauseS1UPathFailure,
}

var _EnumIFCauseNames = []string{
	_EnumIFCauseName_0[0:16],
	_EnumIFCauseName_0[16:35],
	_EnumIFCauseName_0[35:67],
	_EnumIFCauseName_0[67:87],
	_EnumIFCauseName_0[87:137],
	_EnumIFCauseName_0[137:156],
	_EnumIFCauseName_0[156:182],
	_EnumIFCauseName_0[182:221],
	_EnumIFCauseName_0[221:256],
	_EnumIFCauseName_0[256:296],
	_EnumIFCauseName_0[296:317],
	_EnumIFCauseName_0[317:336],
	_EnumIFCauseName_0[336:361],
	_EnumIFCauseName_1[0:20],
	_EnumIFCauseName_1[20:49],
	_EnumIFCauseName_1[49:86],
	_EnumIFCauseName_1[86:129],
	_EnumIFCauseName_2[0:20],
	_EnumIFCauseName_2[20:45],
	_EnumIFCauseName_2[45:79],
	_EnumIFCauseName_2[79:97],
	_EnumIFCauseName_2[97:121],
	_EnumIFCauseName_2[121:146],
	_EnumIFCauseName_2[146:169],
	_EnumIFCauseName_3[0:18],
	_EnumIFCauseName_3[18:43],
	_EnumIFCauseName_3[43:78],
	_EnumIFCauseName_3[78:114],
	_EnumIFCauseName_3[114:148],
	_EnumIFCauseName_3[148:183],
	_EnumIFCauseName_3[183:207],
	_EnumIFCauseName_4[0:19],
	_EnumIFCauseName_4[19:41],
	_EnumIFCauseName_4[41:57],
	_EnumIFCauseName_4[57:90],
	_EnumIFCauseName_4[90:125],
	_EnumIFCauseName_4[125:165],
	_EnumIFCauseName_4[165:194],
	_EnumIFCauseName_4[194:214],
	_EnumIFCauseName_4[214:228],
	_EnumIFCauseName_4[228:246],
	_EnumIFCauseName_4[246:265],
	_EnumIFCauseName_4[265:287],
	_EnumIFCauseName_4[287:316],
	_EnumIFCauseName_4[316:350],
	_EnumIFCauseName_4[350:388],
	_EnumIFCauseName_4[388:415],
	_EnumIFCauseName_4[415:436],
	_EnumIFCauseName_4[436:471],
	_EnumIFCauseName_4[471:507],
	_EnumIFCauseName_5[0:28],
	_EnumIFCauseName_5[28:69],
	_EnumIFCauseName_5[69:103],
	_EnumIFCauseName_5[103:128],
	_EnumIFCauseName_5[128:195],
	_EnumIFCauseName_5[195:278],
	_EnumIFCauseName_5[278:309],
	_EnumIFCauseName_5[309:340],
	_EnumIFCauseName_5[340:360],
	_EnumIFCauseName_5[360:376],
	_EnumIFCauseName_5[376:438],
	_EnumIFCauseName_5[438:478],
	_EnumIFCauseName_5[478:514],
	_EnumIFCauseName_5[514:532],
	_EnumIFCauseName_5[532:563],
	_EnumIFCauseName_5[563:587],
	_EnumIFCauseName_5[587:636],
	_EnumIFCauseName_5[636:679],
	_EnumIFCauseName_6[0:35],
	_EnumIFCauseName_6[35:60],
	_EnumIFCauseName_6[60:87],
	_EnumIFCauseName_6[87:107],
	_EnumIFCauseName_6[107:155],
	_EnumIFCauseName_6[155:203],
	_EnumIFCauseName_6[203:247],
	_EnumIFCauseName_6[247:294],
	_EnumIFCauseName_6[294:331],
	_EnumIFCauseName_6[331:350],
}

// EnumIFCauseString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func EnumIFCauseString(s string) (EnumIFCause, error) {
	if val, ok := _EnumIFCauseNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _EnumIFCauseNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to EnumIFCause values", s)
}

// EnumIFCauseValues returns all values of the enum
func EnumIFCauseValues() []EnumIFCause {
	return _EnumIFCauseValues
}

// EnumIFCauseStrings returns a slice of all String values of the enum
func EnumIFCauseStrings() []string {
	strs := make([]string, len(_EnumIFCauseNames))
	copy(strs, _EnumIFCauseNames)
	return strs
}

// IsAEnumIFCause returns "true" if the value is listed in the enum definition. "false" otherwise
func (i EnumIFCause) IsAEnumIFCause() bool {
	for _, v := range _EnumIFCauseValues {
		if i == v {
			return true
		}
	}
	return false
}
