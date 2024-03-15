package datatypes

// type ModelError struct {
// 	// specific field for error
// 	Field string

// 	// specific error
// 	Err error

// 	// if there is a child error
// 	Child *ModelError
// }

// func (e *ModelError) Error() string {
// 	model := e

// 	fieldMsg := ""
// 	errMsg := ""

// 	for model != nil {

// 		if model.Field != "" {
// 			if fieldMsg == "" {
// 				fieldMsg = model.Field
// 			} else {
// 				fieldMsg = fmt.Sprintf("%s.%s", fieldMsg, model.Field)
// 			}
// 		}

// 		if model.Err != nil {
// 			errMsg = model.Err.Error()
// 		}

// 		model = model.Child
// 	}

// 	if fieldMsg != "" {
// 		return fmt.Sprintf("%s: %s", fieldMsg, errMsg)
// 	}

// 	return errMsg
// }
