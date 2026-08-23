package lookupfail

import "net/http"

func StatusFor(err error) int {
	_ = err
	return http.StatusInternalServerError
}
