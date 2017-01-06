package goConfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Settings default
type Settings struct {
	// Path sets default config path
	Path string
	// File name of default config file
	File string
	// FileRequired config file required
	FileRequired bool
	// Tag set the main tag
	Tag string
	// TagDefault set tag default
	TagDefault string
	// TagDisabled used to not process an input
	TagDisabled string
	// EnvironmentVarSeparator separe names on environment variables
	EnvironmentVarSeparator string
}

// Setup Pointer to internal variables
var Setup *Settings

// ErrNotAPointer error when not a pointer
var ErrNotAPointer = errors.New("Not a pointer")

// ErrNotAStruct error when not a struct
var ErrNotAStruct = errors.New("Not a struct")

// ErrTypeNotSupported error when type not supported
var ErrTypeNotSupported = errors.New("Type not supported")

// ReflectFunc type used to create funcrions to parse struct and tags
type ReflectFunc func(
	field *reflect.StructField,
	value *reflect.Value,
	tag string) (err error)

var parseMap map[reflect.Kind]ReflectFunc

func init() {
	Setup = &Settings{
		Path:                    "./",
		File:                    "config.json",
		Tag:                     "cfg",
		TagDefault:              "cfgDefault",
		TagDisabled:             "-",
		EnvironmentVarSeparator: "_",
		FileRequired:            false,
	}

	parseMap = make(map[reflect.Kind]ReflectFunc)

	parseMap[reflect.Struct] = reflectStruct
	parseMap[reflect.Int] = reflectInt
	parseMap[reflect.String] = reflectString

}

// LoadJSON config file
func LoadJSON(config interface{}) (err error) {
	configFile := Setup.Path + Setup.File
	file, err := os.Open(configFile)
	if os.IsNotExist(err) && !Setup.FileRequired {
		err = nil
		return
	} else if err != nil {
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return
	}

	return
}

// Load config file
func Load(config interface{}) (err error) {

	err = LoadJSON(config)
	if err != nil {
		return
	}

	err = parseTags(config, "")
	if err != nil {
		return
	}

	postProc()

	return
}

// Save config file
func Save(config interface{}) (err error) {
	_, err = os.Stat(Setup.Path)
	if os.IsNotExist(err) {
		os.Mkdir(Setup.Path, 0700)
	} else if err != nil {
		return
	}

	configFile := Setup.Path + Setup.File

	_, err = os.Stat(configFile)
	if err != nil {
		return
	}

	b, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return
	}

	err = ioutil.WriteFile(configFile, b, 0644)
	if err != nil {
		return
	}
	return
}

func parseTags(s interface{}, superTag string) (err error) {

	st := reflect.TypeOf(s)

	if st.Kind() != reflect.Ptr {
		err = ErrNotAPointer
		return
	}

	refField := st.Elem()
	if refField.Kind() != reflect.Struct {
		err = ErrNotAStruct
		return
	}

	//vt := reflect.ValueOf(s)
	refValue := reflect.ValueOf(s).Elem()
	for i := 0; i < refField.NumField(); i++ {
		field := refField.Field(i)
		value := refValue.Field(i)
		kind := field.Type.Kind()

		if field.PkgPath != "" {
			continue
		}

		t := updateTag(&field, superTag)
		if t == "" {
			continue
		}

		if f, ok := parseMap[kind]; ok {
			err = f(&field, &value, t)
			if err != nil {
				return
			}
		} else {
			log.Println("Type not supported" + kind.String())
			err = ErrTypeNotSupported
			return
		}

		fmt.Println("name:", field.Name,
			"| value", value,
			"| cfg:", field.Tag.Get(Setup.Tag),
			"| cfgDefault:", field.Tag.Get(Setup.TagDefault),
			"| type:", field.Type)

	}

	return
}

func updateTag(field *reflect.StructField, superTag string) (ret string) {
	ret = field.Tag.Get(Setup.Tag)
	if ret == Setup.TagDisabled {
		ret = ""
		return
	}

	if ret == "" {
		ret = field.Name
	}

	if superTag != "" {
		ret = superTag + Setup.EnvironmentVarSeparator + ret
	}
	return
}

func getNewValue(field *reflect.StructField, value *reflect.Value, tag string) (ret string) {

	//TODO: get value from parameter.

	// get value from environment variable
	ret = os.Getenv(strings.ToUpper(tag))
	if ret != "" {
		return
	}

	// get value from config file
	switch value.Kind() {
	case reflect.String:
		ret = value.String()
		return
	case reflect.Int:
		ret = strconv.FormatInt(value.Int(), 10)
		return
	}

	// get value from default settings
	ret = field.Tag.Get(Setup.TagDefault)

	return
}

func postProc() {
}

func reflectStruct(field *reflect.StructField, value *reflect.Value, tag string) (err error) {
	err = parseTags(value.Addr().Interface(), tag)
	return
}

func reflectInt(field *reflect.StructField, value *reflect.Value, tag string) (err error) {
	//value.SetInt(999)

	newValue := getNewValue(field, value, tag)

	var intNewValue int64
	intNewValue, err = strconv.ParseInt(newValue, 10, 64)
	if err != nil {
		return
	}
	value.SetInt(intNewValue)

	return
}

func reflectString(field *reflect.StructField, value *reflect.Value, tag string) (err error) {
	//value.SetString("TEST")

	newValue := getNewValue(field, value, tag)

	value.SetString(newValue)

	return
}
