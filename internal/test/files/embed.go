package files

import "embed"

//go:embed schemas/*
var TestKongSchemas embed.FS
