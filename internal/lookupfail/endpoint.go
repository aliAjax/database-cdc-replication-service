package lookupfail

import "net/http"

func StatusFor(err error) int {
	if Classify(err) == KindMissing {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}
