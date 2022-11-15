package errors

import "fmt"

var (
	ProtocolUnkownConnection = fmt.Errorf("Unkown protocol connection")

	Unimplemented = fmt.Errorf("Unimplemented")
)
