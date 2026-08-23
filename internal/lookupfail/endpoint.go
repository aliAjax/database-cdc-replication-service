package lookupfail

import "net/http"

func StatusFor(err error) int {
	switch Classify(err) {
	case KindMissing:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
