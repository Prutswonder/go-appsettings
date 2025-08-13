# go-appsettings

[![Go](https://github.com/Prutswonder/go-appsettings/actions/workflows/go.yml/badge.svg)](https://github.com/Prutswonder/go-appsettings/actions/workflows/go.yml)

go-appsettings is a library that supports the use of application settings similar to .Net. It uses [envconfig](https://github.com/vrischmann/envconfig) to override JSON settings with environment variables and [Package validator](https://github.com/go-validator/validator) for validations.

## How it works

The file `appsettings.json` contains the default values of the application settings. These are loaded first.

After loading the app settings stored in the JSON file, the `AppSettings` struct is used to collect any settings from environment variables. 

The json parameter names are resolved using the dot notation. For example, `Global{Log{Level}}` will be resolved to `Global.Log.Level` or `global.log.level`, following Go's JSON unmarshaling implemetation. In case you want to use different JSON names, you can override them using the `json` tag. For example `json:"msg-level"` will allow you to use the JSON parameter `msg-level`.

The environment variable names are resolved using uppercase names and using underscores for nesting. For example, `Global{Log{Level}}` will be resolved to `GLOBAL_LOG_LEVEL`.

## Validation

By default any settings in `appsettings.json` are optional. You can use validation tags that are defined by Package Validator ([see documentation](https://pkg.go.dev/gopkg.in/validator.v2)). For example, you can use the `validate:"nonzero"` tag to indicate any setting that is mandatory. Validations are executed after the environment variables are merged. 

If a field is mandatory, using the `validate:"nonzero"` means that if the setting does not have a value in the JSON file, it needs to be set by an environment variable. 

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
			Credentials string `validate:"nonzero"`
		}
	}
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

	"github.com/prutswonder/go-appsettings/appsettings"
)

func main() {
	settings := AppSettings{}

	err := appsettings.ReadSettingsFromFileAndEnv(&settings)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Settings loaded: %+v\n", settings)
}
```

Note that this example will fail if the environment variable `GOOGLE_APPLICATION_CREDENTIALS` is missing. You can resolve this by either removing the `Google` struct in case you're not using it, or by removing the `validate:"nonzero"` tag from the `Credentials` field.