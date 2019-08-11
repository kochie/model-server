package model

type size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Depth  int `json:"depth"`
}

// Model contains a description of the machine learning model that is saved in S3 and how to obtain it.
type Model struct {
	Name       string `json:"name"`
	Version    int    `json:"version"`
	InputSize  size   `json:"inputSize"`
	OutputSize size   `json:"outputSize"`
}
