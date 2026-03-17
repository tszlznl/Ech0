package model

const (
	MsgKeyCommonSuccess       = "common.success"
	MsgKeyCommonRequestFailed = "common.request_failed"
	MsgKeyInvalidQueryParams  = "common.invalid_query_params"
	MsgKeySettingUpdateOK     = "setting.update_success"
	MsgKeyAgentModelMissing   = "agent.model_missing"
	MsgKeyAuthTokenMissing    = "auth.token_missing"
	MsgKeyAuthTokenInvalid    = "auth.token_invalid"
	MsgKeyAuthTokenParse      = "auth.token_parse_error"
	MsgKeyDashboardLogsOk     = "dashboard.logs.success"
	MsgKeyDashboardTailBad    = "dashboard.logs.tail_invalid"
	MsgKeyInboxNewVersion     = "inbox.new_version_available"
)

func MessageKeyFromErrorCode(code string) string {
	switch code {
	case ErrCodeInvalidQuery:
		return MsgKeyInvalidQueryParams
	case ErrCodeTokenMissing:
		return MsgKeyAuthTokenMissing
	case ErrCodeTokenInvalid:
		return MsgKeyAuthTokenInvalid
	case ErrCodeTokenParse:
		return MsgKeyAuthTokenParse
	default:
		return ""
	}
}

func MessageKeyFromMessage(msg string) string {
	switch msg {
	case SUCCESS_MESSAGE:
		return MsgKeyCommonSuccess
	case UPDATE_SETTINGS_SUCCESS:
		return MsgKeySettingUpdateOK
	case AGENT_MODEL_MISSING:
		return MsgKeyAgentModelMissing
	default:
		return ""
	}
}
