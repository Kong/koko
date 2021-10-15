package jsonschema

import "embed"

//go:embed schemas/*
var KongSchemas embed.FS
