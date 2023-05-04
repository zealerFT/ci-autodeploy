package yaml

import (
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/rs/zerolog/log"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"
)

func ImageUpdate(yamlFile, containerName, iamgeName string) error {
	file, err := os.Open(yamlFile)
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	data, err := io.ReadAll(file)
	if err != nil {
		log.Err(err).Msgf("git ReadAll fail%v", err)
		return err
	}

	// var deploy Deployment
	var deploy map[any]any
	err = yaml.Unmarshal(data, &deploy)
	if err != nil {
		log.Err(err).Msgf("git yaml.Unmarshal fail%v", err)
		return err
	}

	if spec, ok := deploy["spec"]; !ok {
		return xerrors.Errorf("deploy spec type fail:%v", deploy)
	} else if specMap, err := AnyToMap(spec); err != nil {
		return xerrors.Errorf("deploy specMap type fail:%v", deploy)
	} else if template, ok := specMap["template"]; !ok {
		return xerrors.Errorf("deploy template type fail:%v", deploy)
	} else if templateMap, err := AnyToMap(template); err != nil {
		return xerrors.Errorf("deploy templateMap type fail:%v", deploy)
	} else if specIn, ok := templateMap["spec"]; !ok {
		return xerrors.Errorf("deploy specIn type fail:%v", deploy)
	} else if specInMap, err := AnyToMap(specIn); err != nil {
		return xerrors.Errorf("deploy specInMap type fail:%v", deploy)
	} else if containers, ok := specInMap["containers"]; !ok {
		return xerrors.Errorf("deploy containers type fail:%v", deploy)
	} else if containersSlice, err := AnyToSlice(containers); err != nil {
		return xerrors.Errorf("deploy containersSlice type fail:%v", deploy)
	} else if len(containersSlice) > 0 {
		log.Debug().Msgf("containersSlice:%v", containersSlice)
		for k, container := range containersSlice {
			if containerMap, err := AnyToMap(container); err != nil {
				return xerrors.Errorf("deploy single containerMap type fail:%v", deploy)
			} else {
				if containerMap["name"] == containerName {
					containerMap["image"] = iamgeName
					containersSlice[k] = containerMap
				}
			}

		}
		containers = containersSlice
		specInMap["containers"] = containers
		specIn = specInMap
		templateMap["spec"] = specIn
		template = templateMap
		specMap["template"] = template
		spec = specMap
		deploy["spec"] = spec
	} else {
		return xerrors.Errorf("deploy containersSlice is null:%v", deploy)
	}

	log.Debug().Msgf("deploy:%v", deploy)

	newData, err := yaml.Marshal(deploy)
	if err != nil {
		return err
	}
	err = os.WriteFile(yamlFile, newData, 0664)
	if err != nil {
		log.Err(err).Msgf("git WriteFile fail%v", err)
		return err
	}
	return nil
}

func AnyToMap(item interface{}) (map[string]interface{}, error) {
	value := reflect.ValueOf(item)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct && value.Kind() != reflect.Map {
		return nil, fmt.Errorf("invalid type %s", value.Kind())
	}
	result := make(map[string]interface{})
	var valueType reflect.Type
	if value.Kind() == reflect.Struct {
		valueType = value.Type()
		for i := 0; i < valueType.NumField(); i++ {
			key := valueType.Field(i).Name
			result[key] = value.Field(i).Interface()
		}
	} else if value.Kind() == reflect.Map {
		valueType = value.Type()
		keys := value.MapKeys()
		for _, key := range keys {
			result[key.String()] = value.MapIndex(key).Interface()
		}
	}
	return result, nil
}

func AnyToSlice(item interface{}) ([]interface{}, error) {
	value := reflect.ValueOf(item)
	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		return nil, fmt.Errorf("invalid type %s", value.Kind())
	}
	result := make([]interface{}, value.Len())
	for i := 0; i < value.Len(); i++ {
		result[i] = value.Index(i).Interface()
	}
	return result, nil
}
