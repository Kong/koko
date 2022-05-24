package admin

import (
	"net/http/httptest"
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/plugin"
	"github.com/kong/koko/internal/plugin/validators"
	"github.com/kong/koko/internal/resource"
	serverUtil "github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
)

var validator plugin.Validator

func init() {
	luaValidator, err := validators.NewLuaValidator(validators.Opts{Logger: log.Logger})
	if err != nil {
		panic(err)
	}
	err = luaValidator.LoadSchemasFromEmbed(plugin.Schemas, "schemas")
	if err != nil {
		panic(err)
	}
	validator = luaValidator
	resource.SetValidator(validator)
}

func setup(t *testing.T) (*httptest.Server, func()) {
	p, err := util.GetPersister(t)
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)

	server, cleanup := setupWithDB(t, objectStore.ForCluster("default"))
	return server, func() {
		cleanup()
	}
}

func setupWithDB(t *testing.T, store store.Store) (*httptest.Server, func()) {
	handler, err := NewHandler(HandlerOpts{
		Logger: log.Logger,
		StoreLoader: serverUtil.DefaultStoreLoader{
			Store: store,
		},
		Validator: validator,
	})
	if err != nil {
		t.Fatalf("creating httptest.Server: %v", err)
	}

	h := serverUtil.HandlerWithRecovery(serverUtil.HandlerWithLogger(handler, log.Logger), log.Logger)
	s := httptest.NewServer(h)
	return s, func() {
		s.Close()
	}
}
