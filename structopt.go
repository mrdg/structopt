package structopt

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

const (
	flagSeps  = ".-"
	envSep    = "_"
	tagPrefix = "opt"
)

type option struct {
	iface    interface{}
	ptr      unsafe.Pointer
	envVar   string
	flagName string
	flagDesc string
}

func (o option) set(flags *flag.FlagSet) error {
	var (
		iface   = o.iface
		fromEnv = os.Getenv(o.envVar)
		err     error
	)
	switch iface.(type) {
	case string:
		sp := (*string)(o.ptr)
		*sp = fromEnv
		flags.StringVar(sp, o.flagName, *sp, o.flagDesc)
	case bool:
		bp := (*bool)(o.ptr)
		*bp, err = strconv.ParseBool(fromEnv)
		flags.BoolVar(bp, o.flagName, *bp, o.flagDesc)
	case int:
		ip := (*int)(o.ptr)
		*ip, err = strconv.Atoi(fromEnv)
		flags.IntVar(ip, o.flagName, *ip, o.flagDesc)
	case uint64:
		up := (*uint64)(o.ptr)
		*up, err = strconv.ParseUint(fromEnv, 10, 64)
		flags.Uint64Var(up, o.flagName, *up, o.flagDesc)
	case int64:
		ip := (*int64)(o.ptr)
		*ip, err = strconv.ParseInt(fromEnv, 10, 64)
		flags.Int64Var(ip, o.flagName, *ip, o.flagDesc)
	case float64:
		fp := (*float64)(o.ptr)
		*fp, err = strconv.ParseFloat(fromEnv, 64)
		flags.Float64Var(fp, o.flagName, *fp, o.flagDesc)
	case time.Duration:
		dp := (*time.Duration)(o.ptr)
		*dp, err = time.ParseDuration(fromEnv)
		flags.DurationVar(dp, o.flagName, *dp, o.flagDesc)
	case url.URL:
		up := (*url.URL)(o.ptr)
		var u *url.URL
		u, err = url.Parse(fromEnv)
		*up = *u
		flags.Var(&urlValue{u: up}, o.flagName, o.flagDesc)
	case value:
		v := iface.(value)
		err = v.Set(fromEnv)
		flags.Var(v, o.flagName, o.flagDesc)
	default:
		return fmt.Errorf("unsupported field type: %v", reflect.TypeOf(o.iface))
	}
	if err != nil && strings.TrimSpace(fromEnv) != "" {
		// Only return the error if there was something to parse.
		return fmt.Errorf("parse error %s=%q: %w", o.envVar, fromEnv, err)
	}
	return nil
}

type value interface {
	Set(s string) error
	String() string
}

func inferOptions(prefix string, config interface{}) ([]option, error) {
	if v := reflect.ValueOf(config); v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("config must be a pointer to struct, got %v", reflect.TypeOf(config))
	}
	structType := reflect.TypeOf(config).Elem()
	structVal := reflect.ValueOf(config).Elem()

	var opts []option
	for i := 0; i < structType.NumField(); i++ {
		val := structVal.Field(i)
		typ := structType.Field(i)
		if !val.CanSet() {
			continue
		}
		tag := typ.Tag.Get(tagPrefix)
		if len(tag) == 0 {
			continue
		}

		envVar := strings.ToUpper(tag)
		for _, sep := range flagSeps {
			strings.ReplaceAll(envVar, string(sep), envSep)
		}
		opt := option{
			iface:    val.Interface(),
			ptr:      unsafe.Pointer(val.Addr().Pointer()),
			flagName: tag,
			envVar:   prefix + envSep + envVar,
		}
		opts = append(opts, opt)
	}
	return opts, nil
}

// Load parses environment varibles for each field on config that has an "opt"
// tag. If flags is non-nil, it will be used to define command line flags.
func Load(prefix string, config interface{}, flags *flag.FlagSet) error {
	opts, err := inferOptions(prefix, config)
	if err != nil {
		return err
	}
	fs := flags
	if fs == nil {
		// Caller doesn't need flags, but create a FlagSet anyway to keep things simple. This
		// doesn't get parsed.
		fs = flag.NewFlagSet(prefix, flag.ExitOnError)
	}
	for _, opt := range opts {
		if err := opt.set(fs); err != nil {
			return err
		}
	}
	if flags != nil {
		args := os.Args[1:]
		if err := fs.Parse(args); err != nil {
			return err
		}
	}
	return nil
}

type urlValue struct {
	u *url.URL
}

func (v *urlValue) String() string {
	if v.u == nil {
		return ""
	}
	return v.u.String()
}

func (v *urlValue) Set(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	*v.u = *u
	return nil
}
