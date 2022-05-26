package validators

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jeremywohl/flatten"
	goksPlugin "github.com/kong/goks/plugin"
	grpcModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

var ErrSchemaNotFound = fmt.Errorf("not found")

type Opts struct {
	Logger      *zap.Logger
	InjectFS    *embed.FS
	StoreLoader util.StoreLoader
}

type LuaValidator struct {
	goksV          *goksPlugin.Validator
	logger         *zap.Logger
	rawLuaSchemas  map[string][]byte
	luaSchemaNames []string
	storeLoader    util.StoreLoader

	// customPluginCheckMu ensures that only a single custom plugin schema is
	// loaded and checked at a time.
	customPluginCheckMu sync.Mutex
}

func NewLuaValidator(opts Opts) (*LuaValidator, error) {
	if opts.Logger == nil {
		return nil, fmt.Errorf("opts.Logger required")
	}
	validator, err := goksPlugin.NewValidator(goksPlugin.ValidatorOpts{
		InjectFS: opts.InjectFS,
	})
	if err != nil {
		return nil, err
	}
	return &LuaValidator{
		goksV:          validator,
		logger:         opts.Logger,
		rawLuaSchemas:  map[string][]byte{},
		luaSchemaNames: make([]string, 0),
		storeLoader:    opts.StoreLoader,
	}, nil
}

func (v *LuaValidator) SetStoreLoader(storeLoader util.StoreLoader) {
	v.storeLoader = storeLoader
}

// Validate implements the Validator.Validate interface.
func (v *LuaValidator) Validate(ctx context.Context, plugin *grpcModel.Plugin) error {
	start := time.Now()
	defer func() {
		v.logger.With(zap.String("plugin", plugin.Name),
			zap.Duration("validation-time", time.Since(start))).
			Debug("plugin validated via lua VM")
	}()
	unloadLuaPluginSchema := v.loadLuaPluginSchema(ctx, plugin.Name)
	defer unloadLuaPluginSchema()
	pluginJSON, err := json.ProtoJSONMarshal(plugin)
	if err != nil {
		return fmt.Errorf("marshal JSON: %v", err)
	}
	err = v.goksV.Validate(string(pluginJSON))
	if err != nil {
		return validationErr(plugin.Name, err)
	}
	return nil
}

func (v *LuaValidator) ValidateSchema(ctx context.Context, pluginSchema string) (string, error) {
	start := time.Now()
	pluginName, err := v.goksV.ValidateSchema(pluginSchema)
	defer func() {
		if len(pluginName) == 0 {
			pluginName = "no plugin name could be retrieved"
		}
		v.logger.With(zap.String("plugin-schema", pluginName),
			zap.Duration("validation-time", time.Since(start))).
			Debug("plugin schema validated via lua VM")
	}()
	if err != nil {
		return "", validationSchemaErr(grpcModel.ErrorType_ERROR_TYPE_FIELD, "lua_schema", err.Error())
	}
	for _, luaSchemaName := range v.luaSchemaNames {
		if pluginName == luaSchemaName {
			return "", validationSchemaErr(grpcModel.ErrorType_ERROR_TYPE_ENTITY, "",
				fmt.Sprintf("unique constraint failed: schema already exists for plugin '%s'", pluginName))
		}
	}
	return pluginName, nil
}

func validationErr(name string, e error) error {
	if e == nil {
		return nil
	}
	var errMap map[string]interface{}
	err := json.ProtoJSONUnmarshal([]byte(e.Error()), &errMap)
	if err != nil {
		return fmt.Errorf("unmarshal kong plugin validation error: %v", err)
	}
	res := validation.Error{}
	// name error happens when plugin doesn't exist
	if _, ok := errMap["name"]; ok {
		res.Errs = append(res.Errs, &grpcModel.ErrorDetail{
			Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
			Field: "name",
			Messages: []string{
				fmt.Sprintf("plugin(%v) does not exist", name),
			},
		})
		return res
	}

	// @entity errors
	if eErr, ok := errMap["@entity"]; ok {
		eErr := entityErr(eErr)
		if eErr != nil {
			res.Errs = append(res.Errs, eErr)
		}
		delete(errMap, "@entity")
	}

	// all remaining field errors
	errs, err := f(errMap)
	if err != nil {
		return err
	}
	res.Errs = append(res.Errs, errs...)

	// Sorting errors for predictability.
	validation.SortErrorDetails(res.Errs)

	return res
}

