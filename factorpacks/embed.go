package factorpacks

import "embed"

// FS contains normalized factor-pack data files used by the default loader.
//
//go:embed defra-2025/*.json
var FS embed.FS
