package badtestdata

import (
	"embed"
)

//go:embed lua-tree/share/lua/5.1/*
var BadLuaTree embed.FS
