package drive

import "os"

// ClientID and ClientSecret identify falafal's own registered "Desktop app"
// OAuth client in Google Cloud Console. For installed/desktop apps Google
// does not treat this secret as confidential — there is no way to keep a
// secret hidden inside a distributed binary — access is still gated by
// per-user consent and, while this app is in Testing publishing status, by
// a test-user allowlist Google enforces on their end.
//
// Values are never committed to source. Release binaries get them baked in
// at build time via -ldflags "-X .../drive.ClientID=... -X .../drive.ClientSecret=...",
// sourced from a CI secret store. FALAFAL_DRIVE_CLIENT_ID and
// FALAFAL_DRIVE_CLIENT_SECRET env vars override them for local development.
var (
	ClientID     = ""
	ClientSecret = ""
)

func clientID() string {
	if v := os.Getenv("FALAFAL_DRIVE_CLIENT_ID"); v != "" {
		return v
	}
	return ClientID
}

func clientSecret() string {
	if v := os.Getenv("FALAFAL_DRIVE_CLIENT_SECRET"); v != "" {
		return v
	}
	return ClientSecret
}
