package testdata

import "embed"

//go:embed schemas/*
var TestKongSchemas embed.FS
