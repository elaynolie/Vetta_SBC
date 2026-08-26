package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

const (
	dbHost = "127.0.0.1"
	dbUser = "opensips"
	dbPass = "opensipsrw"
	dbName = "opensips"
)

func main() {
	group := flag.Int("g", 1, "address table group id")
	file := flag.String("f", "", "path to prefix file")
	noReload := flag.Bool("no-reload", false, "skip OpenSIPS address_reload")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 || *file == "" {
		fmt.Println("usage: geo_ip <COUNTRY> -f <file> [-g group] [--no-reload]")
		os.Exit(1)
	}
	country := strings.ToUpper(args[0])

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	if _, err := db.Exec("DELETE FROM address WHERE grp=$1 AND attrs=$2", *group, country); err != nil {
		log.Fatalf("delete old rows: %v", err)
	}

	f, err := os.Open(*file)
	if err != nil {
		log.Fatalf("open file: %v", err)
	}
	defer f.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}
	stmt, err := tx.Prepare(`INSERT INTO address (grp, ip, mask, port, proto, pattern, context_info, attrs)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`)
	if err != nil {
		log.Fatalf("prepare: %v", err)
	}
	defer stmt.Close()

	scanner := bufio.NewScanner(f)
	loaded, skipped := 0, 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var ip string
		var mask int
		var cc string

		fields := strings.Split(line, ",")

		switch {
		case len(fields) == 1 && strings.Contains(fields[0], "/"):
			// plain CIDR: 109.197.192.0/21
			parts := strings.Split(fields[0], "/")
			if len(parts) != 2 {
				skipped++
				continue
			}
			m, err := strconv.Atoi(parts[1])
			if err != nil {
				skipped++
				continue
			}
			ip, mask, cc = parts[0], m, country

		case len(fields) >= 3:
			// ip,mask,country
			m, err := strconv.Atoi(strings.TrimSpace(fields[1]))
			if err != nil {
				skipped++
				continue
			}
			fcc := strings.ToUpper(strings.TrimSpace(fields[2]))
			if fcc != country {
				skipped++
				continue
			}
			ip, mask, cc = strings.TrimSpace(fields[0]), m, fcc

		default:
			skipped++
			continue
		}

		if _, err := stmt.Exec(*group, ip, mask, 5060, "any", "", "", cc); err != nil {
			log.Printf("insert failed for %s/%d: %v", ip, mask, err)
			skipped++
			continue
		}
		loaded++
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("scan error: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("commit: %v", err)
	}

	fmt.Printf("[%s] Loaded %d prefixes, skipped %d\n", country, loaded, skipped)

	if !*noReload {
		out, err := exec.Command("opensips-cli", "-x", "mi", "address_reload").CombinedOutput()
		if err != nil {
			log.Printf("address_reload failed: %v\n%s", err, out)
		} else {
			fmt.Println(string(out))
		}
	}
}