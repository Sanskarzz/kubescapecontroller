package main

import (
	"net/http"
	"os"
	"time"

	// To support the global flag used spf13 like --help etc
	"github.com/spf13/pflag"
	// Package server contains the plumbing to create kubernetes-like API server command
	"k8s.io/apiserver/pkg/server"
	// we need opetions package to create a secure server and want some default values/options
	"k8s.io/apiserver/pkg/server/options"
	// To support the global flag used spf13 like --help etc
	"k8s.io/component-base/cli/globalflag"
)

// used secure server to serve our webhook
type Options struct {
	SecureServingOptions options.SecureServingOptions
}

// AddFlagSet adds flags for a specific server to the specified FlagSet
func (o *Options) AddFlagSet(fs *pflag.FlagSet) {
	o.SecureServingOptions.AddFlags(fs)
}

// Config is the configuration for the webhook server
type Config struct {
	SecureServingInfo *server.SecureServingInfo
}

// Config returns the configuration for the webhook server
func (o *Options) Config() *Config {
	if err := o.SecureServingOptions.MaybeDefaultWithSelfSignedCerts("0.0.0.0", nil, nil); err != nil {
		panic(err)
	}
	c := Config{}
	o.SecureServingOptions.ApplyTo(&c.SecureServingInfo)
	return &c
}

const (
	controller = "controller"
)

// NewDefaultOptions returns a new Options with a default config.
func NewDefaultOptions() *Options {
	o := &Options{
		SecureServingOptions: *options.NewSecureServingOptions(),
	}
	o.SecureServingOptions.BindPort = 9000
	// if we don't specify certification then it will automatically create self-sign certificate
	// but in this example i have specified certificate  manually
	o.SecureServingOptions.ServerCert.PairName = controller
	return o
}

func main() {
	options := NewDefaultOptions()

	fs := pflag.NewFlagSet(controller, pflag.ExitOnError)
	globalflag.AddGlobalFlags(fs, controller)

	options.AddFlagSet(fs)

	if err := fs.Parse(os.Args); err != nil {
		panic(err)
	}

	c := options.Config()
	// mux is a http request multiplexer we can also rape it in some logging middleware
	// Then the output is show in a perticular format
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handleValidate))
	// Serve runs the secure http server.
	// It fails only if certificates cannot be loaded or the initial listen call fails.
	// The actual server loop (stoppable by closing stopCh) runs in a go routine,
	// i.e. Serve does not block.
	stopCh := server.SetupSignalHandler()
	ch, _, err := c.SecureServingInfo.Serve(mux, 30*time.Second, stopCh)
	if err != nil {
		panic(err)
	} else {
		<-ch
	}
}
