// Command unirate is a small, dependency-light command-line interface for the
// UniRate API (https://unirateapi.com) — currency conversion, exchange rates,
// VAT rates, and (Pro-gated) historical data, straight from your terminal.
//
//	unirate convert 100 USD EUR
//	unirate rate USD EUR
//	unirate rates EUR
//	unirate currencies
//	unirate vat DE
//
// The API key is read from the UNIRATE_API_KEY environment variable or the
// --api-key flag. Get a free key at https://unirateapi.com.
package main

import (
	"os"

	"github.com/UniRate-API/unirate-cli/internal/cli"
)

// version is overwritten at build time by goreleaser via
// -ldflags "-X main.version=<tag>". It stays "dev" for `go install` builds.
var version = "dev"

func main() {
	app := &cli.App{
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Getenv:  os.Getenv,
		Version: version,
	}
	os.Exit(app.Run(os.Args[1:]))
}
