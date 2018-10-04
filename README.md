# Common logging infrastructure

This `log` package is taken from [Istio](https://istio.io/) and provides the canonical logging functionality used by Go-based components.

Istio's logging subsystem is built on top of the [Zap](https:godoc.org/go.uber.org/zap) package.
High performance scenarios should use the Error, Warn, Info, and Debug methods. Lower perf
scenarios can use the more expensive convenience methods such as Debugf and Warnw.

The package provides direct integration with the Cobra command-line processor which makes it
easy to build programs that use a consistent interface for logging. Here's an example
of a simple Cobra-based program using this log package:

```go
func main() {
    // get the default logging options
    options := log.DefaultOptions()

    rootCmd := &cobra.Command{
        Run: func(cmd *cobra.Command, args []string) {

            // configure the logging system
            if err := log.Configure(options); err != nil {
                // print an error and quit
            }

            // output some logs
            log.Info("Hello")
            log.Sync()
        },
    }

    // add logging-specific flags to the cobra command
    options.AttachFlags(rootCmd)
    rootCmd.SetArgs(os.Args[1:])
    rootCmd.Execute()
}
```

Once configured, this package intercepts the output of the standard golang "log" package as well as anything
sent to the global zap logger (`zap.L()`).

## Installing

The log package can be installed using `go get`:

    go get -u github.com/tetratelabs/log

To build it from source just run:

    make

## License

This sowftare is licensed under the Apache License 2.0. See LICENSE file for details.
