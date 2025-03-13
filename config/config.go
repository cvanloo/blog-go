package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type TagData struct {
	Name      string
	Default   string
	Mandatory bool
}

func Load(cfg any) (err error) {
	cfgRefl := reflect.ValueOf(cfg)
	cfgType := cfgRefl.Type()
	if cfgType.Kind() != reflect.Pointer {
		panic("Load(cfg): cfg must be a pointer")
	}
	for _, field := range reflect.VisibleFields(cfgType.Elem()) {
		if field.Type.Kind() == reflect.String {
			optionName := optNameFromField(field.Name)
			td := parseTag(field.Tag.Get("cfg"))
			if td.Name != "" {
				optionName = td.Name
			}
			if val := cfgRefl.Elem().Field(field.Index[0]); val.IsValid() {
				if strVal := os.Getenv(optionName); strVal != "" {
					val.Set(reflect.ValueOf(strVal))
				} else if td.Default != "" {
					val.Set(reflect.ValueOf(td.Default))
				} else if td.Mandatory {
					err = errors.Join(err, fmt.Errorf("missing config value for mandatory key: %s", optionName))
				}
			}
		}
	}
	return err
}

func parseTag(rawTag string) (td TagData) {
	if rawTag == "" {
		return td
	}
	rawParts := strings.Split(rawTag, ";")
	for _, rawProperty := range rawParts {
		propertyParts := strings.Split(rawProperty, "=")
		if len(propertyParts) != 2 {
			panic(fmt.Sprintf("invalid format for property in cfg tag: %s", rawProperty)) // @todo: better error message (location?)
		}
		key := propertyParts[0]
		val := propertyParts[1]
		switch key {
		default:
			panic(fmt.Sprintf("unknown property in cfg tag: %s", key))
		case "name":
			td.Name = val
		case "default":
			td.Default = val
		case "mandatory":
			if val == "true" {
				td.Mandatory = true
			} else if val == "false" {
				td.Mandatory = false
			} else {
				panic(fmt.Sprintf("non-boolean value for property: mandatory=%s", val))
			}
		}
	}
	return td
}

func optNameFromField(name string) string {
	return strings.ToUpper(name)
}
