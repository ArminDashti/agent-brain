package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ArminDashti/agent-brain/dispatcher-win/internal/api"
	"github.com/ArminDashti/agent-brain/dispatcher-win/internal/extract"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "ingest":
		os.Exit(runIngest(os.Args[2:]))
	case "sync":
		os.Exit(runSync(os.Args[2:]))
	case "help", "-h", "--help":
		printUsage()
	default:
		fail(fmt.Sprintf("unknown command %q", os.Args[1]))
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `agent-brain-dispatcher — extract Cursor sessions from state.vscdb and ingest via agent-brain-api

Usage:
  agent-brain-dispatcher ingest --uuid <uuid> [flags]
  agent-brain-dispatcher sync [flags]

Flags:
  --uuid      Session/composer UUID (required for ingest)
  --db        Path to state.vscdb (default: %%APPDATA%%\Cursor\User\globalStorage\state.vscdb)
  --api-url   agent-brain-api base URL (default: http://127.0.0.1:8220)
  --out       Optional path to write extract JSON (ingest only)
  --username  API username (env AGENT_BRAIN_USERNAME)
  --password  API password (env AGENT_BRAIN_PASSWORD)

Env:
  AGENT_BRAIN_API_URL, AGENT_BRAIN_USERNAME, AGENT_BRAIN_PASSWORD, AGENT_BRAIN_STATE_VSCDB
`)
}

func runIngest(args []string) int {
	fs := flag.NewFlagSet("ingest", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	uuid := fs.String("uuid", "", "session UUID")
	dbPath := fs.String("db", "", "state.vscdb path")
	apiURL := fs.String("api-url", "", "API base URL")
	outPath := fs.String("out", "", "optional extract JSON output path")
	username := fs.String("username", "", "API username")
	password := fs.String("password", "", "API password")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	u := strings.TrimSpace(*uuid)
	if u == "" {
		return fail("uuid is required")
	}

	db := resolveDB(*dbPath)
	base := firstNonEmpty(*apiURL, os.Getenv("AGENT_BRAIN_API_URL"), "http://127.0.0.1:8220")
	user := firstNonEmpty(*username, os.Getenv("AGENT_BRAIN_USERNAME"), "armin")
	pass := firstNonEmpty(*password, os.Getenv("AGENT_BRAIN_PASSWORD"), "dopadopa123")

	raw, err := extract.Session(u, db)
	if err != nil {
		return fail(err.Error())
	}

	if *outPath != "" {
		if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
			return fail(err.Error())
		}
		if err := os.WriteFile(*outPath, raw, 0o644); err != nil {
			return fail(err.Error())
		}
	}

	client := api.New(base, user, pass)
	res, err := client.IngestSession(raw)
	if err != nil {
		return fail(err.Error())
	}

	out := map[string]any{
		"ok":            true,
		"uuid":          res.UUID,
		"turn_count":    res.TurnCount,
		"context_pct":   res.ContextPct,
		"input_tokens":  res.InputTokens,
		"output_tokens": res.OutputTokens,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(out)
	return 0
}

func runSync(args []string) int {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dbPath := fs.String("db", "", "state.vscdb path")
	apiURL := fs.String("api-url", "", "API base URL")
	username := fs.String("username", "", "API username")
	password := fs.String("password", "", "API password")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	db := resolveDB(*dbPath)
	if db == "" {
		return fail("state.vscdb path not found (set --db or APPDATA)")
	}
	base := firstNonEmpty(*apiURL, os.Getenv("AGENT_BRAIN_API_URL"), "http://127.0.0.1:8220")
	user := firstNonEmpty(*username, os.Getenv("AGENT_BRAIN_USERNAME"), "armin")
	pass := firstNonEmpty(*password, os.Getenv("AGENT_BRAIN_PASSWORD"), "dopadopa123")

	composers, err := extract.ListComposers(db)
	if err != nil {
		return fail(err.Error())
	}

	client := api.New(base, user, pass)
	if err := client.Login(); err != nil {
		return fail(err.Error())
	}

	type item struct {
		UUID       string   `json:"uuid"`
		Name       string   `json:"name,omitempty"`
		OK         bool     `json:"ok"`
		TurnCount  int      `json:"turn_count,omitempty"`
		ContextPct *float64 `json:"context_pct,omitempty"`
		Error      string   `json:"error,omitempty"`
	}

	results := make([]item, 0, len(composers))
	synced := 0
	failed := 0

	for _, c := range composers {
		entry := item{UUID: c.UUID, Name: c.Name}
		raw, err := extract.Session(c.UUID, db)
		if err != nil {
			entry.Error = err.Error()
			failed++
			results = append(results, entry)
			continue
		}
		res, err := client.IngestSession(raw)
		if err != nil {
			entry.Error = err.Error()
			failed++
			results = append(results, entry)
			continue
		}
		entry.OK = true
		entry.TurnCount = res.TurnCount
		entry.ContextPct = res.ContextPct
		synced++
		results = append(results, entry)
	}

	out := map[string]any{
		"ok":      failed == 0,
		"synced":  synced,
		"failed":  failed,
		"total":   len(composers),
		"results": results,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(out)
	if failed > 0 {
		return 1
	}
	return 0
}

func resolveDB(flagVal string) string {
	return firstNonEmpty(flagVal, os.Getenv("AGENT_BRAIN_STATE_VSCDB"), defaultDBPath())
}

func defaultDBPath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return ""
	}
	return filepath.Join(appData, "Cursor", "User", "globalStorage", "state.vscdb")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func fail(msg string) int {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(map[string]any{"ok": false, "error": msg})
	return 1
}
