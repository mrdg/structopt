/*
Package structopt allows you to parse command line flags and environment
variables by defining a configuration struct:

	package main

	type Config struct {
		QueryTimeout time.Duration `opt:"query.timeout"`
		UserName     string `opt:"username"`
	}

	func main() {
		conf := &Config{}
		if err := structopt.Load('APP', conf, flag.CommandLine); err != nil {
			log.Fatal(err)
		}
	}


A struct tag determines both the name as well as the environment variable name
for a particular struct field. In the example above, a query timeout can be
configured by  passing the flag -query.timeout, setting an environment variable
APP_QUERY_TIMEOUT, or by setting a value on the struct directly. This is also
the order of precendence.

Supported field types are string, bool, int, uint64, int64, float64, time.Duration,
 url.URL plus any type that implements the flag.Value interface.

*/
package structopt
