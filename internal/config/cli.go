package config

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

func CustomizeCLI() {
	helpPrinterCustomOrig := cli.HelpPrinterCustom
	cli.HelpPrinterCustom = func(out io.Writer, templ string, data any, customFuncs map[string]any) {
		if customFuncs == nil {
			customFuncs = map[string]any{}
		}
		customFuncs["isEmpty"] = func(v any) bool {
			return reflect.ValueOf(v).IsZero()
		}
		helpPrinterCustomOrig(out, templ, data, customFuncs)
	}

	cli.FlagStringer = customStringer
}

func customStringer(f cli.Flag) string {
	// enforce DocGeneration interface on flags to avoid reflection
	df, ok := f.(cli.DocGenerationFlag)
	if !ok {
		return ""
	}
	placeholder, usage := unquoteUsage(df.GetUsage())
	needsPlaceholder := df.TakesValue()
	// if needsPlaceholder is true, placeholder is empty
	if needsPlaceholder && placeholder == "" {
		// try to get type from flag
		if tname := df.TypeName(); tname != "" {
			placeholder = tname
		} else {
			placeholder = "value"
		}
	}

	defaultValueString := ""
	if rf, ok := f.(cli.RequiredFlag); !ok || !rf.IsRequired() {
		isVisible := df.IsDefaultVisible()
		if s := df.GetDefaultText(); isVisible && s != "" {
			defaultValueString = fmt.Sprintf(formatDefault("%s"), s)
		}
	}

	defaultValueString = strings.TrimSpace(defaultValueString)

	pn := prefixedNames(f.Names(), placeholder)
	sliceFlag, ok := f.(cli.DocGenerationMultiValueFlag)
	if ok && sliceFlag.IsMultiValueFlag() {
		pn = pn + " [ " + pn + " ]"
	}
	first := fmt.Sprintf("%s\t%s", pn, usage)
	envS := withEnvHint(df.GetEnvVars())

	var required bool
	if v, ok := f.(cli.RequiredFlag); ok {
		required = v.IsRequired()
	}

	var isBoolFlag bool
	if v, ok := f.(*cli.BoolFlag); ok {
		isBoolFlag = v.IsBoolFlag()
	}

	isTerm := term.IsTerminal(int(os.Stdout.Fd()))
	ansiReset := "\033[0m"
	ansiRed := "\033[31m"
	ansiYellow := "\033[33m"
	ansiGreen := "\033[32m"

	var after string
	if required {
		if isTerm {
			after += ansiRed
		}
		after += "REQUIRED"
		if isTerm {
			after += ansiReset
		}
	} else {
		v := reflect.ValueOf(f.Get())
		sl := v.Kind() == reflect.Slice && v.Len() == 0
		p := v.Kind() == reflect.Ptr && v.IsNil()
		if v.IsZero() || sl || p {
			if isBoolFlag {
				if isTerm {
					after += ansiGreen
				}
				after += defaultValueString
				if isTerm {
					after += ansiReset
				}
			} else {
				if defaultValueString != "" {
					after += defaultValueString
				} else {
					if isTerm {
						after += ansiYellow
					}
					after += "OPTIONAL"
					if isTerm {
						after += ansiReset
					}
				}
			}
		} else {
			if isTerm {
				after += ansiGreen
			}
			after += defaultValueString
			if isTerm {
				after += ansiReset
			}
		}
	}

	if after != "" {
		after = "\t" + after
	}

	if envS != "" {
		envS = "\t" + envS
	}

	return first + after + envS
}

func unquoteUsage(usage string) (string, string) {
	for i := range len(usage) {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name := usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break
		}
	}
	return "", usage
}

func formatDefault(format string) string {
	return " [default: " + format + "]"
}

func prefixedNames(names []string, placeholder string) string {
	var prefixed string
	for i, name := range names {
		if name == "" {
			continue
		}

		prefixed += prefixFor(name) + name
		if placeholder != "" {
			prefixed += " " + placeholder
		}
		if i < len(names)-1 {
			prefixed += ", "
		}
	}
	return prefixed
}

func prefixFor(name string) (prefix string) {
	if len(name) == 1 {
		prefix = "-"
	} else {
		prefix = "--"
	}

	return
}

func withEnvHint(envVars []string) string {
	envText := ""
	if runtime.GOOS != "windows" || os.Getenv("PSHOME") != "" {
		envText = defaultEnvFormat(envVars)
	} else {
		envText = envFormat(envVars, "%", "%, %", "%")
	}
	return envText
}

func defaultEnvFormat(envVars []string) string {
	return envFormat(envVars, "", ", ", "")
}

func envFormat(envVars []string, prefix, sep, suffix string) string {
	if len(envVars) > 0 {
		return fmt.Sprintf(" [env: %s%s%s]", prefix, strings.Join(envVars, sep), suffix)
	}
	return ""
}
