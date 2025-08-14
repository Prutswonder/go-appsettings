# go-appsettings

[![Go](https://github.com/Prutswonder/go-appsettings/actions/workflows/go.yml/badge.svg)](https://github.com/Prutswonder/go-appsettings/actions/workflows/go.yml)

go-appsettings is a library that supports the use of application settings similar to .Net. It reads from a JSON source and can be configured to override settings with environment variables. Validation can be applied to ensure that the end result contains valid settings.



uses [envconfig](https://github.com/vrischmann/envconfig) to override JSON settings with environment.

## How it works

You can use `appsettings.NewComposer()` to instantiate a new `Composer`. By default the file `appsettings.json` is used to load the initial application settings. You can also use `appsettings.NewComposerWithReader()` to provide your own custom JSON reader in case you want to provide it from a different source.

Next, you need a struct that holds your settings. Make sure that the setting fields are public.

Optionally, you can update the settings from the JSON file with a struct that matches the `Updater` interface. For example, here is one that uses [envconfig](https://github.com/vrischmann/envconfig):

```go
type EnvVarUpdater struct { }

func (u *EnvVarUpdater) Update(settings any) error {
	return envconfig.InitWithOptions(settings, envconfig.Options{AllOptional: true});
}
```

You can use the `WithUpdater()` method to link the `Updater` to the `Composer`, like this:

```go
composer = composer.WithUpdater(&EnvVarUpdater{})
```

Likewise, you can use the `WithValidator()` method to link a `Validator` to the `Composer`. For example, if you want to use the [validator](https://github.com/go-validator/validator) package, you can link it like this:

```go
composer = composer.WithValidator(validator.NewValidator())
```

## Reading settings from JSON

The json parameter names are resolved using the dot notation. For example, `Global{Log{Level}}` will be resolved to `Global.Log.Level` or `global.log.level`, following Go's JSON unmarshaling implemetation. In case you want to use different JSON names, you can override them using the `json` tag. For example `json:"msg-level"` will allow you to use the JSON parameter `msg-level`.

## Using environment variables

In case you're using [envconfig](https://github.com/vrischmann/envconfig) to update the settings struct with environment variables, it is good to know about the default naming convention. 

The environment variable names are resolved using uppercase names and using underscores for nesting. For example, `Global{Log{Level}}` will be resolved to `GLOBAL_LOG_LEVEL`.

## Example

`appsettings.go`
```go
package main

type AppSettings struct {
	Global struct {
		Log struct {
			Level string `json:"msg-level"`
		}
	}
	Cors struct {
		Origins []string
	}
	Custom struct {
		Service struct {
			Name string
		}
		Enabled bool
	}
	Google struct {
		Application struct {
			Credentials string
		}
	}
}
// Implementation of the Updator interface that encapsulates envconfig
func (s *AppSettings) Update(settings any) error {
	return envconfig.InitWithOptions(settings, envconfig.Options{AllOptional: true});
}
// Custom implementation of the Validator interface
func (s *AppSettings) Validate(settings any) error {
	errs := error(nil)
	if s.Global.Log.Level == "" {
		errs = errors.Join(errs, fmt.Errorf("Global.Log.Level is required"))
	}
	if s.Google.App.Credentials == "" {
		errs = errors.Join(errs, fmt.Errorf("Google.App.Credentials is required"))
	}
	return errs
}

```

`appsettings.json`
```json
{
  "global": {
    "log": {
      "msg-level": "Debug"
    }
  },
  "cors": {
    "origins": ["*"]
  }
}
```


`main.go`
```go
package main

import (
	"fmt"

	"github.com/vrischmann/envconfig"
	"github.com/prutswonder/go-appsettings/appsettings"
	"gopkg.in/validator.v2"
)

func main() {
	settings := AppSettings{}

	composer, err := appsettings.NewComposer()
	if err != nil {
		panic(err)
	}

	// Note that settings also holds the Updater and Validator implementations
	composer.WithUpdater(&settings).WithValidator(&settings)

	if err = sut.Read(&settings); err != nil {
		panic(err)
	}

	fmt.Printf("Settings loaded: %+v\n", settings)
}
```

Note that this example will fail if the environment variable `GOOGLE_APPLICATION_CREDENTIALS` is missing. This can be resolved by either removing the `Google` struct from `AppSettings` in case you're not using it, or by removing the validation code in the `Validate()` method.