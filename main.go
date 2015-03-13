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
}
