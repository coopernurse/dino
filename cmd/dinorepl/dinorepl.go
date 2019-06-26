package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/coopernurse/dino"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("Welcome to dino - type help for help screen")
	repl()
}

func repl() {
	c := &context{}
	for {
		fmt.Println(eval(read(), c))
	}
}

func read() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("dino => ")
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func eval(s string, c *context) string {
	if s == "exit" {
		os.Exit(0)
	} else if s == "help" {
		return help
	} else if strings.HasPrefix(s, "namedotcom ") {
		tok := strings.Split(s, " ")
		if len(tok) == 3 {
			c.provider = dino.NewNameDotComProvider(tok[1], tok[2])
			return fmt.Sprintf("Set provider to namedotcom with username: %s", tok[1])
		} else {
			return "Invalid usage. Expected: namedotcom username token"
		}
	} else if strings.HasPrefix(s, "list ") {
		if c.provider == nil {
			return "No provider set yet"
		} else {
			return listDomain(s, c.provider)
		}
	} else if strings.HasPrefix(s, "put ") {
		if c.provider == nil {
			return "No provider set yet"
		} else {
			tok, errmsg := splitAndCheck(s, 6, "put <domain> <host> <type> <answer> <ttl>")
			if errmsg != "" {
				return errmsg
			} else {
				rec := dino.Record{
					Domain: tok[1],
					Host:   tok[2],
					Type:   dino.RecordType(strings.ToUpper(tok[3])),
					Answer: tok[4],
					Ttl:    uint32(parseInt(tok[5], 300)),
				}
				err := c.provider.Put(rec)
				if err == nil {
					return "Put OK"
				} else {
					return fmt.Sprintf("Error in Put: %s", err.Error())
				}
			}
		}
	} else if strings.HasPrefix(s, "delete ") {
		if c.provider == nil {
			return "No provider set yet"
		} else {
			tok, errmsg := splitAndCheck(s, 3, "delete <domain> <id>")
			if errmsg != "" {
				return errmsg
			} else {
				err := c.provider.Delete(tok[1], tok[2])
				if err == nil {
					return "Delete OK"
				} else {
					return fmt.Sprintf("Error in Delete: %s", err.Error())
				}
			}
		}
	}
	return fmt.Sprintf("Unknown command: %s", s)
}

func listDomain(cmd string, provider dino.Provider) string {
	tok, errmsg := splitAndCheck(cmd, 2, "list <domain>")
	if errmsg != "" {
		return errmsg
	}
	records, err := provider.List(tok[1])
	if err == nil {
		b := bytes.NewBufferString("ID           Type   TTL   Host                           Answer\n")
		for _, r := range records {
			b.WriteString(fmt.Sprintf("%-12s %-6s %-5d %-30s %s\n", r.Id, r.Type, r.Ttl, r.Host, r.Answer))
		}
		return b.String()
	} else {
		return fmt.Sprintf("Error in List: %s", err.Error())
	}
}

func splitAndCheck(cmd string, tokens int, usage string) ([]string, string) {
	toks := strings.Split(cmd, " ")
	if len(toks) == tokens {
		return toks, ""
	} else {
		return nil, fmt.Sprintf("Invalid command - Usage: %s", usage)
	}
}

func parseInt(s string, defaultVal int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("parseInt: unable to parse: %s - %v\n", s, err)
		v = defaultVal
	}
	return v
}

type context struct {
	provider dino.Provider
}

const help = `Commands:
namedotcom <username> <token>: Login to name.com

list <domain>: List records for domain
put <domain> <host> <type> <answer> <ttl>: Put dns record (uses host/type as key)
delete <domain> <id>: Delete dns record

help: print help
exit: exit dino
`
