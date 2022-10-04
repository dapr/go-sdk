package version

import (
	// Required for go:embed
	_ "embed"
)

// This weird redirection is courtesy of our linter 💄
//
//go:embed sdk-version
var sdkVersion string

// SDKVersion contains the version of the SDK.
var SDKVersion string = sdkVersion
