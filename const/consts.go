package consts

import "github.com/zoenion/common/types"

const (
	KeyAuth              = types.String("X-Oe-Auth")
	KeyAppHeaders        = types.String("X-Oe-Headers")
	KeyRequestSource     = types.String("X-Oe-Request-Source")
	KeyLocalAddress      = types.String("X-Oe-Local-Address")
	KeyRemoteAddress     = types.String("X-Oe-Remote-Address")
	KeyRemoteCertificate = types.String("X-Oe-Remote1-Certificate")
	KeyService           = types.String("X-Oe-Service")
	KeyServiceName       = types.String("X-Oe-Service-Name")
	KeyInTime            = types.String("X-Oe-In-Time")
	KeyLang              = types.String("X-Oe-Lang")
)
