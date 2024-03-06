package config

import (
	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/config/validation"
)

const (
	backendCurrent            = v1.ConfigBackends + "." + v1.BackendCurrent
	backendConfig             = v1.ConfigBackends + "." + "*"
	backendConfigType         = backendConfig + "." + v1.BackendType
	backendConfigItems        = backendConfig + "." + v1.BackendConfigItems
	backendLocalPath          = backendConfigItems + "." + v1.BackendLocalPath
	backendMysqlDBName        = backendConfigItems + "." + v1.BackendMysqlDBName
	backendMysqlUser          = backendConfigItems + "." + v1.BackendMysqlUser
	backendMysqlPassword      = backendConfigItems + "." + v1.BackendMysqlPassword
	backendMysqlHost          = backendConfigItems + "." + v1.BackendMysqlHost
	backendMysqlPort          = backendConfigItems + "." + v1.BackendMysqlPort
	backendGenericOssEndpoint = backendConfigItems + "." + v1.BackendGenericOssEndpoint
	backendGenericOssAK       = backendConfigItems + "." + v1.BackendGenericOssAK
	backendGenericOssSK       = backendConfigItems + "." + v1.BackendGenericOssSK
	backendGenericOssBucket   = backendConfigItems + "." + v1.BackendGenericOssBucket
	backendGenericOssPrefix   = backendConfigItems + "." + v1.BackendGenericOssPrefix
	backendS3Region           = backendConfigItems + "." + v1.BackendS3Region
)

func newRegisteredItems() map[string]*itemInfo {
	return map[string]*itemInfo{
		backendCurrent:            {"", validation.ValidateCurrentBackend, nil},
		backendConfig:             {&v1.BackendConfig{}, validation.ValidateBackendConfig, validation.ValidateUnsetBackendConfig},
		backendConfigType:         {"", validation.ValidateBackendType, validation.ValidateUnsetBackendType},
		backendConfigItems:        {map[string]any{}, validation.ValidateBackendConfigItems, nil},
		backendLocalPath:          {"", validation.ValidateLocalBackendItem, nil},
		backendMysqlDBName:        {"", validation.ValidateMysqlBackendItem, nil},
		backendMysqlUser:          {"", validation.ValidateMysqlBackendItem, nil},
		backendMysqlPassword:      {"", validation.ValidateMysqlBackendItem, nil},
		backendMysqlHost:          {"", validation.ValidateMysqlBackendItem, nil},
		backendMysqlPort:          {0, validation.ValidateMysqlBackendPort, nil},
		backendGenericOssEndpoint: {"", validation.ValidateGenericOssBackendItem, nil},
		backendGenericOssAK:       {"", validation.ValidateGenericOssBackendItem, nil},
		backendGenericOssSK:       {"", validation.ValidateGenericOssBackendItem, nil},
		backendGenericOssBucket:   {"", validation.ValidateGenericOssBackendItem, nil},
		backendGenericOssPrefix:   {"", validation.ValidateGenericOssBackendItem, nil},
		backendS3Region:           {"", validation.ValidateS3BackendItem, nil},
	}
}

// itemInfo includes necessary information of the config item, which is used when getting, setting and unsetting
// the config item.
type itemInfo struct {
	// zeroValue is the zero value of the type that the config item will be parsed from string to. Support string,
	// int, bool, map, slice, struct and pointer of struct, the parser rule is shown as below:
	//	- string: keep the same
	//	- int: calling strconv.Atoi, e.g. "45" is valid, parsed to 45
	// 	- bool: calling strconv.ParseBool, e.g. "true" is valid, parsed to true
	//	- slice, map, struct(pointer of struct): calling json.Unmarshal, zeroValue of these types must be
	// 	initialized, e.g. map[string]any{}, nil is invalid
	// For other unsupported types, calling json.Unmarshal to do the parse job, unexpected error or panic may
	// happen. Please do not use them.
	zeroValue any

	// validateFunc is used to check the config item is valid or not to set, calling before executing real
	// config setting. The unregistered config item, empty item value and invalid item value type is forbidden
	// by config operator by default, which are unnecessary to check in the validateFunc.
	// Please do not do any real setting job in the validateFunc.
	validateFunc validation.ValidateFunc

	// validateUnsetFunc is used to check the config item is valid or not to unset, calling before executing
	// real config unsetting.
	// Please do not do any real unsetting job in the validateUnsetFunc.
	validateUnsetFunc validation.ValidateUnsetFunc
}
