package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"memento/internal/indexer"
	"memento/internal/searcher"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "index":
		runIndex(os.Args[2:])
	case "search":
		runSearch(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprint(os.Stderr, `Usage:
  memento index [--db .memento/wiki.db] <wiki-dir>
  memento search [--db .memento/wiki.db] [--json] [--limit 10] <query>
`)
}

func runIndex(args []string) {
	fs := flag.NewFlagSet("index", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: memento index [--db .memento/wiki.db] <wiki-dir>\n")
	}
	dbPath := fs.String("db", ".memento/wiki.db", "database path")

	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: wiki directory required")
		fs.Usage()
		os.Exit(1)
	}
	wikiDir := fs.Arg(0)

	fmt.Printf("Indexing %s into %s ...\n", wikiDir, *dbPath)
	if err := indexer.Index(wikiDir, *dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("Indexing complete.")
}

func runSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: memento search [--db .memento/wiki.db] [--json] [--limit 10] <query>\n")
	}
	dbPath := fs.String("db", ".memento/wiki.db", "database path")
	jsonOutput := fs.Bool("json", false, "output results as JSON")
	limit := fs.Int("limit", 10, "max number of results")

	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: search query required")
		fs.Usage()
		os.Exit(1)
	}
	query := strings.Join(fs.Args(), " ")

	s, err := searcher.Open(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	defer s.Close()

	if *jsonOutput {
		if err := s.SearchJSON(query, *limit); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	} else {
		if err := s.SearchText(query, *limit); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	}
}
