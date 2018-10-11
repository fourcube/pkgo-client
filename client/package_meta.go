package client

type PackageMeta struct {
	Version   string `json:"version"`
	Name      string `json:"name"`
	Signature string `json:"signature"`
	// Unpack after download
	// Assumes .tar.gz format
	Unpack bool `json:"unpack"`
}
