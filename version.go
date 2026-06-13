package novada

// Version is the semantic version of this SDK. It is sent as part of the
// default User-Agent header on every request.
const Version = "0.1.0"

// userAgent is the default User-Agent header value used when the caller does
// not override it via WithUserAgent.
const userAgent = "novada-go/" + Version
