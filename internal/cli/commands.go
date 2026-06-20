package cli

import (
	"errors"
	"flag"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// parseArgs parses fs while allowing flags and positional arguments to be
// freely interleaved. The stdlib flag package stops at the first non-flag
// token, so `convert 100 USD EUR --json` would otherwise drop the trailing
// flag. We loop: parse, peel off the first leftover positional, parse the
// remainder, repeat. Returns positionals in order.
func parseArgs(fs *flag.FlagSet, args []string) ([]string, error) {
	var positionals []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			return positionals, nil
		}
		positionals = append(positionals, rest[0])
		args = rest[1:]
	}
}

// parse runs parseArgs and folds flag errors into an exit code. When ok is
// false the caller should return the supplied code (exitOK for -h/--help,
// exitUsage for a malformed flag — flag has already printed the detail).
func (a *App) parse(fs *flag.FlagSet, args []string) (pos []string, code int, ok bool) {
	pos, err := parseArgs(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil, exitOK, false
		}
		return nil, exitUsage, false
	}
	return pos, exitOK, true
}

func upper(s string) string { return strings.ToUpper(s) }

func (a *App) cmdConvert(args []string) int {
	fs, o := a.newFlagSet("convert")
	pos, code, ok := a.parse(fs, args)
	if !ok {
		return code
	}
	if len(pos) != 3 {
		return a.usageErr("convert needs <amount> <from> <to> (got %d)", len(pos))
	}
	amount, err := strconv.ParseFloat(pos[0], 64)
	if err != nil {
		return a.usageErr("invalid amount %q", pos[0])
	}
	from, to := upper(pos[1]), upper(pos[2])

	client, err := a.client(o)
	if err != nil {
		return a.fail(err)
	}
	ctx, cancel := o.ctx()
	defer cancel()

	result, err := client.Convert(ctx, amount, from, to)
	if err != nil {
		return a.fail(err)
	}
	if o.json {
		return a.emitJSON(map[string]any{"from": from, "to": to, "amount": amount, "result": result})
	}
	fmt.Fprintf(a.Stdout, "%s %s = %s %s\n", fmtNum(amount), from, fmtNum(result), to)
	return exitOK
}

func (a *App) cmdRate(args []string) int {
	fs, o := a.newFlagSet("rate")
	pos, code, ok := a.parse(fs, args)
	if !ok {
		return code
	}
	if len(pos) != 2 {
		return a.usageErr("rate needs <from> <to> (got %d)", len(pos))
	}
	from, to := upper(pos[0]), upper(pos[1])

	client, err := a.client(o)
	if err != nil {
		return a.fail(err)
	}
	ctx, cancel := o.ctx()
	defer cancel()

	rate, err := client.GetRate(ctx, from, to)
	if err != nil {
		return a.fail(err)
	}
	if o.json {
		return a.emitJSON(map[string]any{"from": from, "to": to, "rate": rate})
	}
	fmt.Fprintf(a.Stdout, "1 %s = %s %s\n", from, fmtNum(rate), to)
	return exitOK
}

func (a *App) cmdRates(args []string) int {
	fs, o := a.newFlagSet("rates")
	pos, code, ok := a.parse(fs, args)
	if !ok {
		return code
	}
	if len(pos) > 1 {
		return a.usageErr("rates takes at most one [base] argument (got %d)", len(pos))
	}
	base := "USD"
	if len(pos) == 1 {
		base = upper(pos[0])
	}

	client, err := a.client(o)
	if err != nil {
		return a.fail(err)
	}
	ctx, cancel := o.ctx()
	defer cancel()

	rates, err := client.GetAllRates(ctx, base)
	if err != nil {
		return a.fail(err)
	}
	if o.json {
		return a.emitJSON(map[string]any{"base": base, "rates": rates})
	}
	codes := sortedKeys(rates)
	width := maxLen(codes)
	for _, c := range codes {
		fmt.Fprintf(a.Stdout, "%-*s %s\n", width, c, fmtNum(rates[c]))
	}
	return exitOK
}

func (a *App) cmdCurrencies(args []string) int {
	fs, o := a.newFlagSet("currencies")
	pos, code, ok := a.parse(fs, args)
	if !ok {
		return code
	}
	if len(pos) != 0 {
		return a.usageErr("currencies takes no arguments (got %d)", len(pos))
	}

	client, err := a.client(o)
	if err != nil {
		return a.fail(err)
	}
	ctx, cancel := o.ctx()
	defer cancel()

	codes, err := client.GetSupportedCurrencies(ctx)
	if err != nil {
		return a.fail(err)
	}
	sort.Strings(codes)
	if o.json {
		return a.emitJSON(map[string]any{"currencies": codes})
	}
	for _, c := range codes {
		fmt.Fprintln(a.Stdout, c)
	}
	return exitOK
}

