# unirate CLI

Currency exchange rates, conversion, and VAT lookups from the
[UniRate API](https://unirateapi.com) — straight from your terminal.

```console
$ unirate convert 100 USD EUR
100 USD = 92.5 EUR

$ unirate rate USD JPY
1 USD = 150 JPY

$ unirate rates EUR --json | jq '.rates.USD'
1.08
```

- Real-time rates between 170+ currencies (fiat + crypto)
- Currency conversion, all-rates tables, and supported-currency listing
- VAT rates for countries worldwide
- Historical rates + time series (Pro)
- `--json` on every command for clean piping into `jq`
- Single static binary, **zero third-party dependencies** beyond the official
  [Go client](https://github.com/UniRate-API/unirate-api-go)
- Free tier, no credit card required — [get a key](https://unirateapi.com)

## Install

**Homebrew** (macOS):

```bash
brew install --cask UniRate-API/unirate/unirate
```

**Scoop** (Windows):

```powershell
scoop bucket add unirate https://github.com/UniRate-API/scoop-unirate
scoop install unirate
```

**Go** (any platform with a Go toolchain):

```bash
go install github.com/UniRate-API/unirate-cli@latest
```

**Binaries:** prebuilt archives for linux/macOS/windows (amd64 + arm64) are
attached to each [GitHub release](https://github.com/UniRate-API/unirate-cli/releases).

## Authentication

Every command needs an API key. Set it once in your environment:

```bash
export UNIRATE_API_KEY="your-api-key"
```

or pass it per-invocation with `--api-key`.

## Commands

| Command | Description |
|---|---|
| `unirate convert <amount> <from> <to>` | Convert an amount between two currencies |
| `unirate rate <from> <to>` | Exchange rate for a single pair |
| `unirate rates [base]` | All rates for a base currency (default `USD`) |
| `unirate currencies` | List supported currency codes |
| `unirate vat [country]` | VAT rates — all countries, or one ISO-3166 code |
| `unirate historical <date> <from> <to>` | Historical rate on a date (**Pro**) |
| `unirate timeseries <start> <end>` | Historical rates over a range (**Pro**) |
| `unirate version` | Print the CLI version |

### Common flags

Available on every data command:

- `--api-key <key>` — API key (defaults to `$UNIRATE_API_KEY`)
- `--json` — emit JSON instead of human-readable text
- `--timeout <dur>` — per-request timeout (default `30s`, e.g. `--timeout 5s`)

### Command-specific flags

- `historical --amount <n>` — convert `<n>` units instead of returning the unit rate
- `timeseries --base <code>` / `--amount <n>` / `--currencies <a,b,c>`

### Examples

```bash
unirate convert 49.99 USD GBP
unirate rate EUR USD --json
unirate rates JPY
unirate vat FR
unirate historical 2024-01-01 USD EUR --amount 250
unirate timeseries 2024-01-01 2024-01-07 --base USD --currencies EUR,GBP,JPY
```

See [`examples/usage.sh`](examples/usage.sh) for a full session.

## Error handling

The CLI maps API failures to friendly messages and exit codes:

| Exit code | Meaning |
|---|---|
| `0` | success |
| `1` | runtime / API error (auth, rate limit, unknown currency, Pro-gated, network) |
| `2` | usage error (bad arguments or flags) |

Historical and time-series endpoints are **Pro-gated** and return a clear
"requires a UniRate Pro subscription" message on the free tier.

## Distribution

Releases are cut by pushing a `v*` tag; [GoReleaser](https://goreleaser.com)
builds the cross-platform archives, checksums, and the Homebrew/Scoop manifests
in CI. The Homebrew tap and Scoop bucket live in sibling repos
(`UniRate-API/homebrew-unirate`, `UniRate-API/scoop-unirate`); pushing to them
requires the `TAP_GITHUB_TOKEN` repo secret (a PAT with `repo` scope).

## Related clients

This CLI wraps the official Go client. UniRate also ships native libraries for
many languages and frameworks.

<!-- unirate-ecosystem-footer:start -->
## Other UniRate clients

UniRate ships official client libraries and framework integrations across the
ecosystem. The repos below are all maintained under the
[UniRate-API](https://github.com/UniRate-API) org.

- **Languages:** [Python](https://github.com/UniRate-API/unirate-api-python) · [Node.js / TypeScript](https://github.com/UniRate-API/unirate-api-nodejs) · [Go](https://github.com/UniRate-API/unirate-api-go) · [Rust](https://github.com/UniRate-API/unirate-api-rust) · [Java](https://github.com/UniRate-API/unirate-api-java) · [Ruby](https://github.com/UniRate-API/unirate-api-ruby) · [PHP](https://github.com/UniRate-API/unirate-api-php) · [.NET](https://github.com/UniRate-API/unirate-api-dotnet) · [Swift](https://github.com/UniRate-API/unirate-api-swift)
- **Web frameworks:** [NestJS](https://github.com/UniRate-API/nestjs-unirate) · [Django / Wagtail](https://github.com/UniRate-API/wagtail-unirate) · [FastAPI](https://github.com/UniRate-API/fastapi-unirate) · [Flask](https://github.com/UniRate-API/flask-unirate) · [React](https://github.com/UniRate-API/react-unirate) · [tRPC](https://github.com/UniRate-API/trpc-unirate)
- **Static-site generators:** [Astro](https://github.com/UniRate-API/astro-unirate) · [Eleventy](https://github.com/UniRate-API/eleventy-unirate) · [Hugo](https://github.com/UniRate-API/hugo-unirate)
- **Data / orchestration:** [Airflow](https://github.com/UniRate-API/airflow-provider-unirate) · [dbt](https://github.com/UniRate-API/dbt-unirate) · [LangChain](https://github.com/UniRate-API/langchain-unirate)
- **Workflow / no-code:** [n8n](https://github.com/UniRate-API/n8n-nodes-unirate) · [Google Sheets](https://github.com/UniRate-API/unirate-sheets) · [MCP server](https://github.com/UniRate-API/unirate-mcp)
- **Editors / tools:** [VS Code](https://github.com/UniRate-API/vscode-unirate) · [Obsidian](https://github.com/UniRate-API/obsidian-currency)

Get a free API key at [unirateapi.com](https://unirateapi.com).
<!-- unirate-ecosystem-footer:end -->

## License

MIT — see [LICENSE](LICENSE).
