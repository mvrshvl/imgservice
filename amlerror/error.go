package amlerror

type AMLError string

func (err AMLError) Error() string {
	return string(err)
}
