package admin

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/stretchr/testify/require"
)

func TestVaultCreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("creates a valid vault", func(t *testing.T) {
		vault := &v1.Vault{
			Name:   "env",
			Prefix: "test-vault-1",
			Config: &v1.Vault_Config{
				Config: &v1.Vault_Config_Env{
					Env: &v1.Vault_EnvConfig{
						Prefix: "TEST_",
					},
				},
			},
		}
		vaultBytes, err := json.ProtoJSONMarshal(vault)
		require.NoError(t, err)
		res := c.POST("/v1/vaults").WithBytes(vaultBytes).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("name").String().Equal(vault.Name)
		body.Value("prefix").String().Equal(vault.Prefix)
		body.Path("$.config.env.prefix").Equal(vault.Config.GetEnv().Prefix)
	})
	t.Run("creating an invalid vault fails", func(t *testing.T) {
		vault := &v1.Vault{
			Name:   "unsupported-vault",
			Prefix: "SECRET_",
			Config: &v1.Vault_Config{
				Config: &v1.Vault_Config_Env{
					Env: &v1.Vault_EnvConfig{
						Prefix: "TEST-",
					},
				},
			},
		}
		vaultBytes, err := json.ProtoJSONMarshal(vault)
		require.NoError(t, err)
		res := c.POST("/v1/vaults").WithBytes(vaultBytes).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(2)
		errs := body.Value("details").Array()
		var fields []string
		for _, err := range errs.Iter() {
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.
				String())
			fields = append(fields, err.Object().Value("field").String().Raw())
		}
		require.ElementsMatch(t, []string{"name", "prefix"}, fields)
	})
}

func TestVaultUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("upserts a valid vault", func(t *testing.T) {
		id := uuid.NewString()
		vault := &v1.Vault{
			Name:   "env",
			Prefix: "test-vault-1",
			Config: &v1.Vault_Config{
				Config: &v1.Vault_Config_Env{
					Env: &v1.Vault_EnvConfig{
						Prefix: "TEST_",
					},
				},
			},
		}
		vaultBytes, err := json.ProtoJSONMarshal(vault)
		require.NoError(t, err)
		res := c.PUT("/v1/vaults/{id}", id).WithBytes(vaultBytes).Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").String().Equal(id)
		body.Value("name").String().Equal(vault.Name)
		body.Value("prefix").String().Equal(vault.Prefix)
		body.Path("$.config.env.prefix").Equal(vault.Config.GetEnv().Prefix)
		vault.Id = body.Value("id").String().Raw()
	})
	t.Run("upsert an existing vault succeeds", func(t *testing.T) {
		vault := &v1.Vault{
			Name:   "env",
			Prefix: "test-vault-2",
			Config: &v1.Vault_Config{
				Config: &v1.Vault_Config_Env{
					Env: &v1.Vault_EnvConfig{
						Prefix: "TEST_",
					},
				},
			},
		}
		vaultBytes, err := json.ProtoJSONMarshal(vault)
		require.NoError(t, err)
		res := c.POST("/v1/vaults").WithBytes(vaultBytes).Expect()
		res.Status(http.StatusCreated)
		body := res.JSON().Path("$.item").Object()
		body.Value("name").String().Equal(vault.Name)
		body.Value("prefix").String().Equal(vault.Prefix)
		body.Path("$.config.env.prefix").Equal(vault.Config.GetEnv().Prefix)
		id := body.Value("id").String().Raw()

		vaultUpdate := &v1.Vault{
			Id:     id,
			Name:   "env",
			Prefix: "test-vault-3",
			Config: &v1.Vault_Config{
				Config: &v1.Vault_Config_Env{
					Env: &v1.Vault_EnvConfig{
						Prefix: "TEST2_",
					},
				},
			},
		}
		vaultUpdateBytes, err := json.ProtoJSONMarshal(vaultUpdate)
		require.NoError(t, err)
		res = c.PUT("/v1/vaults/{id}", id).WithBytes(vaultUpdateBytes).Expect()
		res.Status(http.StatusOK)
		body = res.JSON().Path("$.item").Object()
		body.Value("id").String().Equal(id)
		body.Value("name").String().Equal(vaultUpdate.Name)
		body.Value("prefix").String().Equal(vaultUpdate.Prefix)
		body.Path("$.config.env.prefix").Equal(vaultUpdate.Config.GetEnv().Prefix)
	})
	t.Run("upsert vault without id fails", func(t *testing.T) {
		res := c.PUT("/v1/vaults/").
			WithJSON(&v1.Vault{}).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
	t.Run("upsert an invalid vault fails", func(t *testing.T) {
		vault := &v1.Vault{
			Name:   "unsupported",
			Prefix: "test-vault-2",
			Config: &v1.Vault_Config{
				Config: &v1.Vault_Config_Env{
					Env: &v1.Vault_EnvConfig{
						Prefix: "TEST_",
					},
				},
			},
		}

		vaultBytes, err := json.ProtoJSONMarshal(vault)
		require.NoError(t, err)
		res := c.PUT("/v1/vaults/{id}", uuid.NewString()).WithBytes(vaultBytes).
			Expect().Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errs := body.Value("details").Array()
		var fields []string
		for _, err := range errs.Iter() {
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.
				String())
			fields = append(fields, err.Object().Value("field").String().Raw())
		}
		require.ElementsMatch(t, []string{"name"}, fields)
	})
}

func TestVaultRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	vault := &v1.Vault{
		Name:   "env",
		Prefix: "test-vault-5",
		Config: &v1.Vault_Config{
			Config: &v1.Vault_Config_Env{
				Env: &v1.Vault_EnvConfig{
					Prefix: "TEST_",
				},
			},
		},
	}
	vaultBytes, err := json.ProtoJSONMarshal(vault)
	require.NoError(t, err)
	res := c.POST("/v1/vaults").WithBytes(vaultBytes).Expect().Status(http.StatusCreated)
	id := res.JSON().Path("$.item.id").String().Raw()
	t.Run("read with an empty id returns 400", func(t *testing.T) {
		res := c.GET("/v1/vaults/").Expect().Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "required ID is missing")
	})
	t.Run("reading a non-existent vault returns 404", func(t *testing.T) {
		c.GET("/v1/vaults/{id}", uuid.NewString()).
			Expect().Status(http.StatusNotFound)
	})
	t.Run("reading with an existing vault id returns 200", func(t *testing.T) {
		res := c.GET("/v1/vaults/{id}", id).
			Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").Equal(id)
		body.Value("name").Equal(vault.Name)
		body.Value("prefix").Equal(vault.Prefix)
		body.Path("$.config.env.prefix").Equal(vault.Config.GetEnv().Prefix)
	})
	t.Run("reading with an existing vault prefix returns 200", func(t *testing.T) {
		res := c.GET("/v1/vaults/{prefix}", vault.Prefix).
			Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").Equal(id)
		body.Value("name").Equal(vault.Name)
		body.Value("prefix").Equal(vault.Prefix)
		body.Path("$.config.env.prefix").Equal(vault.Config.GetEnv().Prefix)
	})
}

func TestVaultDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	vault := &v1.Vault{
		Name:   "env",
		Prefix: "test-vault-1",
		Config: &v1.Vault_Config{
			Config: &v1.Vault_Config_Env{
				Env: &v1.Vault_EnvConfig{
					Prefix: "TEST_",
				},
			},
		},
	}
	vaultBytes, err := json.ProtoJSONMarshal(vault)
	require.NoError(t, err)
	res := c.POST("/v1/vaults").WithBytes(vaultBytes).Expect().Status(http.StatusCreated)
	id := res.JSON().Path("$.item.id").String().Raw()
	t.Run("deleting a non-existent vault returns 404", func(t *testing.T) {
		dres := c.DELETE("/v1/vaults/{id}", uuid.NewString()).Expect()
		dres.Status(http.StatusNotFound)
	})
	t.Run("delete with an invalid id returns 400", func(t *testing.T) {
		dres := c.DELETE("/v1/vaults/").Expect()
		dres.Status(http.StatusBadRequest)
	})
	t.Run("delete an existing vault succeeds", func(t *testing.T) {
		dres := c.DELETE("/v1/vaults/{id}", id).Expect()
		dres.Status(http.StatusNoContent)
	})
}

func TestVaultList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	ids := make([]string, 0, 4)
	vaults := []*v1.Vault{
		{
			Name:   "env",
			Prefix: "my-test-vault-1",
			Config: &v1.Vault_Config{
				Config: &v1.Vault_Config_Env{
					Env: &v1.Vault_EnvConfig{
						Prefix: "TEST1_",
					},
				},
			},
		},
		{
			Name:   "env",
			Prefix: "my-test-vault-2",
			Config: &v1.Vault_Config{
				Config: &v1.Vault_Config_Env{
					Env: &v1.Vault_EnvConfig{
						Prefix: "TEST2_",
					},
				},
			},
		},
		{
			Name:   "env",
			Prefix: "my-test-vault-3",
			Config: &v1.Vault_Config{
				Config: &v1.Vault_Config_Env{
					Env: &v1.Vault_EnvConfig{
						Prefix: "TEST3_",
					},
				},
			},
		},
		{
			Name:   "env",
			Prefix: "my-test-vault-4",
			Config: &v1.Vault_Config{
				Config: &v1.Vault_Config_Env{
					Env: &v1.Vault_EnvConfig{
						Prefix: "TEST4_",
					},
				},
			},
		},
	}
	for _, vault := range vaults {
		vaultBytes, err := json.ProtoJSONMarshal(vault)
		require.NoError(t, err)
		res := c.POST("/v1/vaults").WithBytes(vaultBytes).Expect().Status(http.StatusCreated)
		ids = append(ids, res.JSON().Path("$.item.id").String().Raw())
	}

	t.Run("list returns multiple vaults", func(t *testing.T) {
		body := c.GET("/v1/vaults").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(4)
		gotIDs := make([]string, 0, 4)
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, ids, gotIDs)
	})
	t.Run("list returns multiple vaults with paging", func(t *testing.T) {
		body := c.GET("/v1/vaults").
			WithQuery("page.size", "2").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		gotIDs := make([]string, 0, 4)
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		body.Value("page").Object().Value("total_count").Number().Equal(4)
		next := body.Value("page").Object().Value("next_page_num").Number().Equal(2).Raw()
		body = c.GET("/v1/vaults").
			WithQuery("page.size", "2").
			WithQuery("page.number", next).
			Expect().Status(http.StatusOK).JSON().Object()
		body.Value("page").Object().Value("total_count").Number().Equal(4)
		body.Value("page").Object().NotContainsKey("next_page")
		items = body.Value("items").Array()
		items.Length().Equal(2)
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, ids, gotIDs)
	})
}