func entityErr(err interface{}) *grpcModel.ErrorDetail {
	errs, ok := err.([]interface{})
	if !ok {
		panic(fmt.Sprintf("expected '@entity' key to be []interface{} but got"+
			" %T", err))
	}
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		message, ok := err.(string)
		if !ok {
			panic(fmt.Sprintf("expected '@entity' element to be string but got %T", err))
		}
		messages = append(messages, message)
	}
	if len(messages) > 0 {
		return &grpcModel.ErrorDetail{
			Type:     grpcModel.ErrorType_ERROR_TYPE_ENTITY,
			Messages: messages,
		}
	}
	return nil
}

func validationSchemaErr(errType grpcModel.ErrorType, field string, message string) error {
	err := &grpcModel.ErrorDetail{
		Type:     errType,
		Messages: []string{message},
	}
	if len(field) > 0 {
		err.Field = field
	}
	return validation.Error{
		Errs: []*grpcModel.ErrorDetail{err},
	}
}

var flattenStyle = flatten.SeparatorStyle{
	Middle:                   ".",
	UseBracketsForArrayIndex: true,
}

func f(m map[string]interface{}) ([]*grpcModel.ErrorDetail, error) {
	m, err := flatten.Flatten(m, "", flattenStyle)
	if err != nil {
		return nil, err
	}
	var res []*grpcModel.ErrorDetail
	for k, v := range m {
		switch typedV := v.(type) {
		case string:
			res = append(res, &grpcModel.ErrorDetail{
				Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
				Field:    k,
				Messages: []string{typedV},
			})

		default:
			panic(fmt.Sprintf("unexpected value type for a key(%v) in plugin"+
				" configuration error: %T", k, v))
		}
	}
	return res, nil
}

// ProcessDefaults implements the Validator.ProcessDefaults interface.
func (v *LuaValidator) ProcessDefaults(ctx context.Context, plugin *grpcModel.Plugin) error {
	start := time.Now()
	defer func() {
		v.logger.With(zap.String("plugin", plugin.Name),
			zap.Duration("process-defaults-time", time.Since(start))).
			Debug("plugin defaults processed via lua VM")
	}()
	unloadLuaPluginSchema := v.loadLuaPluginSchema(ctx, plugin.Name)
	defer unloadLuaPluginSchema()
	pluginJSON, err := json.ProtoJSONMarshal(plugin)
	if err != nil {
		return fmt.Errorf("marshal JSON: %v", err)
	}
	updatedPluginJSON, err := v.goksV.ProcessAutoFields(string(pluginJSON))
	if err != nil {
		return fmt.Errorf("process auto fields failed: %v", err)
	}
	err = json.MarshallerWithDiscard.Unmarshal([]byte(updatedPluginJSON), plugin)
	if err != nil {
		return fmt.Errorf("unmarshal JSON: %v", err)
	}
	return nil
}

func (v *LuaValidator) LoadSchemasFromEmbed(fs embed.FS, dirName string) error {
	dirEntries, err := fs.ReadDir(dirName)
	if err != nil {
		return err
	}

	t1 := time.Now()
	for _, entry := range dirEntries {
		name := entry.Name()
		v.logger.With(zap.String("name", name)).Debug("reading/loading plugin schema")
		schema, err := fs.ReadFile(dirName + "/" + name)
		if err != nil {
			return err
		}
		pluginName, err := v.goksV.LoadSchema(string(schema))
		if err != nil {
			return err
		}

		// Get the JSON schema for the plugin that was loaded and store it in mem
		pluginSchema, err := v.goksV.SchemaAsJSON(pluginName)
		if err != nil {
			return err
		}
		err = addLuaSchema(pluginName, pluginSchema, v.rawLuaSchemas, &v.luaSchemaNames)
		if err != nil {
			return err
		}
	}

	// Sorting available plugins for predictability.
	sort.Strings(v.luaSchemaNames)

	t2 := time.Now()
	v.logger.
		With(zap.Duration("loading-time", t2.Sub(t1))).
		Debug("plugin schemas loaded")
	return nil
}

// GetRawLuaSchema implements the Validator.GetRawLuaSchema interface.
func (v *LuaValidator) GetRawLuaSchema(ctx context.Context, name string) ([]byte, error) {
	rawLuaSchema, ok := v.rawLuaSchemas[name]
	if !ok {
		return []byte{}, ErrSchemaNotFound
	}
	return rawLuaSchema, nil
}

// GetRawLuaSchemaForCustomPlugin implements the Validator.GetRawLuaSchemaForCustomPlugin interface.
func (v *LuaValidator) GetRawLuaSchemaForCustomPlugin(ctx context.Context, name string) ([]byte, error) {
	start := time.Now()
	defer func() {
		v.logger.With(zap.String("plugin", name),
			zap.Duration("get-custom-plugin-schema-time", time.Since(start))).
			Debug("custom plugin schema retrieved via lua VM")
	}()
	unloadLuaPluginSchema := v.loadLuaPluginSchema(ctx, name)
	defer unloadLuaPluginSchema()
	pluginSchema, err := v.goksV.SchemaAsJSON(name)
	if err != nil {
		return []byte{}, ErrSchemaNotFound
	}
	return []byte(pluginSchema), nil
}

