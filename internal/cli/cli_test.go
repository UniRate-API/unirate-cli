package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// writeJSON sends a canned JSON body with the right content type.
func writeJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}

// okHandler routes every supported endpoint to a representative success body.
func okHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch r.URL.Path {
		case "/api/rates":
			if q.Get("to") != "" {
				writeJSON(w, `{"rate":"0.92"}`)
			} else {
				writeJSON(w, `{"rates":{"EUR":"0.92","GBP":"0.79","JPY":"150"}}`)
			}
		case "/api/convert":
			writeJSON(w, `{"result":"92.5"}`)
		case "/api/currencies":
			writeJSON(w, `{"currencies":["USD","EUR","GBP"]}`)
		case "/api/vat/rates":
			if q.Get("country") != "" {
				writeJSON(w, `{"country":"DE","vat_data":{"country_code":"DE","country_name":"Germany","vat_rate":19.0}}`)
			} else {
				writeJSON(w, `{"total_countries":2,"date":"2026-06-20","vat_rates":{"DE":{"country_code":"DE","country_name":"Germany","vat_rate":19.0},"FR":{"country_code":"FR","country_name":"France","vat_rate":20.0}}}`)
			}
		case "/api/historical/rates":
			if q.Get("amount") == "1" {
				writeJSON(w, `{"rate":"0.85"}`)
			} else {
				writeJSON(w, `{"result":"85"}`)
			}
		case "/api/historical/timeseries":
			writeJSON(w, `{"amount":1,"base":"USD","start_date":"2024-01-01","end_date":"2024-01-02","total_days":2,"currencies":["EUR","GBP"],"data":{"2024-01-01":{"EUR":0.92,"GBP":0.79},"2024-01-02":{"EUR":0.93,"GBP":0.80}}}`)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}
}

func statusHandler(code int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", code)
	}
}

// run executes the CLI against handler with key in the environment, appending
// --base-url so the client points at the test server. Returns exit code,
// stdout, stderr.
func run(t *testing.T, handler http.HandlerFunc, key string, args ...string) (int, string, string) {
	t.Helper()
	srv := httptest.NewServer(handler)
	defer srv.Close()

	var out, errb bytes.Buffer
	app := &App{
		Stdout: &out,
		Stderr: &errb,
		Getenv: func(k string) string {
			if k == "UNIRATE_API_KEY" {
				return key
			}
			return ""
		},
		Version: "1.2.3",
	}
	full := append([]string{}, args...)
	full = append(full, "--base-url", srv.URL)
	code := app.Run(full)
	return code, out.String(), errb.String()
}

// runBare executes the CLI without a server (for usage/version paths).
func runBare(key string, args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	app := &App{
		Stdout: &out,
		Stderr: &errb,
		Getenv: func(k string) string {
			if k == "UNIRATE_API_KEY" {
				return key
			}
			return ""
		},
		Version: "1.2.3",
	}
	code := app.Run(args)
	return code, out.String(), errb.String()
}

func TestConvert(t *testing.T) {
	code, out, errb := run(t, okHandler(), "k", "convert", "100", "usd", "eur")
	if code != 0 {
		t.Fatalf("exit %d, stderr=%q", code, errb)
	}
	if out != "100 USD = 92.5 EUR\n" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestConvertJSON(t *testing.T) {
	// --json placed AFTER positionals exercises interspersed-flag parsing.
	code, out, _ := run(t, okHandler(), "k", "convert", "100", "USD", "EUR", "--json")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output not JSON: %v\n%s", err, out)
	}
	if got["from"] != "USD" || got["to"] != "EUR" || got["result"].(float64) != 92.5 {
		t.Fatalf("unexpected JSON: %v", got)
	}
}

func TestConvertWrongArgs(t *testing.T) {
	code, _, errb := run(t, okHandler(), "k", "convert", "100", "USD")
	if code != 2 {
		t.Fatalf("want usage exit 2, got %d", code)
	}
	if !strings.Contains(errb, "convert needs") {
		t.Fatalf("missing usage message: %q", errb)
	}
}

func TestConvertInvalidAmount(t *testing.T) {
	code, _, errb := run(t, okHandler(), "k", "convert", "abc", "USD", "EUR")
	if code != 2 || !strings.Contains(errb, "invalid amount") {
		t.Fatalf("code=%d errb=%q", code, errb)
	}
}

func TestRate(t *testing.T) {
	code, out, _ := run(t, okHandler(), "k", "rate", "USD", "EUR")
	if code != 0 || out != "1 USD = 0.92 EUR\n" {
		t.Fatalf("code=%d out=%q", code, out)
	}
}

func TestRateJSON(t *testing.T) {
	code, out, _ := run(t, okHandler(), "k", "rate", "USD", "EUR", "--json")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if got["rate"].(float64) != 0.92 {
		t.Fatalf("unexpected: %v", got)
	}
}

func TestRates(t *testing.T) {
	code, out, _ := run(t, okHandler(), "k", "rates", "USD")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	want := "EUR 0.92\nGBP 0.79\nJPY 150\n"
	if out != want {
		t.Fatalf("got %q want %q", out, want)
	}
}

func TestRatesDefaultBase(t *testing.T) {
	// no positional → base defaults to USD, still succeeds.
	code, _, errb := run(t, okHandler(), "k", "rates")
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, errb)
	}
}

func TestRatesTooManyArgs(t *testing.T) {
	code, _, _ := run(t, okHandler(), "k", "rates", "USD", "EUR")
	if code != 2 {
		t.Fatalf("want 2 got %d", code)
	}
}

