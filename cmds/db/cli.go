package main

import (
	"context"
	"database/sql"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	usage := "\tUsage: migrate"
	if len(os.Args) < 2 {
		fmt.Println(usage)
		return
	}

	subCmd := flag.NewFlagSet(os.Args[1], flag.ExitOnError)

	switch os.Args[1] {
	case "migrate":
		dsn := subCmd.String("dsn", "", "usage: -dsn=<sqlite dsn>")
		subCmd.Parse(os.Args[2:])
		migrate(*dsn)
	default:
		fmt.Println(usage)
	}
}

func migrate(dsn string) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		fmt.Println("DSN is required")
		return
	}
	ddl, err := os.ReadFile("schema.sql")

	sqlDb, err := sql.Open("sqlite3", dsn)
	if err != nil {
		fmt.Printf("Error opening database: %v", err)
		return
	}
	defer close(sqlDb, "database")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := sqlDb.ExecContext(ctx, string(ddl)); err != nil {
		fmt.Printf("Error creating tables: %v", err)
		return
	}
}

type Closer interface {
	Close() error
}

func close(c Closer, name string) {
	if err := c.Close(); err != nil {
		fmt.Printf("Error closing %s: %v", name, err)
	}
}
