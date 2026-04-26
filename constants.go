package doauth

// Well-known paths
const (
	PathOpenIDConfig        = "/.well-known/openid-configuration"
	PathOAuthAuthServer     = "/.well-known/oauth-authorization-server"
	PathOAuthProtectedRoute = "/.well-known/oauth-protected-resource"
)

// Header names
const (
	HeaderAuthorization   = "Authorization"
	HeaderWWWAuthenticate = "WWW-Authenticate"
	HeaderXDiscoveryURL   = "X-Discovery-URL"
)

// OAuth2 parameter names
const (
	ParamCode                = "code"
	ParamState               = "state"
	ParamError               = "error"
	ParamResource            = "resource"
	ParamCodeChallenge       = "code_challenge"
	ParamCodeChallengeMethod = "code_challenge_method"
	ParamCodeVerifier        = "code_verifier"
)

// PKCE methods
const (
	MethodS256 = "S256"
)

// Browser commands
const (
	CmdWindows  = "rundll32"
	ArgsWindows = "url.dll,FileProtocolHandler"
	CmdDarwin   = "open"
	CmdLinux    = "xdg-open"
)

// Default values
const (
	DefaultPort         = 8080
	DefaultCallbackPath = "/callback"
	BearerPrefix        = "Bearer "
)
