package scripts

import "net/http"

//go:generate mockgen -destination interfaces_mocks_test.go -package scripts_test -source=interfaces.go

type HTTPClient interface {
	Do(request *http.Request) (*http.Response, error)
}
