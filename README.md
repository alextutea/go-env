# go-env

Lightweight go module which helps with importing environment variables and/or configuration files at run time.

## Quick-start example
```go
package main

import "github.com/alextutea/go-env"

type envVars struct {
	TimeoutSeconds int       `env:"TIMEOUT_SECONDS,default=60"`
	Port           int       `env:"PORT,default=8080"`
	Mongo          mongoVars `env:"MONGO"`
}

type mongoVars struct {
	URI        string `env:"URI,required=true"`
	DBName     string `env:"DB_NAME,default=geo"`
	Collection string `env:"COLLECTION,default=datapoints"`
}

func main(){
	e := envVars{}
	err := env.Unmarshal(&e)
	if err != nil {
		//handle error 
	}
	
	//now e contains values imported from the environment
	//based on the keys specified in the env tag of the struct
}
```

## Features
This module attempts to simplify a number of common use cases when it comes to passing configuration settings into Go applications.
The current features of the module are presented below:

### Required variables & Default values
The env struct can be configured by specifying certain options in the tags of each field. Currently, two tag options are supported:
- **required** - if the required option is set to true, then env.Unmarshal will return an error if there is no env var set for the key corresponding to the required field
- **default**  - if the default option is present, then, if there is no env var specified for the key corresponding to the field, then the field will be populated by env.Unmarshal with the default value specified in the tag option
### Typecasting
In the env struct, each field can only be one of the following types: `bool`, `string`, `int`, `int8`, `int16`, `int32`, `int64`, `float32`, `float64` or another `struct` in the case of nested env structs.
When calling env.Unmarshal the values of each env var, which are organically just strings, will be parsed into the correct type specified by the field of the env struct. IF the casting fails, env.Unmarshal returns an error.
### Nested structs
The env struct can recursively have one or more struct fields. The env vars get unmarshalled based on the concatenation of the env keys of the nested structs.
e.g. in the snippet below 5 variables can be successfully unmarshalled into the env struct: `TIMEOUT_SECONDS`, `PORT`, `MONGO_URI`, `MONGO_DB_NAME` and `MONGO_COLLECTION`.
```go
type envVars struct {
	TimeoutSeconds int       `env:"TIMEOUT_SECONDS,default=60"`
	Port           int       `env:"PORT,default=8080"`
	Mongo          mongoVars `env:"MONGO"`
}

type mongoVars struct {
	URI        string `env:"URI,required=true"`
	DBName     string `env:"DB_NAME,default=geo"`
	Collection string `env:"COLLECTION,default=datapoints"`
}

```
### Prioritized import between env vars and config file(s)
The env.Unmarshal method accepts, besides a reference to the env struct, an optional list of file paths.
These file paths should be absolute paths to config files in JSON format. An example of a config file could look as shown below:
```json
{
  "TIMEOUT_SECONDS": 4,
  "PORT": 8080,
  "MONGO": {
    "URI": "some.mongo.connection.uri",
    "DB_NAME": "someDB",
    "COLLECTION": "someCollectionName"
  }
}
```
There is another important thing to note: when passing one or more config files into env.Unmarshal (e.g. `env.Unmarshal(&e, filePath1, filePath2)`),
the env vars from the OS environment have priority over the ones specified in the configs, which in turn have priority over the default values specified in the default tag options of the struct
The module will always unmarshal the value from the source with the highest priority.

`OS ENVIRONMENT > CONFIG.JSON > DEFAULT VALS`
