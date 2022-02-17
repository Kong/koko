package testdata

import "embed"

//go:embed lua-tree/share/lua/5.1/*
var LuaTree embed.FS
