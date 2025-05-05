package config

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

type CustomFlag struct {
	Parent *CustomFlag
	Name   string
	Type   string
}

type ParseOptions struct {
	Parent                  *ParseOptions
	EnvPrefix               string
	EnvIsDisabled           bool
	FlagPrefix              string
	Category                string
	customFlagPrefix        bool
	customEnvPrefix         bool
	AlreadyHasDefaultValues bool
	RequiredByDefault       bool
}

func ParseFlags(c any, opts ParseOptions) ([]cli.Flag, error) {
	if c == nil {
		return nil, errors.New("config must not be nil")
	}

	v := reflect.ValueOf(c)

	if v.Kind() != reflect.Ptr {
		return nil, errors.New("config must be pointer")
	}

	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return nil, errors.New("config must be struct")
	}

	t := reflect.TypeOf(c).Elem()

	flags := make([]cli.Flag, 0, v.NumField())
	for i := range v.NumField() {
		res, err := parseField(t.Field(i), v.Field(i), opts)
		if err != nil {
			return nil, err
		}
		flags = append(flags, res...)
	}

	return flags, nil
}

type flagOptions[T any] struct {
	Value T
	Dest  *T
	flagOptionsCommon
}

type flagOptionsCommon struct {
	Name       string
	Category   string
	HasValue   bool
	Env        string
	DisableEnv bool
	Usage      string
	Required   bool
}

