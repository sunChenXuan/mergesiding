// Package cli implements the mergesiding command-line interface.
package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sunChenXuan/mergesiding/internal/abort"
	"github.com/sunChenXuan/mergesiding/internal/cleanup"
	"github.com/sunChenXuan/mergesiding/internal/integrate"
	mcpserver "github.com/sunChenXuan/mergesiding/internal/mcp"
	"github.com/sunChenXuan/mergesiding/internal/paths"
	"github.com/sunChenXuan/mergesiding/internal/ready"
	"github.com/sunChenXuan/mergesiding/internal/start"
	"github.com/sunChenXuan/mergesiding/internal/status"
)

// Execute parses argv and runs the requested subcommand. Returns process exit code.
func Execute(argv []string) int {
	if len(argv) < 1 {
		printHelp()
		return 2
	}
	cmd := argv[0]
	args := argv[1:]
	p := paths.Default()
	switch cmd {
	case "start":
		return cmdStart(p, args)
	case "ready":
		return cmdReady(p, args)
	case "integrate":
		return cmdIntegrate(p, args)
	case "status":
		slug := ""
		asJSON := false
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--slug":
				i++
				if i < len(args) {
					slug = args[i]
				}
			case "--json":
				asJSON = true
			}
		}
		return cmdStatus(p, slug, asJSON)
	case "abort":
		return cmdAbort(p, args)
	case "cleanup":
		return cmdCleanup(p, args)
	case "mcp":
		if err := mcpserver.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	case "help", "-h", "--help":
		printHelp()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printHelp()
		return 2
	}
}

func printHelp() {
	fmt.Println(`mergesiding — parallel agent worktrees with serial integrate

Usage:
  mergesiding start --slug <slug> --repo <abs> [--worktree-root <path>] [--writer-id <id>] [--brief <path>]
  mergesiding ready --slug <slug>
  mergesiding integrate [--all | --slug <slug>] [--json]
  mergesiding status [--slug <slug>] [--json]
  mergesiding abort --slug <slug>
  mergesiding cleanup --slug <slug> (--remove-worktree and/or --delete-branch)
  mergesiding mcp   # MERGESIDING_MCP_ROLE=writer|scheduler`)
}

func cmdStart(p paths.Paths, args []string) int {
	var slug, writerID, brief, shared string
	var repos []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--slug":
			i++
			if i < len(args) {
				slug = args[i]
			}
		case "--repo":
			i++
			if i < len(args) {
				repos = append(repos, args[i])
			}
		case "--worktree-root":
			i++
			if i < len(args) {
				shared = args[i]
			}
		case "--writer-id":
			i++
			if i < len(args) {
				writerID = args[i]
			}
		case "--brief":
			i++
			if i < len(args) {
				brief = args[i]
			}
		}
	}
	opt := start.Options{Slug: slug, Repos: repos}
	if shared != "" {
		opt.SharedRoot = &shared
	}
	if writerID != "" {
		opt.WriterID = &writerID
	}
	if brief != "" {
		opt.BriefPath = &brief
	}
	task, err := start.Start(p, opt)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, rb := range task.Repos {
		fmt.Println(rb.WorktreePath)
	}
	return 0
}

func cmdReady(p paths.Paths, args []string) int {
	slug := flagValue(args, "--slug")
	if slug == "" {
		fmt.Fprintln(os.Stderr, "--slug is required")
		return 2
	}
	if _, err := ready.MarkReady(p, slug); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("ready:", slug)
	return 0
}

func cmdIntegrate(p paths.Paths, args []string) int {
	all := hasFlag(args, "--all")
	asJSON := hasFlag(args, "--json")
	slugArg := flagValue(args, "--slug")
	if all && slugArg != "" {
		fmt.Fprintln(os.Stderr, "--all and --slug are mutually exclusive")
		return 2
	}
	if all {
		results, summary := integrate.IntegrateAll(p)
		failed := false
		for _, out := range results {
			if out.Failed() {
				failed = true
			}
		}
		if asJSON {
			if err := json.NewEncoder(os.Stdout).Encode(map[string]any{
				"results": mapOutcomes(results),
				"summary": summary,
			}); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
		} else {
			for _, out := range results {
				printIntegrate(out)
			}
		}
		if failed {
			return 1
		}
		return 0
	}
	var slugPtr *string
	if slugArg != "" {
		slugPtr = &slugArg
	}
	out := integrate.IntegrateOne(p, slugPtr)
	if asJSON {
		if err := json.NewEncoder(os.Stdout).Encode(out.ToDict()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	} else if !out.Empty {
		printIntegrate(out)
	}
	if out.Failed() {
		return 1
	}
	return 0
}

func mapOutcomes(results []integrate.Outcome) []map[string]any {
	out := make([]map[string]any, 0, len(results))
	for _, r := range results {
		out = append(out, r.ToDict())
	}
	return out
}

func printIntegrate(out integrate.Outcome) {
	slug := out.Slug
	if slug == "" {
		slug = "-"
	}
	st := "-"
	if out.FinalStatus != nil {
		st = string(*out.FinalStatus)
	} else if out.Error != nil {
		st = *out.Error
	} else if out.Message != "" {
		st = out.Message
	}
	line := fmt.Sprintf("integrate: %s -> %s", slug, st)
	if out.Error != nil && out.FinalStatus != nil {
		line += " " + *out.Error
	}
	fmt.Println(line)
}

func cmdStatus(p paths.Paths, slug string, asJSON bool) int {
	var slugPtr *string
	if slug != "" {
		slugPtr = &slug
	}
	if asJSON {
		payload, err := status.Payload(p, slugPtr)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(payload)
		return 0
	}
	lines, err := status.FormatLines(p, slugPtr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, ln := range lines {
		fmt.Println(ln)
	}
	return 0
}

func cmdAbort(p paths.Paths, args []string) int {
	if hasFlag(args, "--remove-worktree") || hasFlag(args, "--delete-branch") {
		fmt.Fprintln(os.Stderr, "abort no longer removes git artifacts; use: mergesiding cleanup --slug <slug> --remove-worktree and/or --delete-branch")
		return 2
	}
	slug := flagValue(args, "--slug")
	if slug == "" {
		fmt.Fprintln(os.Stderr, "--slug is required")
		return 2
	}
	task, err := abort.Abort(p, slug)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("aborted:", task.Slug)
	return 0
}

func cmdCleanup(p paths.Paths, args []string) int {
	slug := flagValue(args, "--slug")
	if slug == "" {
		fmt.Fprintln(os.Stderr, "--slug is required")
		return 2
	}
	actions, err := cleanup.Run(p, slug, hasFlag(args, "--remove-worktree"), hasFlag(args, "--delete-branch"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, a := range actions {
		fmt.Println(a)
	}
	return 0
}

func flagValue(args []string, name string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == name && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func hasFlag(args []string, name string) bool {
	for _, a := range args {
		if a == name {
			return true
		}
	}
	return false
}
