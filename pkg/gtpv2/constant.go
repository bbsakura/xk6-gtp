package gtpv2

// cf. https://github.com/wmnsk/go-gtp/blob/master/gtpv2/constants.go#L24-L65
//
//go:generate enumer -type=EnumIFType
type EnumIFType uint8

const (
	IFTypeS1UeNodeBGTPU EnumIFType = iota
	IFTypeS1USGWGTPU
	IFTypeS12RNCGTPU
	IFTypeS12SGWGTPU
	IFTypeS5S8SGWGTPU
	IFTypeS5S8PGWGTPU
	IFTypeS5S8SGWGTPC
	IFTypeS5S8PGWGTPC
	IFTypeS5S8SGWPMIPv6
	IFTypeS5S8PGWPMIPv6
	IFTypeS11MMEGTPC
	IFTypeS11S4SGWGTPC
	IFTypeS10MMEGTPC
	IFTypeS3MMEGTPC
	IFTypeS3SGSNGTPC
	IFTypeS4SGSNGTPU
	IFTypeS4SGWGTPU
	IFTypeS4SGSNGTPC
	IFTypeS16SGSNGTPC
	IFTypeeNodeBGTPUForDL
	IFTypeeNodeBGTPUForUL
	IFTypeRNCGTPUForData
	IFTypeSGSNGTPUForData
	IFTypeSGWUPFGTPUForDL
	IFTypeSmMBMSGWGTPC
	IFTypeSnMBMSGWGTPC
	IFTypeSmMMEGTPC
	IFTypeSnSGSNGTPC
	IFTypeSGWGTPUForUL
	IFTypeSnSGSNGTPU
	IFTypeS2bePDGGTPC
	IFTypeS2bUePDGGTPU
	IFTypeS2bPGWGTPC
	IFTypeS2bUPGWGTPU
	IFTypeS2aTWANGTPU
	IFTypeS2aTWANGTPC
	IFTypeS2aPGWGTPC
	IFTypeS2aPGWGTPU
	IFTypeS11MMEGTPU
	IFTypeS11SGWGTPU
)

//go:generate enumer -type=EnumIFCause
type EnumIFCause uint8

