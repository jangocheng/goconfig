package goEnv

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/crgimenes/goConfig/structTag"
)

// Usage is the function that is called when an error occurs.
var Usage func()

// Setup maps and variables
func Setup(tag string, tagDefault string) {
	Usage = DefaultUsage

	structTag.Setup()
	SetTag(tag)
	SetTagDefault(tagDefault)

	structTag.ParseMap[reflect.Int] = reflectInt
	structTag.ParseMap[reflect.String] = reflectString
}

// SetTag set a new tag
func SetTag(tag string) {
	structTag.Tag = tag
}

// SetTagDefault set a new TagDefault to retorn default values
func SetTagDefault(tag string) {
	structTag.TagDefault = tag
}

// Parse configuration
func Parse(config interface{}) (err error) {
	err = structTag.Parse(config, "")
	return
}

var PrintDefaultsOutput string

func getNewValue(field *reflect.StructField, value *reflect.Value, tag string, datatype string) (ret string) {

	defaultValue := field.Tag.Get(structTag.TagDefault)

	// create PrintDefaults output
	tag = strings.ToUpper(tag)
	if runtime.GOOS == "windows" {
		if defaultValue == "" {
			PrintDefaultsOutput += `  %` + tag + `% ` + datatype + "\n\n"
		} else {
			printDV := " (default \"" + defaultValue + "\")"
			PrintDefaultsOutput += `  %` + tag + `% ` + datatype + "\n\t" + printDV + "\n"
		}
	} else {
		if defaultValue == "" {
			PrintDefaultsOutput += "  $" + tag + " " + datatype + "\n\n"
		} else {
			printDV := " (default \"" + defaultValue + "\")"
			PrintDefaultsOutput += "  $" + tag + " " + datatype + "\n\t" + printDV + "\n"
		}
	}

	// get value from environment variable
	ret = os.Getenv(tag)
	if ret != "" {
		return
	}

	// get value from default settings
	ret = defaultValue

	return
}

func reflectInt(field *reflect.StructField, value *reflect.Value, tag string) (err error) {
	newValue := getNewValue(field, value, tag, "int")
	if newValue == "" {
		return
	}

	var intNewValue int64
	intNewValue, err = strconv.ParseInt(newValue, 10, 64)
	if err != nil {
		return
	}

	value.SetInt(intNewValue)

	return
}

func reflectString(field *reflect.StructField, value *reflect.Value, tag string) (err error) {
	newValue := getNewValue(field, value, tag, "string")
	if newValue == "" {
		return
	}

	value.SetString(newValue)

	return
}

func PrintDefaults() {
	fmt.Println("Environment variables:")
	fmt.Println(PrintDefaultsOutput)

}

func DefaultUsage() {
	fmt.Println("Usage")
	PrintDefaults()
}
