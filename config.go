package main

type Configuration struct {
	InputFiles []*FileConfig `yaml:"input-files"`
}

type FileConfig struct {
	Path string
}