func TestCurrencies(t *testing.T) {
	code, out, _ := run(t, okHandler(), "k", "currencies")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if out != "EUR\nGBP\nUSD\n" {
		t.Fatalf("got %q", out)
	}
}

func TestCurrenciesJSON(t *testing.T) {
	code, out, _ := run(t, okHandler(), "k", "currencies", "--json")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var got struct {
		Currencies []string `json:"currencies"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if len(got.Currencies) != 3 || got.Currencies[0] != "EUR" {
		t.Fatalf("unexpected: %v", got.Currencies)
	}
}

func TestVATSingle(t *testing.T) {
	code, out, _ := run(t, okHandler(), "k", "vat", "de")
	if code != 0 || out != "DE (Germany): 19%\n" {
		t.Fatalf("code=%d out=%q", code, out)
	}
}

func TestVATAll(t *testing.T) {
	code, out, _ := run(t, okHandler(), "k", "vat")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out, "DE") || !strings.Contains(out, "Germany") || !strings.Contains(out, "19%") {
		t.Fatalf("missing DE row: %q", out)
	}
	// sorted: DE before FR
	if strings.Index(out, "DE") > strings.Index(out, "FR") {
		t.Fatalf("not sorted: %q", out)
	}
}

func TestVATSingleJSON(t *testing.T) {
	code, out, _ := run(t, okHandler(), "k", "vat", "DE", "--json")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out, `"vat_rate": 19`) {
		t.Fatalf("unexpected JSON: %q", out)
	}
}

func TestHistoricalRate(t *testing.T) {
	code, out, _ := run(t, okHandler(), "k", "historical", "2024-01-01", "USD", "EUR")
	if code != 0 || out != "1 USD = 0.85 EUR on 2024-01-01\n" {
		t.Fatalf("code=%d out=%q", code, out)
	}
}

func TestHistoricalConvert(t *testing.T) {
	code, out, _ := run(t, okHandler(), "k", "historical", "2024-01-01", "USD", "EUR", "--amount", "100")
	if code != 0 || out != "100 USD = 85 EUR on 2024-01-01\n" {
		t.Fatalf("code=%d out=%q", code, out)
	}
}

func TestHistoricalProGated(t *testing.T) {
	code, _, errb := run(t, statusHandler(http.StatusForbidden), "k", "historical", "2024-01-01", "USD", "EUR")
	if code != 1 {
		t.Fatalf("want 1 got %d", code)
	}
	if !strings.Contains(errb, "Pro subscription") {
		t.Fatalf("missing Pro message: %q", errb)
	}
}

func TestTimeSeries(t *testing.T) {
	code, out, _ := run(t, okHandler(), "k", "timeseries", "2024-01-01", "2024-01-02", "--currencies", "eur,gbp")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	want := "2024-01-01  EUR 0.92  GBP 0.79\n2024-01-02  EUR 0.93  GBP 0.8\n"
	if out != want {
		t.Fatalf("got %q want %q", out, want)
	}
}

func TestTimeSeriesJSON(t *testing.T) {
	code, out, _ := run(t, okHandler(), "k", "timeseries", "2024-01-01", "2024-01-02", "--json")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out, `"base": "USD"`) || !strings.Contains(out, `"2024-01-01"`) {
		t.Fatalf("unexpected JSON: %q", out)
	}
}

func TestErrorMappings(t *testing.T) {
	cases := []struct {
		status int
		want   string
	}{
		{http.StatusUnauthorized, "authentication failed"},
		{http.StatusNotFound, "not found"},
		{http.StatusTooManyRequests, "rate limit exceeded"},
		{http.StatusBadRequest, "invalid request parameters"},
		{http.StatusServiceUnavailable, "status 503"},
	}
	for _, c := range cases {
		code, _, errb := run(t, statusHandler(c.status), "k", "rate", "USD", "EUR")
		if code != 1 {
			t.Errorf("status %d: want exit 1 got %d", c.status, code)
		}
		if !strings.Contains(errb, c.want) {
			t.Errorf("status %d: stderr %q missing %q", c.status, errb, c.want)
		}
	}
}

func TestMissingKey(t *testing.T) {
	code, _, errb := run(t, okHandler(), "", "convert", "100", "USD", "EUR")
	if code != 1 {
		t.Fatalf("want 1 got %d", code)
	}
	if !strings.Contains(errb, "no API key") {
		t.Fatalf("missing key message: %q", errb)
	}
}

func TestAPIKeyFlagOverridesEnv(t *testing.T) {
	// env empty, but --api-key supplied → should still work.
	code, _, errb := run(t, okHandler(), "", "convert", "100", "USD", "EUR", "--api-key", "flagkey")
	if code != 0 {
		t.Fatalf("want 0 got %d stderr=%q", code, errb)
	}
}

func TestVersion(t *testing.T) {
	code, out, _ := runBare("k", "version")
	if code != 0 || out != "unirate 1.2.3\n" {
		t.Fatalf("code=%d out=%q", code, out)
	}
}

func TestHelp(t *testing.T) {
	code, out, _ := runBare("k", "-h")
	if code != 0 || !strings.Contains(out, "Usage:") {
		t.Fatalf("code=%d out=%q", code, out)
	}
}

func TestNoArgs(t *testing.T) {
	code, _, errb := runBare("k")
	if code != 2 || !strings.Contains(errb, "Usage:") {
		t.Fatalf("code=%d errb=%q", code, errb)
	}
}

func TestUnknownCommand(t *testing.T) {
	code, _, errb := runBare("k", "frobnicate")
	if code != 2 || !strings.Contains(errb, "unknown command") {
		t.Fatalf("code=%d errb=%q", code, errb)
	}
}
