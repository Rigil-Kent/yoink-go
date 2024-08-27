package comic

type ImageParseError struct {
	Message string
	Code    int
}

type ComicDownloadError struct {
	Message string
	Code    int
}

func (i ImageParseError) Error() string {
	return i.Message
}

func (c ComicDownloadError) Error() string {
	return c.Message
}
