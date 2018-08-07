package util

// ToString get the string of error
func ToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