func parseField(
	t reflect.StructField,
	v reflect.Value,
	opts ParseOptions,
) ([]cli.Flag, error) {
	var envPrefix, flagPrefix string
	if opts.EnvPrefix != "" {
		envPrefix = opts.EnvPrefix + "_"
	}
	if opts.FlagPrefix != "" {
		flagPrefix = opts.FlagPrefix + "-"
	}

	argName, ok := t.Tag.Lookup("name")
	if !ok {
		argName = flagPrefix + toKebabCase(t.Name)
	}

	disableEnv := opts.EnvIsDisabled
	var envName string
	if !disableEnv {
		envName, ok = t.Tag.Lookup("env")
		if !ok {
			envName = envPrefix + toScreamingSnakeCase(t.Name)
		} else {
			if envName == "-" {
				disableEnv = true
			}
		}
	}

	if !v.CanSet() {
		return nil, fmt.Errorf("private field: %s", t.Name)
	}

	var strValue string
	var hasStrValue bool
	if !opts.AlreadyHasDefaultValues {
		strValue, hasStrValue = t.Tag.Lookup("default")
	}

	usage, _ := t.Tag.Lookup("usage")

	var (
		cliRequired bool
		cliOptional bool
	)
	cliOptionsStr, _ := t.Tag.Lookup("cli")
	if cliOptionsStr != "" {
		cliOptions := strings.Split(cliOptionsStr, ",")
		cliRequired = slices.Contains(cliOptions, "required")
		cliOptional = slices.Contains(cliOptions, "optional")
	}

	if !cliOptional {
		cliRequired = cliRequired || opts.RequiredByDefault
	}

	foc := flagOptionsCommon{
		Name:       argName,
		Category:   opts.Category,
		HasValue:   false,
		Env:        envName,
		DisableEnv: disableEnv,
		Usage:      usage,
		Required:   cliRequired,
	}

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	iface := v.Addr().Interface()

	switch v.Kind() {
	case reflect.Slice:
		sv := v.Type().Elem().Kind()
		switch sv {
		case reflect.String:
			dst, ok := iface.(*[]string)
			if !ok {
				return nil, fmt.Errorf("failed to cast *[]string: %s", t.Name)
			}

			fo := flagOptions[[]string]{
				flagOptionsCommon: foc,
			}

			if opts.AlreadyHasDefaultValues {
				fo.HasValue = true
				fo.Value = *dst
			} else if hasStrValue {
				fo.HasValue = true
				fo.Value = strings.Split(strValue, ",")
			}

			if fo.HasValue {
				fo.Required = false
			}

			return []cli.Flag{stringSliceFlag(fo)}, nil
		default:
			return nil, fmt.Errorf("slice type %v is unsupported", sv)
		}

	case reflect.Struct:
		envPrefixFromTag, hasEnvPrefixFromTag := t.Tag.Lookup("env")
		if hasEnvPrefixFromTag {
			envPrefix = envPrefix + envPrefixFromTag
		} else {
			envPrefix = envPrefix + toScreamingSnakeCase(t.Name)
		}

		flagPrefixFromTag, hasFlagPrefixFromTag := t.Tag.Lookup("flag")
		if hasFlagPrefixFromTag {
			flagPrefix = flagPrefix + flagPrefixFromTag
		} else {
			flagPrefix = flagPrefix + toKebabCase(t.Name)
		}

		newOpts := ParseOptions{
			Parent:            &opts,
			Category:          t.Name,
			EnvPrefix:         envPrefix,
			EnvIsDisabled:     opts.EnvIsDisabled || envPrefixFromTag == "-",
			FlagPrefix:        flagPrefix,
			RequiredByDefault: cliRequired,
		}
		return ParseFlags(iface, newOpts)

	case reflect.TypeOf(time.Duration(0)).Kind():
		dst, ok := iface.(*time.Duration)
		if !ok {
			return nil, fmt.Errorf("failed to cast *time.Duration: %s", t.Name)
		}

		fo := flagOptions[time.Duration]{
			flagOptionsCommon: foc,
			Dest:              dst,
		}

		if opts.AlreadyHasDefaultValues {
			fo.HasValue = true
			fo.Value = *dst
		} else if hasStrValue {
			v, err := time.ParseDuration(strValue)
			if err != nil {
				return nil, fmt.Errorf("invalid duration format for %s: %v", t.Name, err)
			}
			fo.HasValue = true
			fo.Value = v
		}

		if fo.HasValue {
			fo.Required = false
		}

		return []cli.Flag{durationFlag(fo)}, nil

	case reflect.String:
		dst, ok := iface.(*string)
		if !ok {
			return nil, fmt.Errorf("failed to cast *string: %s", t.Name)
		}

		fo := flagOptions[string]{
			flagOptionsCommon: foc,
			Dest:              dst,
		}

		if opts.AlreadyHasDefaultValues {
			fo.HasValue = true
			fo.Value = *dst
		} else if hasStrValue {
			fo.HasValue = true
			fo.Value = strValue
		}

		if fo.HasValue {
			fo.Required = false
		}

		return []cli.Flag{stringFlag(fo)}, nil

	case reflect.Int:
		dst, ok := iface.(*int)
		if !ok {
			return nil, fmt.Errorf("failed to cast *int: %s", t.Name)
		}

		fo := flagOptions[int]{
			flagOptionsCommon: foc,
			Dest:              dst,
		}

		if opts.AlreadyHasDefaultValues {
			fo.HasValue = true
			fo.Value = *dst
		} else if hasStrValue {
			v, err := strconv.Atoi(strValue)
			if err != nil {
				return nil, err
			}
			fo.HasValue = true
			fo.Value = v
		}

		if fo.HasValue {
			fo.Required = false
		}

		return []cli.Flag{intFlag(fo)}, nil

	case reflect.Int8:
		dst, ok := iface.(*int8)
		if !ok {
			return nil, fmt.Errorf("failed to cast *int8: %s", t.Name)
		}

		fo := flagOptions[int8]{
			flagOptionsCommon: foc,
			Dest:              dst,
		}

		if opts.AlreadyHasDefaultValues {
			fo.HasValue = true
			fo.Value = *dst
		} else if hasStrValue {
			v, err := strconv.ParseInt(strValue, 10, 8)
			if err != nil {
				return nil, err
			}
			fo.HasValue = true
			fo.Value = int8(v)
		}

		if fo.HasValue {
			fo.Required = false
		}

		return []cli.Flag{int8Flag(fo)}, nil

	case reflect.Int16:
		dst, ok := iface.(*int16)
		if !ok {
			return nil, fmt.Errorf("failed to cast *int16: %s", t.Name)
		}

		fo := flagOptions[int16]{
			flagOptionsCommon: foc,
			Dest:              dst,
		}

		if opts.AlreadyHasDefaultValues {
			fo.HasValue = true
			fo.Value = *dst
		} else if hasStrValue {
			v, err := strconv.ParseInt(strValue, 10, 16)
			if err != nil {
				return nil, err
			}
			fo.HasValue = true
			fo.Value = int16(v)
		}

		if fo.HasValue {
			fo.Required = false
		}

		return []cli.Flag{int16Flag(fo)}, nil

	case reflect.Int32:
		dst, ok := iface.(*int32)
		if !ok {
			return nil, fmt.Errorf("failed to cast *int32: %s", t.Name)
		}

		fo := flagOptions[int32]{
			flagOptionsCommon: foc,
			Dest:              dst,
		}

		if opts.AlreadyHasDefaultValues {
			fo.HasValue = true
			fo.Value = *dst
		} else if hasStrValue {
			v, err := strconv.ParseInt(strValue, 10, 32)
			if err != nil {
				return nil, err
			}
			fo.HasValue = true
			fo.Value = int32(v)
		}

		if fo.HasValue {
			fo.Required = false
		}

		return []cli.Flag{int32Flag(fo)}, nil

	case reflect.Int64:
		dst, ok := iface.(*int64)
		if !ok {
			return nil, fmt.Errorf("failed to cast *int64: %s", t.Name)
		}

		fo := flagOptions[int64]{
			flagOptionsCommon: foc,
			Dest:              dst,
		}

		if opts.AlreadyHasDefaultValues {
			fo.HasValue = true
			fo.Value = *dst
		} else if hasStrValue {
			v, err := strconv.ParseInt(strValue, 10, 64)
			if err != nil {
				return nil, err
			}
			fo.HasValue = true
			fo.Value = v
		}

		if fo.HasValue {
			fo.Required = false
		}

		return []cli.Flag{int64Flag(fo)}, nil

	case reflect.Uint:
		dst, ok := iface.(*uint)
		if !ok {
			return nil, fmt.Errorf("failed to cast *uint: %s", t.Name)
		}

		fo := flagOptions[uint]{
			flagOptionsCommon: foc,
			Dest:              dst,
		}

		if opts.AlreadyHasDefaultValues {
			fo.HasValue = true
			fo.Value = *dst
		} else if hasStrValue {
			v, err := strconv.ParseUint(strValue, 10, 0)
			if err != nil {
				return nil, err
			}
			fo.HasValue = true
			fo.Value = uint(v)
		}

		if fo.HasValue {
			fo.Required = false
		}

		return []cli.Flag{uintFlag(fo)}, nil

	case reflect.Uint8:
		dst, ok := iface.(*uint8)
		if !ok {
			return nil, fmt.Errorf("failed to cast *uint8: %s", t.Name)
		}

		fo := flagOptions[uint8]{
			flagOptionsCommon: foc,
			Dest:              dst,
		}

		if opts.AlreadyHasDefaultValues {
			fo.HasValue = true
			fo.Value = *dst
		} else if hasStrValue {
			v, err := strconv.ParseUint(strValue, 10, 8)
			if err != nil {
				return nil, err
			}
			fo.HasValue = true
			fo.Value = uint8(v)
		}

		if fo.HasValue {
			fo.Required = false
		}

		return []cli.Flag{uint8Flag(fo)}, nil

	case reflect.Uint16:
		dst, ok := iface.(*uint16)
		if !ok {
			return nil, fmt.Errorf("failed to cast *uint16: %s", t.Name)
		}

		fo := flagOptions[uint16]{
			flagOptionsCommon: foc,
			Dest:              dst,
		}

		if opts.AlreadyHasDefaultValues {
			fo.HasValue = true
			fo.Value = *dst
		} else if hasStrValue {
			v, err := strconv.ParseUint(strValue, 10, 16)
			if err != nil {
				return nil, err
			}
			fo.HasValue = true
			fo.Value = uint16(v)
		}

		if fo.HasValue {
			fo.Required = false
		}

		return []cli.Flag{uint16Flag(fo)}, nil

	case reflect.Uint32:
		dst, ok := iface.(*uint32)
		if !ok {
			return nil, fmt.Errorf("failed to cast *uint32: %s", t.Name)
		}

		fo := flagOptions[uint32]{
			flagOptionsCommon: foc,
			Dest:              dst,
		}

		if opts.AlreadyHasDefaultValues {
			fo.HasValue = true
			fo.Value = *dst
		} else if hasStrValue {
			v, err := strconv.ParseUint(strValue, 10, 32)
			if err != nil {
				return nil, err
			}
			fo.HasValue = true
			fo.Value = uint32(v)
		}

		if fo.HasValue {
			fo.Required = false
		}

		return []cli.Flag{uint32Flag(fo)}, nil

	case reflect.Uint64:
		dst, ok := iface.(*uint64)
		if !ok {
			return nil, fmt.Errorf("failed to cast *uint64: %s", t.Name)
		}

		fo := flagOptions[uint64]{
			flagOptionsCommon: foc,
			Dest:              dst,
		}

		if opts.AlreadyHasDefaultValues {
			fo.HasValue = true
			fo.Value = *dst
		} else if hasStrValue {
			v, err := strconv.ParseUint(strValue, 10, 64)
			if err != nil {
				return nil, err
			}
			fo.HasValue = true
			fo.Value = v
		}

		if fo.HasValue {
			fo.Required = false
		}

		return []cli.Flag{uint64Flag(fo)}, nil

	case reflect.Float64:
		dst, ok := iface.(*float64)
		if !ok {
			return nil, fmt.Errorf("failed to cast *float64: %s", t.Name)
		}

		fo := flagOptions[float64]{
			flagOptionsCommon: foc,
			Dest:              dst,
		}

		if opts.AlreadyHasDefaultValues {
			fo.HasValue = true
			fo.Value = *dst
		} else if hasStrValue {
			v, err := strconv.ParseFloat(strValue, 64)
			if err != nil {
				return nil, err
			}
			fo.HasValue = true
			fo.Value = v
		}

		if fo.HasValue {
			fo.Required = false
		}

		return []cli.Flag{float64Flag(fo)}, nil

	case reflect.Bool:
		dst, ok := iface.(*bool)
		if !ok {
			return nil, fmt.Errorf("failed to cast *bool: %s", t.Name)
		}

		fo := flagOptions[bool]{
			flagOptionsCommon: foc,
			Dest:              dst,
		}

		if opts.AlreadyHasDefaultValues {
			fo.HasValue = true
			fo.Value = *dst
		} else if hasStrValue {
			v, err := strconv.ParseBool(strValue)
			if err != nil {
				return nil, err
			}
			fo.HasValue = true
			fo.Value = v
		}

		if fo.HasValue {
			fo.Required = false
		}

		return []cli.Flag{boolFlag(fo)}, nil

	default:
		return nil, fmt.Errorf("type %v is unsupported", v)
	}
}

