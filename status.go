package rtsp

// RTSP response status codes
const (
	StatusContinue                      = 100
	StatusOK                            = 200
	StatusCreated                       = 201
	StatusLowOnStorageSpace             = 250
	StatusMultipleChoices               = 300
	StatusMovedPermanently              = 301
	StatusMovedTemporarily              = 302
	StatusSeeOther                      = 303
	StatusNotModified                   = 304
	StatusUseProxy                      = 305
	StatusBadRequest                    = 400
	StatusUnauthorized                  = 401
	StatusPaymentRequired               = 402
	StatusForbidden                     = 403
	StatusNotFound                      = 404
	StatusMethodNotAllowed              = 405
	StatusNotAcceptable                 = 406
	StatusProxyAuthenticationRequired   = 407
	StatusRequestTimeout                = 408
	StatusGone                          = 410
	StatusLengthRequired                = 411
	StatusPreconditionFailed            = 412
	StatusRequestEntityTooLarge         = 413
	StatusRequestURITooLong             = 414
	StatusUnsupportedMediaType          = 415
	StatusInvalidparameter              = 451
	StatusIllegalConferenceIdentifier   = 452
	StatusNotEnoughBandwidth            = 453
	StatusSessionNotFound               = 454
	StatusMethodNotValidInThisState     = 455
	StatusHeaderFieldNotValid           = 456
	StatusInvalidRange                  = 457
	StatusParameterIsReadOnly           = 458
	StatusAggregateOperationNotAllowed  = 459
	StatusOnlyAggregateOperationAllowed = 460
	StatusUnsupportedTransport          = 461
	StatusDestinationUnreachable        = 462
	StatusInternalServerError           = 500
	StatusNotImplemented                = 501
	StatusBadGateway                    = 502
	StatusServiceUnavailable            = 503
	StatusGatewayTimeout                = 504
	StatusRTSPVersionNotSupported       = 505
	StatusOptionNotsupport              = 551
)

// StatusText returns a text of the RTSP status code. It returns the
// empty string if the code is unknown
func StatusText(code int) string {
	return statusText[code]
}

var statusText = map[int]string{
	StatusContinue:                      "Continue",
	StatusOK:                            "OK",
	StatusCreated:                       "Created",
	StatusLowOnStorageSpace:             "Low on Storage Space",
	StatusMultipleChoices:               "Multiple Choices",
	StatusMovedPermanently:              "Moved Permanently",
	StatusMovedTemporarily:              "Moved Temporarily",
	StatusSeeOther:                      "See Other",
	StatusNotModified:                   "Not Modified",
	StatusUseProxy:                      "Use Proxy",
	StatusBadRequest:                    "Bad Request",
	StatusUnauthorized:                  "Unauthorized",
	StatusPaymentRequired:               "Payment Required",
	StatusForbidden:                     "Forbidden",
	StatusNotFound:                      "Not Found",
	StatusMethodNotAllowed:              "Method Not Allowed",
	StatusNotAcceptable:                 "Not Acceptable",
	StatusProxyAuthenticationRequired:   "Proxy Authentication Required",
	StatusRequestTimeout:                "Request Time-out",
	StatusGone:                          "Gone",
	StatusLengthRequired:                "Length Required",
	StatusPreconditionFailed:            "Precondition Failed",
	StatusRequestEntityTooLarge:         "Request Entity Too Large",
	StatusRequestURITooLong:             "Request-URI Too Large",
	StatusUnsupportedMediaType:          "Unsupported Media Type",
	StatusInvalidparameter:              "Parameter Not Understood",
	StatusIllegalConferenceIdentifier:   "Conference Not Found",
	StatusNotEnoughBandwidth:            "Not Enough Bandwidth",
	StatusSessionNotFound:               "Session Not Found",
	StatusMethodNotValidInThisState:     "Method Not Valid in This State",
	StatusHeaderFieldNotValid:           "Header Field Not Valid for Resource",
	StatusInvalidRange:                  "Invalid Range",
	StatusParameterIsReadOnly:           "Parameter Is Read-Only",
	StatusAggregateOperationNotAllowed:  "Aggregate operation not allowed",
	StatusOnlyAggregateOperationAllowed: "Only aggregate operation allowed",
	StatusUnsupportedTransport:          "Unsupported transport",
	StatusDestinationUnreachable:        "Destination unreachable",
	StatusInternalServerError:           "Internal Server Error",
	StatusNotImplemented:                "Not Implemented",
	StatusBadGateway:                    "Bad Gateway",
	StatusServiceUnavailable:            "Service Unavailable",
	StatusGatewayTimeout:                "Gateway Time-out",
	StatusRTSPVersionNotSupported:       "RTSP Version not supported",
	StatusOptionNotsupport:              "Option not supported",
}
