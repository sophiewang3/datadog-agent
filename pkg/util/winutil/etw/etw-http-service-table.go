// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build windows && npm

package etw

//revive:disable:var-naming Name is intended to match the Windows const name

// Windows ETW HTTP Provider consts
// From https://github.com/repnz/etw-providers-docs/blob/master/Manifests-Win10-18990/Microsoft-Windows-HttpService.xml
const (
	IDHTTPDummyStart uint16 = iota
	IDHTTPRequestTraceTaskRecvReq
	IDHTTPRequestTraceTaskParse
	IDHTTPRequestTraceTaskDeliver
	IDHTTPRequestTraceTaskRecvResp
	IDHTTPRequestTraceTaskRecvRespLast
	IDHTTPRequestTraceTaskRecvBody
	IDHTTPRequestTraceTaskRecvBodyLast
	IDHTTPRequestTraceTaskFastResp
	IDHTTPRequestTraceTaskFastRespLast
	IDHTTPRequestTraceTaskSendComplete
	IDHTTPRequestTraceTaskCachedAndSend
	IDHTTPRequestTraceTaskFastSend
	IDHTTPRequestTraceTaskZeroSend
	IDHTTPRequestTraceTaskLastSndError
	IDHTTPRequestTraceTaskSndError
	IDHTTPRequestTraceTaskSrvdFrmCache
	IDHTTPRequestTraceTaskCachedNotModified
	IDHTTPSetupTraceTaskResvUrl
	IDHTTPSetupTraceTaskReadIpListEntry
	IDHTTPSetupTraceTaskCreatedSslCred
	IDHTTPConnectionTraceTaskConnConnect
	IDHTTPConnectionTraceTaskConnIdAssgn
	IDHTTPConnectionTraceTaskConnClose
	IDHTTPConnectionTraceTaskConnCleanup
	IDHTTPCacheTraceTaskAddedCacheEntry
	IDHTTPCacheTraceTaskAddCacheEntryFailed
	IDHTTPCacheTraceTaskFlushedCache
	IDHTTPConfigurationPropertyTraceTaskChgUrlGrpProp
	IDHTTPConfigurationPropertyTraceTaskChgSrvSesProp
	IDHTTPConfigurationPropertyTraceTaskChgReqQueueProp
	IDHTTPConfigurationPropertyTraceTaskAddUrl
	IDHTTPConfigurationPropertyTraceTaskRemUrl
	IDHTTPConfigurationPropertyTraceTaskRemAllUrls
	IDHTTPSSLTraceTaskSslConnEvent
	IDHTTPSSLTraceTaskSslInitiateHandshake
	IDHTTPSSLTraceTaskSslHandshakeComplete
	IDHTTPSSLTraceTaskSslInititateSslRcvClientCert
	IDHTTPSSLTraceTaskSslRcvClientCertFailed
	IDHTTPSSLTraceTaskSslRcvdRawData
	IDHTTPSSLTraceTaskSslDlvrdStreamData
	IDHTTPSSLTraceTaskSslAcceptStreamData
	IDHTTPAuthenticationTraceTaskSspiCall
	IDHTTPAuthenticationTraceTaskAuthCacheEntryAdded
	IDHTTPAuthenticationTraceTaskAuthCacheEntryFreed
	IDHTTPConnectionTraceTaskQosFlowSetReset
	IDHTTPLoggingTraceTaskLoggingConfigFailed
	IDHTTPLoggingTraceTaskLoggingConfig
	IDHTTPLoggingTraceTaskLogFileCreateFailed
	IDHTTPLoggingTraceTaskLogFileCreate
	IDHTTPLoggingTraceTaskLogFileWrite
	IDHTTPRequestTraceTaskParseRequestFailed
	IDHTTPTimeoutTraceTaskConnTimedOut
	IDHTTPSSLTraceTaskSslEndpointCreationFailed
	IDHTTPSSLTraceTaskSslDisconnEvent
	IDHTTPSSLTraceTaskSslDisconnReq
	IDHTTPSSLTraceTaskSslUnsealMsg
	IDHTTPSSLTraceTaskSslQueryConnInfoFailed
	IDHTTPSSLTraceTaskSslEndpointConfigNotFound
	IDHTTPSSLTraceTaskSslAsc
	IDHTTPSSLTraceTaskSslSealMsg
	IDHTTPRequestTraceTaskRequestRejected
	IDHTTPRequestTraceTaskRequestCancelled
	IDHTTPDriverGlobalSettingsTaskHotAddProcFailed
	IDHTTPDriverGlobalSettingsTaskHotAddProcSucceeded
	IDHTTPRequestTraceTaskUserResponseFlowInit
	IDHTTPRequestTraceTaskCachedResponseFlowInit
	IDHTTPRequestTraceTaskFlowInitFailed
	IDHTTPConnectionTraceTaskSetConnectionFlow
	IDHTTPConnectionTraceTaskRequestAssociatedToConfigurationFlow
	IDHTTPConnectionTraceTaskConnectionFlowFailed
	IDHTTPRequestTraceTaskResponseRangeProcessingOK
	IDHTTPCacheTraceTaskBeginBuildingSlices
	IDHTTPCacheTraceTaskSendSliceCacheContent
	IDHTTPCacheTraceTaskCachedSlicesMatchContent
	IDHTTPCacheTraceTaskMergeSlicesToCache
	IDHTTPCacheTraceTaskFlatCacheRangeSend
	IDHTTPAuthenticationTraceTaskChannelBindAscParams
	IDHTTPAuthenticationTraceTaskServiceBindCheckComplete
	IDHTTPAuthenticationTraceTaskChannelBindConfigCapture
	IDHTTPAuthenticationTraceTaskChannelBindPerResponseConfig
	IDHTTPConnectionTraceTaskUsePolicyBasedQoSFlow
	IDHTTPThreadPoolThreadPoolExtension
	IDHTTPThreadPoolThreadReady
	IDHTTPThreadPoolThreadPoolTrim
	IDHTTPThreadPoolThreadGone
	IDHTTPSSLTraceTaskSniParsed
	IDHTTPRequestTraceTaskInitiateOpaqueMode
	IDHTTPSSLTraceTaskEndpointAutoGenerated
	IDHTTPSSLTraceTaskAutoGeneratedEndpointDeleted
	IDHTTPSSLTraceTaskSslEndpointConfigFound
	IDHTTPSSLTraceTaskSslEndpointConfigRejected
	IDHTTPResponseTraceTaskParseRequestFailed
	IDHTTPSSLTraceTaskSslHandshakeFailure
	IDHTTPRequestTraceTaskHttpErrorResponseSent
	IDHTTPSSLTraceTaskSslRenegotiateTimedOut
	IDHTTPRequestTraceTaskHttp11Required
	IDHTTPConnectionTraceTaskQuicConnection
	IDHTTPConnectionTraceTaskQuicConnectionCallback
	IDHTTPConnectionTraceTaskQuicStream
	IDHTTPConnectionTraceTaskQuicStreamCallback
	IDHTTPDriverGlobalSettingsTaskQuicRegistration
	IDHTTPDummyMax
)

