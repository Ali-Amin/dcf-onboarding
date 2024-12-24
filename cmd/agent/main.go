package main

import (
	"flag"
	"os"
)

func main() {
	// Load vrrs
	var onboardingServiceURL string
	flag.StringVar(&onboardingServiceURL,
		"onboarding-service-url",
		"./cmd/onboarder/res/config-k8s.json",
		"Location of the onboarding service (e.g., http://180.16.12.5:35000)")
	flag.Parse()

	os.WriteFile("/tmp/agent-logs", []byte("This is a test"), os.ModeAppend)
}
