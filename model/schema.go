package model

// Inputs models a set of files (Schema files and other files) that have been found in a
// given directory.
type Inputs struct {
	Directory         string
	SchemaFiles       []SchemaFile
	OtherTypesOfFiles int // Placeholder for illustration
}

type SchemaFile struct {
	FileName string
	Contents string
}
