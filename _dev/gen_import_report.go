// TamaGo import report
// Based on https://github.com/trashhalo/tinygo-import-report.git

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"text/template"
)

const mainTemplate = `
package main

import (
	_ "unsafe"
	_ "{{.}}"
)

//go:linkname getRandomData runtime.getRandomData
func getRandomData(b []byte) {
}

//go:linkname printk runtime.printk
func printk(byte) {
	return
}

//go:linkname hwinit runtime.hwinit
func hwinit() {
	return
}

//go:linkname initRNG runtime.initRNG
func initRNG() {
	return
}

//go:linkname nanotime1 runtime.nanotime1
func nanotime1() (ns int64) {
	return
}

//go:linkname stackOffset runtime.stackOffset
var stackOffset uint32

func main() {
}
`

const readmeTemplate = `
# Tamago Import Report
This project imports each package in the stdlib and reports if it imports cleanly in tamago.
A package with a check may work a package with a x will definately not currently work.

| Package | Imported? | 
| --- | --- |{{ range $key, $value := .}}
| {{$value.Name}} | {{if $value.Imported}} ok {{else}} [failed](#{{$value.Link}}) {{end}} | {{ end }}


{{ range $key, $value := .}}
## {{$value.Name}}

BTBTBT
{{$value.Output}}
BTBTBT

{{ end }}
`

type Result struct {
	Name     string
	Imported bool
	Output   string
	Link     string
}

func main() {
	tamagoGo := os.Getenv("TAMAGO")

	if tamagoGo == "" {
		panic("You need to set the TAMAGO variable to a compiled version of https://github.com/f-secure-foundry/tamago-go")
	}

	content, err := ioutil.ReadFile("imports")

	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(content), "\n")
	maint := template.Must(template.New("main").Parse(mainTemplate))
	readmet := template.Must(template.New("readme").Parse(strings.Replace(readmeTemplate, "BTBTBT", "```", -1)))
	results := make(chan Result)

	wg := &sync.WaitGroup{}
	resultsArr := []Result{}

	sem := make(chan struct{}, 4)

	go func() {
		for r := range results {
			resultsArr = append(resultsArr, r)
			println(r.Name)

			wg.Done()
		}
	}()

	for _, line := range lines {
		wg.Add(1)

		go func(line string) {
			sem <- struct{}{}
			defer func() {
				<-sem
			}()

			noslash := strings.Replace(line, "/", "_", -1)
			dirsafe := fmt.Sprintf("tests/%v", noslash)
			os.Mkdir(dirsafe, 0755)
			f, err := os.Create(fmt.Sprintf("%v/main.go", dirsafe))

			if err != nil {
				panic(err)
			}
			defer f.Close()

			err = maint.Execute(f, line)

			if err != nil {
				panic(err)
			}

			cmd := exec.Command(tamagoGo, "build", fmt.Sprintf("tests/%v/main.go", noslash))
			stdoutStderr, err := cmd.CombinedOutput()

			results <- Result{
				line,
				err == nil,
				string(stdoutStderr),
				strings.Replace(line, "/", "", -1),
			}
		}(line)
	}

	wg.Wait()
	close(results)

	sort.Slice(resultsArr, func(i, j int) bool {
		return resultsArr[i].Name < resultsArr[j].Name
	})

	f, err := os.Create("import_report.md")

	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = readmet.Execute(f, resultsArr)

	if err != nil {
		panic(err)
	}
}
