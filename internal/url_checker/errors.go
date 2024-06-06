package url_checker

import "net/url"

func DescribeURLError(err *url.Error) string {
	if err.Timeout() {
		return "Timeout"
	}
	if err.Temporary() {
		return "Temporary"
	}
	return err.Error()
}
