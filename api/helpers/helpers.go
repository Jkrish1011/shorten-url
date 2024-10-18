package helpers

import (
	"os"
	"strings"
)

func EnforceHTTP(url string) string {
	if url[:4] != "http" {
		return "http://" + url
	}
	return url

}

func RemoveDomainError(url string) bool {
	if url == os.Getenv("DOMAIN") {
		return false
	}
	/*
		Remove all commonly found prefix for localhost - IF we try to process localhost, the system will go into an infinte loop
		crashing our systems.

		example:
		localhost:3000
		http://localhost:3000
		https://localhost:3000

	*/
	newURL := strings.Replace(url, "http://", "", 1)
	newURL = strings.Replace(newURL, "http://", "", 1)
	newURL = strings.Replace(newURL, "www.", "", 1)

	newURL = strings.Split(newURL, "/")[0]

	if newURL == os.Getenv("DOMAIN") {
		return false
	}

	return true
}
