package lpadata

type File struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

type FileUpload struct {
	Filename string `json:"filename"`
	Data     string `json:"data"`
}
