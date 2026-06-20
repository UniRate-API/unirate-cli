#!/usr/bin/env bash
# Example session for the unirate CLI. Set a free API key first:
#   export UNIRATE_API_KEY=...   # get one at https://unirateapi.com
set -euo pipefail

# Convert an amount between two currencies
unirate convert 100 USD EUR

# Show the exchange rate for a single pair
unirate rate USD JPY

# All rates for a base currency (defaults to USD)
unirate rates EUR

# List every supported currency code
unirate currencies

# VAT rate for one country, or all of them
unirate vat DE
unirate vat

# Machine-readable output for scripting / jq
unirate rates USD --json | jq '.rates.EUR'

# Historical data and time series are Pro-gated (403 on the free tier)
unirate historical 2024-01-01 USD EUR
unirate timeseries 2024-01-01 2024-01-07 --currencies EUR,GBP
