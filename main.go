package main

import (
	"flag"
	"fmt"
	"github.com/leibowitz/gojenkins"
	"os"
	"regexp"
	"time"
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
	if jenkins == nil {
		fmt.Printf("Unable to connect to: %s", url)
		os.Exit(1)
	}

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

	status := "FAILED"

	rsp, err := job.Build(params)
	if err == nil && job.Successful(rsp) {
		status = "OK"
	}

	fmt.Printf("%s\n", status)

	if !job.Successful(rsp) {
		os.Exit(1)
	}

	fmt.Printf("%s\n", rsp.Header.Get("Location"))

	// Wait for a maximum of 2 minutes
	maxWait := 2 * time.Minute

	start := time.Now()

	var build *gojenkins.Build
	// Wait for build to exist
	for {
		//fmt.Printf("1.Waiting for build: %d\n", job.GetDetails().NextBuildNumber)
		build = job.GetBuild(job.GetDetails().NextBuildNumber)
		if build != nil {
			fmt.Printf("\nBuild found:\n%s\n", build.GetUrl())
			break
		}
		if time.Now().Sub(start) > maxWait {
			fmt.Printf("\nGiving up waiting for build to exist\n")
			os.Exit(1)
		}
		fmt.Printf(".")
		time.Sleep(3 * time.Second)
	}

	go func(build *gojenkins.Build) {

		err := build.StreamOutput()
		if err != nil {
			fmt.Printf("Couldn't stream output: %s\n", err)
		}
	}(build)

	if build.GetResult() == "" {
		// Wait for build to start
		for {
			//fmt.Printf("2.Result: %s\n", build.GetResult())
			if build.IsRunning() || build.GetResult() != "" {
				fmt.Printf("\nJob is Running\n")
				break
			}

			if time.Now().Sub(start) > maxWait {
				fmt.Printf("\nGiving up waiting for build to start\n")
				os.Exit(1)
			}
			fmt.Printf(".")
			time.Sleep(3 * time.Second)
		}
	}

	// Waiting for build to finish
	for {
		build.Poll()
		//fmt.Printf("3.Result: %s\n", build.GetResult())
		if build.GetResult() != "" {
			fmt.Printf("\nJob is Finished\n")
			break
		}

		if time.Now().Sub(start) > maxWait {
			fmt.Printf("\nGiving up waiting for build to finish\n")
			os.Exit(1)
		}
		fmt.Printf(".")
		time.Sleep(3 * time.Second)
	}

	if build.GetResult() == "SUCCESS" {
		fmt.Printf("Result: %v\n", build.Raw.Description.(string))
	} else {
		fmt.Printf("Job %s\n", build.GetResult())
	}
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
