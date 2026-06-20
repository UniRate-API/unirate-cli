// Package cli implements the unirate command-line interface. It is kept
// separate from package main so the whole dispatch path is testable with an
// injected stdout/stderr and a fake environment, and so commands can be
// pointed at an httptest server via the hidden --base-url flag.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	unirate "github.com/UniRate-API/unirate-api-go"
)

// App carries the IO and environment dependencies for a single CLI run.
// Tests construct an App with buffers and a fake Getenv so no global state or
// real environment is touched.
type App struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Getenv  func(string) string
	Version string
}

// exit codes
const (
	exitOK    = 0
	exitError = 1 // runtime / API error
	exitUsage = 2 // bad invocation
)

// errMissingKey is returned when no API key is found via flag or environment.
var errMissingKey = errors.New("no API key: set UNIRATE_API_KEY or pass --api-key")

// Run dispatches argv (os.Args[1:]) to the matching subcommand and returns the
// process exit code. It never panics on bad input — usage problems return
// exitUsage and API failures return exitError.
func (a *App) Run(argv []string) int {
	if len(argv) == 0 {
		a.usage(a.Stderr)
		return exitUsage
	}

	cmd, rest := argv[0], argv[1:]
	switch cmd {
	case "convert":
		return a.cmdConvert(rest)
	case "rate":
		return a.cmdRate(rest)
	case "rates":
		return a.cmdRates(rest)
	case "currencies":
		return a.cmdCurrencies(rest)
	case "vat":
		return a.cmdVAT(rest)
	case "historical":
		return a.cmdHistorical(rest)
	case "timeseries":
		return a.cmdTimeSeries(rest)
	case "version", "--version", "-v":
		fmt.Fprintf(a.Stdout, "unirate %s\n", a.Version)
		return exitOK
	case "help", "-h", "--help":
		a.usage(a.Stdout)
		return exitOK
	default:
		fmt.Fprintf(a.Stderr, "unirate: unknown command %q\n\n", cmd)
		a.usage(a.Stderr)
		return exitUsage
	}
}

// commonOpts holds the flags shared by every subcommand.
type commonOpts struct {
	apiKey  string
	json    bool
	timeout time.Duration
	baseURL string
}

// newFlagSet builds a FlagSet for a subcommand, wired to write errors to the
// app's stderr, with the common flags registered. The returned commonOpts is
// populated after fs.Parse runs.
func (a *App) newFlagSet(name string) (*flag.FlagSet, *commonOpts) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	o := &commonOpts{}
	fs.StringVar(&o.apiKey, "api-key", "", "UniRate API key (defaults to $UNIRATE_API_KEY)")
	fs.BoolVar(&o.json, "json", false, "emit machine-readable JSON instead of text")
	fs.DurationVar(&o.timeout, "timeout", 30*time.Second, "per-request timeout")
	fs.StringVar(&o.baseURL, "base-url", "", "override the API base URL (advanced/testing)")
	return fs, o
}

// client constructs a UniRate client from the parsed common options,
// resolving the API key from the flag then the environment.
func (a *App) client(o *commonOpts) (*unirate.Client, error) {
	key := o.apiKey
	if key == "" {
		key = a.Getenv("UNIRATE_API_KEY")
	}
	if key == "" {
		return nil, errMissingKey
	}
	opts := []unirate.Option{unirate.WithTimeout(o.timeout)}
	if o.baseURL != "" {
		opts = append(opts, unirate.WithBaseURL(o.baseURL))
	}
	return unirate.New(key, opts...), nil
}

// ctx returns a context bounded by the request timeout. The caller must invoke
// the returned cancel func.
func (o *commonOpts) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), o.timeout)
}

// fail prints a friendly, classified error to stderr and returns exitError.
// It translates the client's sentinel/typed errors into actionable messages
// rather than leaking the wrapped Go error chain.
func (a *App) fail(err error) int {
	var msg string
	var apiErr *unirate.APIError
	switch {
	case errors.Is(err, errMissingKey):
		msg = err.Error()
	case errors.Is(err, unirate.ErrAuthentication):
		msg = "authentication failed — check your API key (UNIRATE_API_KEY)"
	case errors.Is(err, unirate.ErrRateLimit):
		msg = "rate limit exceeded — slow down or upgrade your plan"
	case errors.Is(err, unirate.ErrInvalidCurrency):
		msg = "currency or country not found, or no data available for it"
	case errors.Is(err, unirate.ErrInvalidDate):
		msg = "invalid request parameters (check the date format, YYYY-MM-DD)"
	case errors.As(err, &apiErr) && apiErr.StatusCode == 403:
		msg = "this endpoint requires a UniRate Pro subscription (historical/time-series)"
	default:
		msg = err.Error()
	}
	fmt.Fprintln(a.Stderr, "error: "+msg)
	return exitError
}

// usageErr reports a bad-invocation problem and returns exitUsage.
func (a *App) usageErr(format string, args ...any) int {
	fmt.Fprintf(a.Stderr, "error: "+format+"\n", args...)
	return exitUsage
}

func (a *App) usage(w io.Writer) {
	fmt.Fprint(w, `unirate — currency exchange rates from the UniRate API

Usage:
  unirate <command> [flags]

Commands:
  convert <amount> <from> <to>     Convert an amount between two currencies
  rate <from> <to>                 Show the exchange rate for one pair
  rates [base]                     Show all rates for a base currency (default USD)
  currencies                       List supported currency codes
  vat [country]                    Show VAT rates (all countries, or one ISO code)
  historical <date> <from> <to>    Historical rate on a date (Pro)  [YYYY-MM-DD]
  timeseries <start> <end>         Historical rates over a range (Pro)
  version                          Print the CLI version

Common flags:
  --api-key <key>    API key (defaults to $UNIRATE_API_KEY)
  --json             Emit JSON instead of human-readable text
  --timeout <dur>    Per-request timeout (default 30s)

Examples:
  unirate convert 100 USD EUR
  unirate rate USD JPY
  unirate rates EUR --json
  unirate vat DE

Get a free API key at https://unirateapi.com
`)
}
