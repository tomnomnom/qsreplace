package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
)

func main() {
	var appendMode bool
	var uniqueMode bool
	flag.BoolVar(&appendMode, "a", false, "Append the value instead of replacing it")
	flag.BoolVar(&uniqueMode, "u", false, "Uniquely modify one parameter at a time")
	flag.Parse()

	seen := make(map[string]bool)

	// read URLs on stdin, then replace the values in the query string
	// with some user-provided value
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		u, err := url.Parse(sc.Text())
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse url %s [%s]\n", sc.Text(), err)
			continue
		}

		// Go's maps aren't ordered, but we want to use all the param names
		// as part of the key to output only unique requests. To do that, put
		// them into a slice and then sort it.
		pp := make([]string, 0)
		for p := range u.Query() {
			pp = append(pp, p)
		}
		sort.Strings(pp)

		key := fmt.Sprintf("%s%s?%s", u.Hostname(), u.EscapedPath(), strings.Join(pp, "&"))

		// Only output each host + path + params combination once
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = true

		if uniqueMode {
			old_qs := u.Query()
			for param := range u.Query() {
				qs := url.Values{}
				if appendMode {
					qs.Set(param, old_qs.Get(param)+flag.Arg(0))
				} else {
					qs.Set(param, flag.Arg(0))
				}
				for p := range u.Query() {
					if p != param {
						qs.Set(p, old_qs.Get(p))
					}
				}

				u.RawQuery = qs.Encode()

				fmt.Printf("%s\n", u)
			}
		} else {
			qs := url.Values{}
			for param, vv := range u.Query() {
				if appendMode {
					qs.Set(param, vv[0]+flag.Arg(0))
				} else {
					qs.Set(param, flag.Arg(0))
				}
			}

			u.RawQuery = qs.Encode()

			fmt.Printf("%s\n", u)
		}

	}

}
