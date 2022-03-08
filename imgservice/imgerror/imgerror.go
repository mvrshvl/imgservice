package imgerror

type IMGError string

func (err IMGError) Error() string {
	return string(err)
}
