# jenkins-builder
Build projects in jenkins from your command line

```
Usage of ./jenkins-builder:
  -jenkins="": Your jenkins url
  -job="": Your job name
  -maxwait=2m0s: Maximum time to wait for a build
  -regex="": Regex for job name
  -tail=false: Tail output to console

```

Example:

`./jenkins-builder -jenkins="$JENKINS_URL" -job="$PROJECT_NAME"`

Support custom build parameters as arguments:

`./jenkins-builder -jenkins="$JENKINS_URL" -job="$PROJECT_NAME" BUILDPARAM1=A BUILDPARAM2=B`

Also include option to tail output directly to your terminal (will exit after 5min):

`./jenkins-builder -jenkins="$JENKINS_URL" -job="$PROJECT_NAME" -tail -maxwait=5m`