const (
	_                                                                                   EnumIFCause = 0
	_                                                                                   EnumIFCause = 1
	CauseLocalDetach                                                                    EnumIFCause = 2
	CauseCompleteDetach                                                                 EnumIFCause = 3
	CauseRATChangedFrom3GPPToNon3GPP                                                    EnumIFCause = 4
	CauseISRDeactivation                                                                EnumIFCause = 5
	CauseErrorIndicationReceivedFromRNCeNodeBS4SGSNMME                                  EnumIFCause = 6
	CauseIMSIDetachOnly                                                                 EnumIFCause = 7
	CauseReactivationRequested                                                          EnumIFCause = 8
	CausePDNReconnectionToThisAPNDisallowed                                             EnumIFCause = 9
	CauseAccessChangedFromNon3GPPTo3GPP                                                 EnumIFCause = 10
	CausePDNConnectionInactivityTimerExpires                                            EnumIFCause = 11
	CausePGWNotResponding                                                               EnumIFCause = 12
	CauseNetworkFailure                                                                 EnumIFCause = 13
	CauseQoSParameterMismatch                                                           EnumIFCause = 14
	_                                                                                   EnumIFCause = 15
	CauseRequestAccepted                                                                EnumIFCause = 16
	CauseRequestAcceptedPartially                                                       EnumIFCause = 17
	CauseNewPDNTypeDueToNetworkPreference                                               EnumIFCause = 18
	CauseNewPDNTypeDueToSingleAddressBearerOnly                                         EnumIFCause = 19
	CauseContextNotFound                                                                EnumIFCause = 64
	CauseInvalidMessageFormat                                                           EnumIFCause = 65
	CauseVersionNotSupportedByNextPeer                                                  EnumIFCause = 66
	CauseInvalidLength                                                                  EnumIFCause = 67
	CauseServiceNotSupported                                                            EnumIFCause = 68
	CauseMandatoryIEIncorrect                                                           EnumIFCause = 69
	CauseMandatoryIEMissing                                                             EnumIFCause = 70
	_                                                                                   EnumIFCause = 71
	CauseSystemFailure                                                                  EnumIFCause = 72
	CauseNoResourcesAvailable                                                           EnumIFCause = 73
	CauseSemanticErrorInTheTFTOperation                                                 EnumIFCause = 74
	CauseSyntacticErrorInTheTFTOperation                                                EnumIFCause = 75
	CauseSemanticErrorsInPacketFilters                                                  EnumIFCause = 76
	CauseSyntacticErrorsInPacketFilters                                                 EnumIFCause = 77
	CauseMissingOrUnknownAPN                                                            EnumIFCause = 78
	_                                                                                   EnumIFCause = 79
	CauseGREKeyNotFound                                                                 EnumIFCause = 80
	CauseRelocationFailure                                                              EnumIFCause = 81
	CauseDeniedInRAT                                                                    EnumIFCause = 82
	CausePreferredPDNTypeNotSupported                                                   EnumIFCause = 83
	CauseAllDynamicAddressesAreOccupied                                                 EnumIFCause = 84
	CauseUEContextWithoutTFTAlreadyActivated                                            EnumIFCause = 85
	CauseProtocolTypeNotSupported                                                       EnumIFCause = 86
	CauseUENotResponding                                                                EnumIFCause = 87
	CauseUERefuses                                                                      EnumIFCause = 88
	CauseServiceDenied                                                                  EnumIFCause = 89
	CauseUnableToPageUE                                                                 EnumIFCause = 90
	CauseNoMemoryAvailable                                                              EnumIFCause = 91
	CauseUserAuthenticationFailed                                                       EnumIFCause = 92
	CauseAPNAccessDeniedNoSubscription                                                  EnumIFCause = 93
	CauseRequestRejectedReasonNotSpecified                                              EnumIFCause = 94
	CausePTMSISignatureMismatch                                                         EnumIFCause = 95
	CauseIMSIIMEINotKnown                                                               EnumIFCause = 96
	CauseSemanticErrorInTheTADOperation                                                 EnumIFCause = 97
	CauseSyntacticErrorInTheTADOperation                                                EnumIFCause = 98
	_                                                                                   EnumIFCause = 99
	CauseRemotePeerNotResponding                                                        EnumIFCause = 100
	CauseCollisionWithNetworkInitiatedRequest                                           EnumIFCause = 101
	CauseUnableToPageUEDueToSuspension                                                  EnumIFCause = 102
	CauseConditionalIEMissing                                                           EnumIFCause = 103
	CauseAPNRestrictionTypeIncompatibleWithCurrentlyActivePDNConnection                 EnumIFCause = 104
	CauseInvalidOverallLengthOfTheTriggeredResponseMessageAndAPiggybackedInitialMessage EnumIFCause = 105
	CauseDataForwardingNotSupported                                                     EnumIFCause = 106
	CauseInvalidReplyFromRemotePeer                                                     EnumIFCause = 107
	CauseFallbackToGTPv1                                                                EnumIFCause = 108
	CauseInvalidPeer                                                                    EnumIFCause = 109
	CauseTemporarilyRejectedDueToHandoverTAURAUProcedureInProgress                      EnumIFCause = 110
	CauseModificationsNotLimitedToS1UBearers                                            EnumIFCause = 111
	CauseRequestRejectedForAPMIPv6Reason                                                EnumIFCause = 112
	CauseAPNCongestion                                                                  EnumIFCause = 113
	CauseBearerHandlingNotSupported                                                     EnumIFCause = 114
	CauseUEAlreadyReattached                                                            EnumIFCause = 115
	CauseMultiplePDNConnectionsForAGivenAPNNotAllowed                                   EnumIFCause = 116
	CauseTargetAccessRestrictedForTheSubscriber                                         EnumIFCause = 117
	_                                                                                   EnumIFCause = 118
	CauseMMESGSNRefusesDueToVPLMNPolicy                                                 EnumIFCause = 119
	CauseGTPCEntityCongestion                                                           EnumIFCause = 120
	CauseLateOverlappingRequest                                                         EnumIFCause = 121
	CauseTimedOutRequest                                                                EnumIFCause = 122
	CauseUEIsTemporarilyNotReachableDueToPowerSaving                                    EnumIFCause = 123
	CauseRelocationFailureDueToNASMessageRedirection                                    EnumIFCause = 124
	CauseUENotAuthorisedByOCSOrExternalAAAServer                                        EnumIFCause = 125
	CauseMultipleAccessesToAPDNConnectionNotAllowed                                     EnumIFCause = 126
	CauseRequestRejectedDueToUECapability                                               EnumIFCause = 127
	CauseS1UPathFailure                                                                 EnumIFCause = 128
)