//revive:enable:var-naming (const)

var (
	httpServiceEventID2Name = []string{
		"HTTP_DUMMY_START",
		"HTTPRequestTraceTaskRecvReq",
		"HTTPRequestTraceTaskParse",
		"HTTPRequestTraceTaskDeliver",
		"HTTPRequestTraceTaskRecvResp",
		"HTTPRequestTraceTaskRecvRespLast",
		"HTTPRequestTraceTaskRecvBody",
		"HTTPRequestTraceTaskRecvBodyLast",
		"HTTPRequestTraceTaskFastResp",
		"HTTPRequestTraceTaskFastRespLast",
		"HTTPRequestTraceTaskSendComplete",
		"HTTPRequestTraceTaskCachedAndSend",
		"HTTPRequestTraceTaskFastSend",
		"HTTPRequestTraceTaskZeroSend",
		"HTTPRequestTraceTaskLastSndError",
		"HTTPRequestTraceTaskSndError",
		"HTTPRequestTraceTaskSrvdFrmCache",
		"HTTPRequestTraceTaskCachedNotModified",
		"HTTPSetupTraceTaskResvUrl",
		"HTTPSetupTraceTaskReadIpListEntry",
		"HTTPSetupTraceTaskCreatedSslCred",
		"HTTPConnectionTraceTaskConnConnect",
		"HTTPConnectionTraceTaskConnIdAssgn",
		"HTTPConnectionTraceTaskConnClose",
		"HTTPConnectionTraceTaskConnCleanup",
		"HTTPCacheTraceTaskAddedCacheEntry",
		"HTTPCacheTraceTaskAddCacheEntryFailed",
		"HTTPCacheTraceTaskFlushedCache",
		"HTTPConfigurationPropertyTraceTaskChgUrlGrpProp",
		"HTTPConfigurationPropertyTraceTaskChgSrvSesProp",
		"HTTPConfigurationPropertyTraceTaskChgReqQueueProp",
		"HTTPConfigurationPropertyTraceTaskAddUrl",
		"HTTPConfigurationPropertyTraceTaskRemUrl",
		"HTTPConfigurationPropertyTraceTaskRemAllUrls",
		"HTTPSSLTraceTaskSslConnEvent",
		"HTTPSSLTraceTaskSslInitiateHandshake",
		"HTTPSSLTraceTaskSslHandshakeComplete",
		"HTTPSSLTraceTaskSslInititateSslRcvClientCert",
		"HTTPSSLTraceTaskSslRcvClientCertFailed",
		"HTTPSSLTraceTaskSslRcvdRawData",
		"HTTPSSLTraceTaskSslDlvrdStreamData",
		"HTTPSSLTraceTaskSslAcceptStreamData",
		"HTTPAuthenticationTraceTaskSspiCall",
		"HTTPAuthenticationTraceTaskAuthCacheEntryAdded",
		"HTTPAuthenticationTraceTaskAuthCacheEntryFreed",
		"HTTPConnectionTraceTaskQosFlowSetReset",
		"HTTPLoggingTraceTaskLoggingConfigFailed",
		"HTTPLoggingTraceTaskLoggingConfig",
		"HTTPLoggingTraceTaskLogFileCreateFailed",
		"HTTPLoggingTraceTaskLogFileCreate",
		"HTTPLoggingTraceTaskLogFileWrite",
		"HTTPRequestTraceTaskParseRequestFailed",
		"HTTPTimeoutTraceTaskConnTimedOut",
		"HTTPSSLTraceTaskSslEndpointCreationFailed",
		"HTTPSSLTraceTaskSslDisconnEvent",
		"HTTPSSLTraceTaskSslDisconnReq",
		"HTTPSSLTraceTaskSslUnsealMsg",
		"HTTPSSLTraceTaskSslQueryConnInfoFailed",
		"HTTPSSLTraceTaskSslEndpointConfigNotFound",
		"HTTPSSLTraceTaskSslAsc",
		"HTTPSSLTraceTaskSslSealMsg",
		"HTTPRequestTraceTaskRequestRejected",
		"HTTPRequestTraceTaskRequestCancelled",
		"HTTPDriverGlobalSettingsTaskHotAddProcFailed",
		"HTTPDriverGlobalSettingsTaskHotAddProcSucceeded",
		"HTTPRequestTraceTaskUserResponseFlowInit",
		"HTTPRequestTraceTaskCachedResponseFlowInit",
		"HTTPRequestTraceTaskFlowInitFailed",
		"HTTPConnectionTraceTaskSetConnectionFlow",
		"HTTPConnectionTraceTaskRequestAssociatedToConfigurationFlow",
		"HTTPConnectionTraceTaskConnectionFlowFailed",
		"HTTPRequestTraceTaskResponseRangeProcessingOK",
		"HTTPCacheTraceTaskBeginBuildingSlices",
		"HTTPCacheTraceTaskSendSliceCacheContent",
		"HTTPCacheTraceTaskCachedSlicesMatchContent",
		"HTTPCacheTraceTaskMergeSlicesToCache",
		"HTTPCacheTraceTaskFlatCacheRangeSend",
		"HTTPAuthenticationTraceTaskChannelBindAscParams",
		"HTTPAuthenticationTraceTaskServiceBindCheckComplete",
		"HTTPAuthenticationTraceTaskChannelBindConfigCapture",
		"HTTPAuthenticationTraceTaskChannelBindPerResponseConfig",
		"HTTPConnectionTraceTaskUsePolicyBasedQoSFlow",
		"HTTPThreadPoolThreadPoolExtension",
		"HTTPThreadPoolThreadReady",
		"HTTPThreadPoolThreadPoolTrim",
		"HTTPThreadPoolThreadGone",
		"HTTPSSLTraceTaskSniParsed",
		"HTTPRequestTraceTaskInitiateOpaqueMode",
		"HTTPSSLTraceTaskEndpointAutoGenerated",
		"HTTPSSLTraceTaskAutoGeneratedEndpointDeleted",
		"HTTPSSLTraceTaskSslEndpointConfigFound",
		"HTTPSSLTraceTaskSslEndpointConfigRejected",
		"HTTPResponseTraceTaskParseRequestFailed",
		"HTTPSSLTraceTaskSslHandshakeFailure",
		"HTTPRequestTraceTaskHttpErrorResponseSent",
		"HTTPSSLTraceTaskSslRenegotiateTimedOut",
		"HTTPRequestTraceTaskHttp11Required",
		"HTTPConnectionTraceTaskQuicConnection",
		"HTTPConnectionTraceTaskQuicConnectionCallback",
		"HTTPConnectionTraceTaskQuicStream",
		"HTTPConnectionTraceTaskQuicStreamCallback",
		"HTTPDriverGlobalSettingsTaskQuicRegistration",
	}
)

// FormatHTTPServiceEventID converts an ID to a string
func FormatHTTPServiceEventID(eventID uint16) string {
	if eventID == IDHTTPDummyStart || eventID >= IDHTTPDummyMax {
		return "<UNKNOWN_HTTP_EVENT_ID>"
	}

	return httpServiceEventID2Name[eventID]
}
