package command

// RemoteError is an error (in contrast to a faliure) that occurred on the
// remote peer when calling a command.
type RemoteError string

func (err RemoteError) Error() string {
	if err == "" {
		return "unknown error"
	}

	return string(err)
}
