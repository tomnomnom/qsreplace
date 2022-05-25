package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
)

func addParam(u *url.URL, value string, appendMode bool) {
	seen := make(map[string]bool)

	// Go's maps aren't ordered, but we want to use all the param names
	// as part of the key to output only unique requests. To do that, put
	// them into a slice and then sort it.
	pp := make([]string, 0)
	for p, _ := range u.Query() {
		pp = append(pp, p)
	}
	sort.Strings(pp)

	key := fmt.Sprintf("%s%s?%s", u.Hostname(), u.EscapedPath(), strings.Join(pp, "&"))

	// Only output each host + path + params combination once
	if _, exists := seen[key]; exists {
		return
	}
	seen[key] = true

	qs := url.Values{}
	for param, vv := range u.Query() {
		if appendMode {
			qs.Set(param, vv[0]+value)
		} else {
			qs.Set(param, value)
		}
	}
	u.RawQuery = qs.Encode()
	fmt.Printf("%s\n", u)
}

func main() {
	var appendMode bool
	var bothMode bool
	var wordlist string
	flag.BoolVar(&appendMode, "a", false, "Append the value instead of replacing it")
	flag.BoolVar(&bothMode, "b", false, "Replace the value once and append it once for each url")
	flag.StringVar(&wordlist, "w", "", "Wordlist to use")
	flag.Parse()

	var data []byte

	if wordlist != "" {
		var err error
		data, err = os.ReadFile(wordlist)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read wordlist %s [%s]\n", wordlist, err)
			os.Exit(1)
		}
	}

	// read URLs on stdin, then replace the values in the query string
	// with some user-provided value
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		u, err := url.Parse(sc.Text())
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse url %s [%s]\n", sc.Text(), err)
			continue
		}

		if len(u.Query()) == 0 {
			// skip URLs with no query string
			continue
		}

		if data != nil {
			split := bytes.Split(bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n")), []byte("\n"))
			for _, v := range split {
				if bothMode {
					addParam(u, string(v), true)
					addParam(u, string(v), false)
					continue
				}
				addParam(u, string(v), appendMode)
			}
		} else {
			if bothMode {
				addParam(u, flag.Arg(0), true)
				addParam(u, flag.Arg(0), false)
			} else {
				addParam(u, flag.Arg(0), appendMode)
			}
		}
	}

}
