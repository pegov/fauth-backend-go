package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Config struct {
	Host            string `env:"HOST"`
	Port            string `env:"PORT"`
	DatabaseURL     string `env:"DATABASE_URL"`
	CacheURL        string `env:"CACHE_URL"`
	RecaptchaSecret string `env:"RECAPTCHA_SECRET"`
	Debug           bool   `env:"DEBUG"`
	LoginRatelimit  int    `env:"LOGIN_RATELIMIT"`
}

func New() (*Config, error) {
	cfg := Config{}
	var missingEnvs []string
	var wrongEnvs []string

	v := reflect.ValueOf(cfg)
	fields := reflect.VisibleFields(v.Type())
	for _, field := range fields {
		f, _ := reflect.TypeOf(cfg).FieldByName(field.Name)
		name := f.Tag.Get("env")
		value := reflect.ValueOf(cfg).FieldByName(f.Name).Interface()
		switch value.(type) {
		case string:
			v, ok := readString(name)
			if ok {
				reflect.ValueOf(&cfg).Elem().FieldByName(f.Name).Set(reflect.ValueOf(v))
			} else {
				missingEnvs = append(missingEnvs, name)
			}
		case bool:
			v, ok := readString(name)
			if ok {
				v := v == "1" || v == "True" || v == "TRUE" || v == "true" || v == "yes"
				reflect.ValueOf(&cfg).Elem().FieldByName(f.Name).Set(reflect.ValueOf(v))
			} else {
				missingEnvs = append(missingEnvs, name)
			}
		case int:
			v, ok := readString(name)
			if ok {
				if n, err := strconv.Atoi(v); err == nil {
					reflect.ValueOf(&cfg).Elem().FieldByName(f.Name).Set(reflect.ValueOf(n))
				} else {
					wrongEnvs = append(wrongEnvs, name)
				}
			} else {
				missingEnvs = append(missingEnvs, name)
			}
		}
	}

	if len(missingEnvs) > 0 || len(wrongEnvs) > 0 {
		var s string
		if len(missingEnvs) > 0 {
			s += fmt.Sprintf("missing envs: %s", strings.Join(missingEnvs, ", "))
		}
		if len(wrongEnvs) > 0 {
			if len(missingEnvs) > 0 {
				s += ", "
			}
			s += fmt.Sprintf("wrong envs: %s", strings.Join(missingEnvs, ", "))
		}
		return &cfg, fmt.Errorf(s)
	}

	return &cfg, nil
}

func (cfg *Config) Pretty() []byte {
	prettyCfg, _ := json.MarshalIndent(cfg, "", "\t")
	return prettyCfg
}

func readString(name string) (string, bool) {
	v, ok := os.LookupEnv(name)
	if !ok || v == "" {
		return "", false
	}

	return v, true
}
