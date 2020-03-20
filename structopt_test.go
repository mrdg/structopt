package structopt

import (
	"flag"
	"log"
	"net"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"
)

type Config struct {
	String   string        `opt:"string"`
	Int      int           `opt:"int"`
	Int64    int64         `opt:"int64"`
	Uint64   uint64        `opt:"uint64"`
	Float64  float64       `opt:"float64"`
	Duration time.Duration `opt:"duration"`
	URL      url.URL       `opt:"url"`
	Bool     bool          `opt:"bool"`
}

const prefix = "APP"

func TestEnv(t *testing.T) {
	os.Setenv("APP_DURATION", "1s")
	os.Setenv("APP_INT", "25")
	os.Setenv("APP_INT64", "9223372036854775807")
	os.Setenv("APP_UINT64", "10")
	os.Setenv("APP_FLOAT64", "1.2")
	os.Setenv("APP_BOOL", "true")
	os.Setenv("APP_STRING", "string from env")
	os.Setenv("APP_URL", "https://example.com/env")

	conf := &Config{}
	if err := Load(prefix, conf, nil); err != nil {
		t.Fatal(err)
	}

	if want, got := time.Second, conf.Duration; !reflect.DeepEqual(want, got) {
		t.Errorf("want %v, got %v", want, got)
	}
	if want, got := 25, conf.Int; want != got {
		t.Errorf("want %v, got %v", want, got)
	}
	if want, got := int64(9223372036854775807), conf.Int64; want != got {
		t.Errorf("want %v, got %v", want, got)
	}
	if want, got := uint64(10), conf.Uint64; want != got {
		t.Errorf("want %v, got %v", want, got)
	}
	if want, got := float64(1.2), conf.Float64; want != got {
		t.Errorf("want %v, got %v", want, got)
	}
	if want, got := "string from env", conf.String; want != got {
		t.Errorf("want %q, got %q", want, got)
	}
	if want, got := "/env", conf.URL.Path; want != got {
		t.Errorf("want %q, got %q", want, got)
	}
	if !conf.Bool {
		t.Error("verbose should have been true")
	}
}

func TestFlags(t *testing.T) {
	// We're testing flags, but set env vars anyway to check that
	// they're being overridden.
	env := map[string]string{
		"APP_DURATION": "1s",
		"APP_INT":      "25",
		"APP_INT64":    "9223372036854775807",
		"APP_UINT64":   "10",
		"APP_FLOAT64":  "1.2",
		"APP_BOOL":     "true",
		"APP_STRING":   "string from env",
		"APP_URL":      "https://example.com/env",
	}
	for k, v := range env {
		os.Setenv(k, v)
	}

	flags := []string{
		"-duration", "2s",
		"-int", "50",
		"-int64", "1",
		"-uint64", "1",
		"-float64", "2.4",
		"-bool=false",
		"-string", "string from flag",
		"-url", "https://example.com/flag",
	}

	os.Args = append([]string{""}, flags...)

	conf := &Config{}
	fs := flag.NewFlagSet("TestFlags", flag.ContinueOnError)
	if err := Load(prefix, conf, fs); err != nil {
		log.Fatal(err)
	}
	if want, got := time.Second*2, conf.Duration; !reflect.DeepEqual(want, got) {
		t.Errorf("want %v, got %v", want, got)
	}
	if want, got := 50, conf.Int; want != got {
		t.Errorf("want %v, got %v", want, got)
	}
	if want, got := int64(1), conf.Int64; want != got {
		t.Errorf("want %v, got %v", want, got)
	}
	if want, got := uint64(1), conf.Uint64; want != got {
		t.Errorf("want %v, got %v", want, got)
	}
	if want, got := float64(2.4), conf.Float64; want != got {
		t.Errorf("want %v, got %v", want, got)
	}
	if want, got := "string from flag", conf.String; want != got {
		t.Errorf("want %q, got %q", want, got)
	}
	if want, got := "/flag", conf.URL.Path; want != got {
		t.Errorf("want %q, got %q", want, got)
	}
	if conf.Bool {
		t.Error("verbose should have been false")
	}
}

func TestInvalidVavlue(t *testing.T) {
	type Test struct {
		D time.Duration `opt:"d"`
	}

	test := Test{}
	os.Setenv("APP_D", "foo")
	err := Load(prefix, &test, nil)
	if err == nil {
		t.Fatalf("expected an error")
	}
}

func TestUnsupportedType(t *testing.T) {
	type T int
	type Test struct {
		T T `opt:"T"`
	}
	err := Load(prefix, new(Test), nil)
	if err == nil {
		t.Fatalf("expected an error")
	}
}

func TestNonStructConfig(t *testing.T) {
	type T int
	err := Load(prefix, new(T), nil)
	if err == nil {
		t.Fatalf("expected an error")
	}
}

func TestUserDefinedType(t *testing.T) {
	type Test struct {
		MACAddress *macAddr `opt:"mac.address"`
	}
	test := Test{&macAddr{}}
	os.Setenv("APP_MAC_ADDRESS", "aa:bb:cc:dd:ee:ff")
	os.Args = []string{"", "-mac.address", "ff:ee:dd:cc:bb:aa"}

	if err := Load(prefix, &test, flag.CommandLine); err != nil {
		t.Fatal(err)
	}
	if want, got := "ff:ee:dd:cc:bb:aa", test.MACAddress.val.String(); want != got {
		t.Fatalf("want %v, got %v", want, got)
	}
}

type macAddr struct {
	val net.HardwareAddr
}

func (m macAddr) String() string {
	if m.val == nil {
		return ""
	}
	return m.val.String()
}

func (m *macAddr) Set(s string) error {
	mac, err := net.ParseMAC(s)
	if err != nil {
		return err
	}
	m.val = mac
	return nil
}