// GetAvailablePluginNames implements the Validator.GetAvailablePluginNames interface.
func (v *LuaValidator) GetAvailablePluginNames(ctx context.Context) []string {
	return v.luaSchemaNames
}

func addLuaSchema(name string, schema string, rawLuaSchemas map[string][]byte, luaSchemaNames *[]string) error {
	if _, found := rawLuaSchemas[name]; found {
		return fmt.Errorf("schema for plugin '%s' already exists", name)
	}
	trimmedSchema := strings.TrimSpace(schema)
	if len(trimmedSchema) == 0 {
		return errors.New("schema cannot be empty")
	}
	rawLuaSchemas[name] = []byte(schema)
	*luaSchemaNames = append(*luaSchemaNames, name)
	return nil
}

func (v *LuaValidator) getDB(ctx context.Context) (store.Store, error) {
	// To ensure current tests can remain executing with a single validator a special case is being
	// made to allow for a nil StoreLoader. This also ensures a panic will never occur when attempting
	// to use a nil StoreLoader since it not being forced during instantiation on the LuaValidator.
	if v.storeLoader == nil {
		return nil, errors.New("invalid StoreLoader: store loader cannot be nil")
	}

	if ctx == nil {
		return nil, errors.New("invalid context: cannot be nil")
	}
	reqCluster, ok := ctx.Value(util.ContextKeyCluster).(*grpcModel.RequestCluster)
	if !ok {
		return nil, errors.New("invalid context: failed to retrieve RequestCluster from context")
	}

	store, err := v.storeLoader.Load(ctx, reqCluster)
	if err != nil {
		if storeLoadErr, ok := err.(util.StoreLoadErr); ok {
			// TODO(hbagdi): status errors shouldn't be deep in here,
			// this should be higher up in the stack
			return nil, status.Error(storeLoadErr.Code, storeLoadErr.Message)
		}
		return nil, err
	}
	return store, nil
}

func (v *LuaValidator) getPluginSchemaFromDB(ctx context.Context, pluginName string) string {
	db, err := v.getDB(ctx)
	if err != nil {
		v.logger.With(zap.Any("context", ctx)).
			With(zap.String("name", pluginName)).
			With(zap.Error(err)).Warn("retrieving database")
		return ""
	}
	pluginName = strings.TrimSpace(pluginName)
	if len(pluginName) == 0 {
		v.logger.Debug("retrieving plugin schema failed: empty plugin name")
		return ""
	}

	ps := resource.NewPluginSchema()
	if err = db.Read(ctx, ps, store.GetByID(pluginName)); err != nil {
		v.logger.With(zap.String("name", pluginName)).With(zap.Error(err)).Warn("retrieving plugin schema")
	}
	return ps.PluginSchema.LuaSchema
}

// loadLuaPluginSchema is used for non-bundled plugins. It will load a plugin schema from the StoreLoader
// based on the plugin name and provide a mechanism to unload the schema from memory when goks operations
// are complete. This method will never return an error in order to allow validation to pass to the
// operations that are requesting a schema to load (e.g. Validate() and ProcessDefaults()).
func (v *LuaValidator) loadLuaPluginSchema(ctx context.Context, pluginName string) func() {
	if _, found := v.rawLuaSchemas[pluginName]; found {
		// plugin schema is bundled and already loaded into memory
		return func() {}
	}

	schema := v.getPluginSchemaFromDB(ctx, pluginName)
	if len(schema) == 0 {
		// ensure the schema isn't loaded to allow validate error
		return func() {}
	}

	v.customPluginCheckMu.Lock()
	unloadSchema := func() {
		defer v.customPluginCheckMu.Unlock()
	}
	if _, err := v.goksV.LoadSchema(schema); err != nil {
		// load schema errors shouldn't trigger failure to allow validate error
		return unloadSchema
	}
	unloadSchema = func() {
		defer v.customPluginCheckMu.Unlock()
		if err := v.goksV.UnloadSchema(pluginName); err != nil {
			v.logger.With(zap.String("name", pluginName)).With(zap.Error(err)).Error("unloading plugin schema")
		} else {
			v.logger.With(zap.String("name", pluginName)).Debug("unloading plugin schema")
		}
	}

	// Return the cleanup function for properly unloading the plugin schema
	return unloadSchema
}
