package plugin

import (
	"embed"
	"fmt"
	"time"

	"github.com/jeremywohl/flatten"
	goksPlugin "github.com/kong/goks/plugin"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model/json/validation"
	"go.uber.org/zap"
)

type Validator interface {
	Validate(*model.Plugin) error
	ProcessDefaults(*model.Plugin) error
}

type Opts struct {
	Logger *zap.Logger
}

type LuaValidator struct {
	goksV  *goksPlugin.Validator
	logger *zap.Logger
}

func NewLuaValidator(opts Opts) (*LuaValidator, error) {
	if opts.Logger == nil {
		return nil, fmt.Errorf("opts.Logger required")
	}
	validator, err := goksPlugin.NewValidator()
	if err != nil {
		return nil, err
	}
	return &LuaValidator{
		goksV:  validator,
		logger: opts.Logger,
	}, nil
}

func (v *LuaValidator) Validate(plugin *model.Plugin) error {
	start := time.Now()
	defer func() {
		v.logger.With(zap.String("plugin", plugin.Name),
			zap.Duration("validation-time", time.Since(start))).
			Debug("plugin validated via lua VM")
	}()
	pluginJSON, err := json.Marshal(plugin)
	if err != nil {
		return fmt.Errorf("marshal JSON: %v", err)
	}
	err = v.goksV.Validate(string(pluginJSON))
	if err != nil {
		return validationErr(plugin.Name, err)
	}
	return nil
}

func validationErr(name string, e error) error {
	if e == nil {
		return nil
	}
	var errMap map[string]interface{}
	err := json.Unmarshal([]byte(e.Error()), &errMap)
	if err != nil {
		return fmt.Errorf("unmarshal kong plugin validation error: %v", err)
	}
	res := validation.Error{}
	// name error happens when plugin doesn't exist
	if _, ok := errMap["name"]; ok {
		res.Errs = append(res.Errs, &model.ErrorDetail{
			Type:  model.ErrorType_ERROR_TYPE_FIELD,
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

	return res
}

func entityErr(err interface{}) *model.ErrorDetail {
	errs, ok := err.([]interface{})
	if !ok {
		panic(fmt.Sprintf("expected '@entity' key to be []interface{} but got"+
			" %T", err))
	}
	var messages []string
	for _, err := range errs {
		message, ok := err.(string)
		if !ok {
			panic(fmt.Sprintf("expected '@entity' element to be string but got %T", err))
		}
		messages = append(messages, message)
	}
	if len(messages) > 0 {
		return &model.ErrorDetail{
			Type:     model.ErrorType_ERROR_TYPE_ENTITY,
			Messages: messages,
		}
	}
	return nil
}

var flattenStyle = flatten.SeparatorStyle{
	Middle:                   ".",
	UseBracketsForArrayIndex: true,
}

func f(m map[string]interface{}) ([]*model.ErrorDetail, error) {
	m, err := flatten.Flatten(m, "", flattenStyle)
	if err != nil {
		return nil, err
	}
	var res []*model.ErrorDetail
	for k, v := range m {
		switch typedV := v.(type) {
		case string:
			res = append(res, &model.ErrorDetail{
				Type:     model.ErrorType_ERROR_TYPE_FIELD,
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

func (v *LuaValidator) ProcessDefaults(plugin *model.Plugin) error {
	pluginJSON, err := json.Marshal(plugin)
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
		err = AddLuaSchema(pluginName, pluginSchema)
		if err != nil {
			return err
		}
	}
	t2 := time.Now()
	v.logger.
		With(zap.Duration("loading-time", t2.Sub(t1))).
		Debug("plugin schemas loaded")
	return nil
}
