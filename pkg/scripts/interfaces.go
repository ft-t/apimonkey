package scripts

import "net/http"

//go:generate mockgen -destination interfaces_mocks_test.go -package scripts_test -source=interfaces.go
//go:generate mockgen -destination read_closer_mocks_test.go -package scripts_test io ReadCloser

type HTTPClient interface {
	Do(request *http.Request) (*http.Response, error)
}
