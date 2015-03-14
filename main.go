package main

import (
	"flag"
	"fmt"
	"github.com/bndr/gojenkins"
	"os"
	"regexp"
)

func main() {
	var url, name, regex string

	flag.StringVar(&url, "jenkins", "", "Your jenkins url")
	flag.StringVar(&name, "job", "", "Your job name")
	flag.StringVar(&regex, "regex", "", "Regex for job name")
	flag.Parse()

	if url == "" {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("Using jenkins URL %s\n", url)
	jenkins := gojenkins.CreateJenkins(url).Init()

	if name == "" && regex == "" {
		flag.Usage()
		os.Exit(1)
	} else if regex != "" {
		var names []string
		var nameMatch = regexp.MustCompile(regex)
		for _, job := range jenkins.GetAllJobNames() {
			if nameMatch.MatchString(job.Name) {
				names = append(names, job.Name)
			}
		}

		if len(names) == 0 {
			fmt.Printf("No matching job name found\n")
			os.Exit(1)
		} else if len(names) > 1 {
			fmt.Printf("More than one job name found: %v\n", names)
			os.Exit(1)
		}

		// use the matched name
		name = names[0]
	}

	job := jenkins.GetJob(name)
	if job == nil {
		fmt.Printf("No job found\n")
		os.Exit(1)
	}
	fmt.Printf("Found job %s\n", job.GetName())
	fmt.Printf("Next build number: %d\n", job.GetDetails().NextBuildNumber)

	var paramNames []string
	var buildParams flag.FlagSet
	for _, property := range job.GetDetails().Property {
		for _, param := range property.ParameterDefinitions {
			//fmt.Printf("%s\n", param.Name)
			var value string
			buildParams.StringVar(&value, param.Name, "", "")
			paramNames = append(paramNames, param.Name)
		}
	}

	params := make(map[string]string)
	if len(paramNames) > 0 {
		err := buildParams.Parse(args())
		if err != nil {
			fmt.Printf("Couldn't parse the build arguments: %s\n", err.Error())
			os.Exit(1)
		}

		for _, name := range paramNames {
			f := buildParams.Lookup(name)
			fmt.Printf("%s: %s\n", f.Name, f.Value)
			params[f.Name] = f.Value.String()
		}
	}

	fmt.Printf("Triggering build with params: %+v\n", params)
	job.InvokeSimple(params)
}

// get list of command-line arguments pre-prended with "-"
// so we can parse them using flag.Parse()
func args() []string {
	args := flag.Args()
	var nArgs []string
	for _, s := range args {
		nArgs = append(nArgs, "-"+s)
	}
	return nArgs
}
