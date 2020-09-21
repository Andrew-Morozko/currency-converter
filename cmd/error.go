package main

type UIError struct {
	msg string
}

func (r UIError) Error() string {
	return r.msg
}
