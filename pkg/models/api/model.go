package api

// // All pkg models need to suport the functions to ensure encoding works properly
type APIObject interface {
	Validate() error
}
