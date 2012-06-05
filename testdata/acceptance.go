package main

import (
	"bytes"
	"code.google.com/p/go-html-transform/h5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)


func runDatTests(ps []string) {
	for _, p := range ps {
		fmt.Println("Running tests in file: ", p)
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
			runDatCase(c)
		}
	}
}

func runDatCase(c []byte) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("ERROR while running test case:", e)
		}
	}()
	parts := bytes.Split(c, []byte("#"))
	if len(parts) != 4 {
		fmt.Printf("Malformed test case: %d, %q\n", len(parts), string(c))
		return
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
		fmt.Println("ERROR parsing file: ", err)
	} else {
		fmt.Println("SUCCESS!!!")
	}
}

func runTestTests(ps []string) {
	//for _, p := range ps {
	//}
}

func runHtmlTests(ps []string) {
	// TODO(jwall): with timings?
	for _, p := range ps {
		fmt.Println("Attempting to parse file: ", p)
		f, err := os.Open(p)
		if err != nil {
			fmt.Println("ERROR opening file: ", err)
		}
		parse := h5.NewParser(f)
		err = parse.Parse()
		if err != nil {
			fmt.Println("ERROR parsing file: ", err)
		} else {
			fmt.Println("SUCCESS!!!")
		}
	}
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

func main() {
	datRe := regexp.MustCompile("dat$")
	testRe := regexp.MustCompile("test$")
	htmlRe := regexp.MustCompile("html?$")
	spec := make(map[*regexp.Regexp][]string)
	spec[datRe] = []string{}
	spec[testRe] = []string{}
	spec[htmlRe] = []string{}
	err := grep("./", spec)
	if err != nil {
		fmt.Println("ERROR while grepping", err)
	}
	runDatTests(spec[datRe])
	runTestTests(spec[testRe])
	runHtmlTests(spec[htmlRe])
}