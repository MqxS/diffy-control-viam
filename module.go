package diffy3

import (
	"context"
	"errors"
	"fmt"

	generic "go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/components/motor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var (
	DiffyModel       = resource.NewModel("namespace", "diffy3", "diffy_model")
	errUnimplemented = errors.New("unimplemented")
)

func init() {
	resource.RegisterComponent(generic.API, DiffyModel,
		resource.Registration[resource.Resource, *Config]{
			Constructor: newDiffy3DiffyModel,
		},
	)
}

type Config struct {
	LeftMotor  string `json:"left_motor"`
	RightMotor string `json:"right_motor"`
	/*
		Put config attributes here. There should be public/exported fields
		with a `json` parameter at the end of each attribute.

		Example config struct:
			type Config struct {
				Pin   string `json:"pin"`
				Board string `json:"board"`
				MinDeg *float64 `json:"min_angle_deg,omitempty"`
			}

		If your model does not need a config, replace *Config in the init
		function with resource.NoNativeConfig
	*/
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns three values:
//  1. Required dependencies: other resources that must exist for this resource to work.
//  2. Optional dependencies: other resources that may exist but are not required.
//  3. An error if any Config fields are missing or invalid.
//
// The `path` parameter indicates
// where this resource appears in the machine's JSON configuration
// (for example, "components.0"). You can use it in error messages
// to indicate which resource has a problem.
func (cfg *Config) Validate(path string) ([]string, []string, error) {
	requiredDeps := []string{}
	optionalDeps := []string{}

	if cfg.LeftMotor == "" {
		return nil, nil, fmt.Errorf("%s: missing required field LeftMotor", path)
	}
	requiredDeps = append(requiredDeps, cfg.LeftMotor)

	if cfg.RightMotor == "" {
		return nil, nil, fmt.Errorf("%s: missing required field RightMotor", path)
	}
	requiredDeps = append(requiredDeps, cfg.RightMotor)

	if cfg.LeftMotor == cfg.RightMotor {
		return nil, nil, fmt.Errorf("%s: LeftMotor and RightMotor cannot be the same motor", path)
	}

	return requiredDeps, optionalDeps, nil
}

type diffy3DiffyModel struct {
	resource.AlwaysRebuild

	name resource.Name

	logger logging.Logger
	cfg    *Config

	leftMotor  motor.Motor
	rightMotor motor.Motor

	cancelCtx  context.Context
	cancelFunc func()
}

func newDiffy3DiffyModel(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (resource.Resource, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewDiffyModel(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewDiffyModel(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (resource.Resource, error) {

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	s := &diffy3DiffyModel{
		name:       name,
		logger:     logger,
		cfg:        conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}
	return s, nil
}

func (s *diffy3DiffyModel) Name() resource.Name {
	return s.name
}

func (s *diffy3DiffyModel) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	err := s.leftMotor.SetPower(context.Background(), 1, cmd)
	if err != nil {
		return nil, fmt.Errorf("error in DoCommand: %w", err)
	}

	return map[string]interface{}{"result": "success"}, nil
}

func (s *diffy3DiffyModel) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}
