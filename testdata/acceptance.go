package main

import (
	"bytes"
	"code.google.com/p/go-html-transform/h5"
	"fmt"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

var (
	verbose = flag.Bool("verbose", false, "Verbosity for test output")
	testSpec = StringEnum("test_spec", map[string]struct{}{"dat": struct{}{},
		"file":struct{}{},
		"all": struct{}{},
	}, "all", "Type of test to run")
)

func runDatTests(ps []string) int {
	var counter int
	for _, p := range ps {
		if *verbose { fmt.Println("Running tests in file: ", p); }
		f, err := os.Open(p)
		if err != nil {
			fmt.Println("ERROR opening file: ", err)
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println("ERROR reading file: ", err)
		}
		cases := bytes.Split(data, []byte("\n\n"))
		for _, c := range cases {
			counter += runDatCase(c)
		}
	}
	return counter
}

func runDatCase(c []byte) int {
	var counter int
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("ERROR while running test case:", e)
			counter++
		}
	}()
	parts := bytes.Split(c, []byte("#"))
	if len(parts) != 4 { counter++; }
	if len(parts) != 4 && *verbose {
		fmt.Printf("Malformed test case: %d, %q\n", len(parts), string(c))
		return counter
	}
	fmt.Println("Running test case:", string(c))
	testData := make(map[string]string)
	for _, p := range parts[1:] {
		t := bytes.Split(p, []byte("\n"))
		testData[string(t[0])] = string(t[1])
	}
	p := h5.NewParserFromString(string(testData["data"]))
	err := p.Parse()
	if err != nil {
		fmt.Println("Test case:", string(c))
		fmt.Println("ERROR parsing: ", err)
		counter++
	} else {
		if *verbose { fmt.Println("SUCCESS!!!") }
	}
	return counter
}

func runTestTests(ps []string) int {
	var counter int
	//for _, p := range ps {
	//}
	return counter
}

func runHtmlTests(ps []string) int {
	var counter int
	// TODO(jwall): with timings?
	for _, p := range ps {
		if *verbose { fmt.Println("Attempting to parse file: ", p); }
		f, err := os.Open(p)
		if err != nil {
			fmt.Println("ERROR opening file: ", err)
			counter++
		}
		parse := h5.NewParser(f)
		err = parse.Parse()
		if err != nil {
			if !*verbose { fmt.Println("Attempting to parse file: ", p); }
			fmt.Println("ERROR parsing file: ", err)
			counter++
		} else {
			if *verbose { fmt.Println("SUCCESS!!!") }
		}
	}
	return counter
}

type grepSpec map[*regexp.Regexp][]string

func grep(path string, spec grepSpec) error {
	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			for re, _ := range spec {
				if re.MatchString(p) {
					spec[re] = append(spec[re], p)
				}
			}
		}
		return nil
	})
}

// TODO(jwall): failure counter
// TODO(jwall): Output overall success
func main() {
	flag.Parse()
	datRe := regexp.MustCompile("dat$")
	testRe := regexp.MustCompile("test$")
	htmlRe := regexp.MustCompile("html?$")
	spec := make(map[*regexp.Regexp][]string)
	spec[datRe] = []string{}
	spec[testRe] = []string{}
	spec[htmlRe] = []string{}
	var counter int
	err := grep("./", spec)
	if err != nil {
		fmt.Println("ERROR while grepping", err)
	}
	specType := testSpec.String()
	if specType == "all" || specType == "dat" {
		counter += runDatTests(spec[datRe])
	}
	if specType == "all" || specType == "test" {
		counter += runTestTests(spec[testRe])
	}
	if specType == "all" || specType == "dat" {
		counter += runHtmlTests(spec[htmlRe])
	}
	if counter > 0 {
		fmt.Printf("%d Errors found\n", counter)
	} else {
		fmt.Println("All Tests passed!!")
	}
}