func stringFlag(opts flagOptions[string]) *cli.StringFlag {
	flag := &cli.StringFlag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage, Required: opts.Required}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

func stringSliceFlag(opts flagOptions[[]string]) *cli.StringSliceFlag {
	flag := &cli.StringSliceFlag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage, Required: opts.Required}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

func boolFlag(opts flagOptions[bool]) *cli.BoolFlag {
	flag := &cli.BoolFlag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

func intFlag(opts flagOptions[int]) *cli.IntFlag {
	flag := &cli.IntFlag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage, Required: opts.Required}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

func int8Flag(opts flagOptions[int8]) *cli.Int8Flag {
	flag := &cli.Int8Flag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage, Required: opts.Required}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

func int16Flag(opts flagOptions[int16]) *cli.Int16Flag {
	flag := &cli.Int16Flag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage, Required: opts.Required}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

func int32Flag(opts flagOptions[int32]) *cli.Int32Flag {
	flag := &cli.Int32Flag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage, Required: opts.Required}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

func int64Flag(opts flagOptions[int64]) *cli.Int64Flag {
	flag := &cli.Int64Flag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage, Required: opts.Required}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

func uintFlag(opts flagOptions[uint]) *cli.UintFlag {
	flag := &cli.UintFlag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage, Required: opts.Required}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

func uint8Flag(opts flagOptions[uint8]) *cli.Uint8Flag {
	flag := &cli.Uint8Flag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage, Required: opts.Required}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

func uint16Flag(opts flagOptions[uint16]) *cli.Uint16Flag {
	flag := &cli.Uint16Flag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage, Required: opts.Required}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

func uint32Flag(opts flagOptions[uint32]) *cli.Uint32Flag {
	flag := &cli.Uint32Flag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage, Required: opts.Required}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

func uint64Flag(opts flagOptions[uint64]) *cli.Uint64Flag {
	flag := &cli.Uint64Flag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage, Required: opts.Required}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

func float64Flag(opts flagOptions[float64]) *cli.FloatFlag {
	flag := &cli.FloatFlag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage, Required: opts.Required}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

func durationFlag(opts flagOptions[time.Duration]) *cli.DurationFlag {
	flag := &cli.DurationFlag{Name: opts.Name, Category: opts.Category, Destination: opts.Dest, Usage: opts.Usage, Required: opts.Required}
	if opts.HasValue {
		flag.Value = opts.Value
	}
	if !opts.DisableEnv {
		flag.Sources = cli.EnvVars(opts.Env)
	}
	return flag
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func toKebabCase(str string) string {
	return strings.ReplaceAll(toSnakeCase(str), "_", "-")
}

func toScreamingSnakeCase(str string) string {
	return strings.ToUpper(toSnakeCase(str))
}
