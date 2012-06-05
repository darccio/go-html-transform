package main

import (
	"code.google.com/p/go-html-transform/h5"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)


func runDatTests(ps []string) {
	//for _, p := range ps {
	//}
}

func runTestTests(ps []string) {
	//for _, p := range ps {
	//}
}

func runHtmlTests(ps []string) {
	for _, p := range ps {
		fmt.Println("Attempting to parse file: ", p)
		f, err := os.Open(p)
		if err != nil {
			fmt.Println("Error opening file: ", err)
		}
		parse := h5.NewParser(f)
		err = parse.Parse()
		if err != nil {
			fmt.Println("Error parsing file: ", err)
		} else {
			fmt.Println("Success!!!")
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
		fmt.Println("Error while grepping: ", err)
	}
	runDatTests(spec[datRe])
	runTestTests(spec[testRe])
	runHtmlTests(spec[htmlRe])
}