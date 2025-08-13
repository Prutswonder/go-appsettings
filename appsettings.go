package appsettings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/vrischmann/envconfig"
)

type Validator interface {
	Validate(settings any) error
}

// ReadSettingsFromFileAndEnv reads the settings from a local file and overrides them with
// environment variables.
func ReadSettingsFromFileAndEnv(settings any, validator Validator) error {
	// Step 1: Read settings from file
	if file, err := os.Open("appsettings.json"); err != nil {
		return errors.Join(fmt.Errorf("failed to open appsettings file"), err)
	} else {
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				err = errors.Join(fmt.Errorf("failed to close appsettings file"), closeErr)
			}
		}()
		data, err := io.ReadAll(file)

		if err != nil {
			return errors.Join(fmt.Errorf("failed to read appsettings file"), err)
		}

		if err = json.Unmarshal(data, settings); err != nil {
			return errors.Join(fmt.Errorf("failed to unmarshal appsettings file"), err)
		}
	}

	// Step 2: Override with environment variables
	if err := envconfig.InitWithOptions(settings, envconfig.Options{AllOptional: true}); err != nil {
		return errors.Join(fmt.Errorf("failed to update settings with env vars"), err)
	}

	//Step 3: Validate settings in case a validator is provided
	if validator == nil {
		return nil
	}
	if errs := validator.Validate(settings); errs != nil {
		return errors.Join(fmt.Errorf("failed to validate settings"), errs)
	}
	return nil
}