func (a *App) cmdVAT(args []string) int {
	fs, o := a.newFlagSet("vat")
	pos, code, ok := a.parse(fs, args)
	if !ok {
		return code
	}
	if len(pos) > 1 {
		return a.usageErr("vat takes at most one [country] argument (got %d)", len(pos))
	}

	client, err := a.client(o)
	if err != nil {
		return a.fail(err)
	}
	ctx, cancel := o.ctx()
	defer cancel()

	if len(pos) == 1 {
		resp, err := client.GetVATRate(ctx, upper(pos[0]))
		if err != nil {
			return a.fail(err)
		}
		if o.json {
			return a.emitJSON(resp)
		}
		v := resp.VATData
		fmt.Fprintf(a.Stdout, "%s (%s): %s%%\n", v.CountryCode, v.CountryName, fmtNum(v.VATRate))
		return exitOK
	}

	resp, err := client.GetVATRates(ctx)
	if err != nil {
		return a.fail(err)
	}
	if o.json {
		return a.emitJSON(resp)
	}
	codes := make([]string, 0, len(resp.VATRates))
	for c := range resp.VATRates {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	nameWidth := 0
	for _, c := range codes {
		if n := len(resp.VATRates[c].CountryName); n > nameWidth {
			nameWidth = n
		}
	}
	for _, c := range codes {
		v := resp.VATRates[c]
		fmt.Fprintf(a.Stdout, "%-2s  %-*s  %s%%\n", v.CountryCode, nameWidth, v.CountryName, fmtNum(v.VATRate))
	}
	return exitOK
}

func (a *App) cmdHistorical(args []string) int {
	fs, o := a.newFlagSet("historical")
	var amount float64
	fs.Float64Var(&amount, "amount", 1, "amount to convert (default 1 returns the unit rate)")
	pos, code, ok := a.parse(fs, args)
	if !ok {
		return code
	}
	if len(pos) != 3 {
		return a.usageErr("historical needs <date> <from> <to> (got %d) [date is YYYY-MM-DD]", len(pos))
	}
	date, from, to := pos[0], upper(pos[1]), upper(pos[2])

	client, err := a.client(o)
	if err != nil {
		return a.fail(err)
	}
	ctx, cancel := o.ctx()
	defer cancel()

	if amount == 1 {
		rate, err := client.GetHistoricalRate(ctx, date, from, to)
		if err != nil {
			return a.fail(err)
		}
		if o.json {
			return a.emitJSON(map[string]any{"date": date, "from": from, "to": to, "rate": rate})
		}
		fmt.Fprintf(a.Stdout, "1 %s = %s %s on %s\n", from, fmtNum(rate), to, date)
		return exitOK
	}

	result, err := client.ConvertHistorical(ctx, amount, from, to, date)
	if err != nil {
		return a.fail(err)
	}
	if o.json {
		return a.emitJSON(map[string]any{"date": date, "from": from, "to": to, "amount": amount, "result": result})
	}
	fmt.Fprintf(a.Stdout, "%s %s = %s %s on %s\n", fmtNum(amount), from, fmtNum(result), to, date)
	return exitOK
}

func (a *App) cmdTimeSeries(args []string) int {
	fs, o := a.newFlagSet("timeseries")
	var (
		base       string
		amount     float64
		currencies string
	)
	fs.StringVar(&base, "base", "USD", "base currency")
	fs.Float64Var(&amount, "amount", 1, "amount to convert")
	fs.StringVar(&currencies, "currencies", "", "comma-separated target currencies (default all)")
	pos, code, ok := a.parse(fs, args)
	if !ok {
		return code
	}
	if len(pos) != 2 {
		return a.usageErr("timeseries needs <start-date> <end-date> (got %d) [YYYY-MM-DD]", len(pos))
	}
	start, end := pos[0], pos[1]

	var codes []string
	for _, c := range strings.Split(currencies, ",") {
		if c = strings.TrimSpace(c); c != "" {
			codes = append(codes, upper(c))
		}
	}

	client, err := a.client(o)
	if err != nil {
		return a.fail(err)
	}
	ctx, cancel := o.ctx()
	defer cancel()

	series, err := client.GetTimeSeries(ctx, start, end, amount, upper(base), codes)
	if err != nil {
		return a.fail(err)
	}
	if o.json {
		return a.emitJSON(series)
	}
	dates := make([]string, 0, len(series.Data))
	for d := range series.Data {
		dates = append(dates, d)
	}
	sort.Strings(dates)
	for _, d := range dates {
		row := series.Data[d]
		cur := sortedKeys(row)
		parts := make([]string, len(cur))
		for i, c := range cur {
			parts[i] = c + " " + fmtNum(row[c])
		}
		fmt.Fprintf(a.Stdout, "%s  %s\n", d, strings.Join(parts, "  "))
	}
	return exitOK
}
