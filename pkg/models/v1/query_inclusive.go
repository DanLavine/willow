package v1

// TODO: revisit this. I don't need this right now
//
//type Join string
//
//var (
//	JoinAnd Join = "and" // higher precedence over or
//	JoinOr  Join = "or"  // lower precedence over and
//)
//
//type KeyValues map[datatypes.String]datatypes.String
//
//// Query to use for any APIs
//type QueryInclusive struct {
//	// required broker name to search
//	BrokerName datatypes.String
//
//	// specific matches to find
//	Select *SelectInclusive
//}
//
//type SelectInclusive struct {
//	// only 1 can be used at a time
//	//// Find a Set of key value pairs
//	Set *SetInclusive
//	//// Find a subset of key value pairs
//	SubsetSelect *SelectInclusive
//
//	Join       *Join
//	JoinSelect *SelectInclusive
//}
//
//type SetIncusive struct {
//	// only 1 can be used at a time
//	//// Exactly searches for a key value set where these are the only key value pairs
//	Exactly KeyValues
//	//// Matches searches for any sets that contain all provided key value pairs
//	Matches KeyValues
//
//	// Joins are used to select specific items all from the same table
//	// type of join (and, or)
//	Join *Join
//
//	// join at the current level
//	JoinSet *SetIncusive
//}
//
//func (iw *WhereInclusive) And(inclusiveWhere WhereInclusive) {
//	iw.join = &and
//	iw.joinWhere = &inclusiveWhere
//}
//
//func (iw *WhereInclusive) Or(inclusiveWhere WhereInclusive) {
//	iw.join = &or
//	iw.joinWhere = &inclusiveWhere
//}
//
//func (iw *WhereInclusive) JoinAnd(inclusiveWhere WhereInclusive) {
//	iw.join = &and
//	iw.joinWhereInclusive = &inclusiveWhere
//}
//
//func (iw *WhereInclusive) JoinOr(inclusiveWhere WhereInclusive) {
//	iw.join = &or
//	iw.joinWhereInclusive = &inclusiveWhere
//}
//
//func ParseQueryInclusive(reader io.ReadCloser) (*QueryInclusive, *Error) {
//	body, err := io.ReadAll(reader)
//	if err != nil {
//		return nil, InvalidRequestBody.With("", err.Error())
//	}
//
//	query := &QueryInclusive{}
//	if err := json.Unmarshal(body, query); err != nil {
//		return nil, ParseRequestBodyError.With("query to be valid json", err.Error())
//	}
//
//	if err := query.Validate(); err != nil {
//		return nil, err
//	}
//
//	return query, nil
//}
//
//func (qi *QueryInclusive) Validate() *Error {
//	if qi.BrokerName == "" {
//		return &Error{Message: "BrokerName cannot be empty", StatusCode: http.StatusBadRequest}
//	}
//
//	if qi.Where != nil {
//		return qi.Where.Validate()
//	}
//
//	return nil
//}
//
//func (iw *WhereInclusive) Validate() *Error {
//	if iw.ExactWhere != nil && iw.AddWhere != nil {
//		return &Error{Message: "Only ExactWhere OR AddWhere can be inclueded at a time, not both"}
//	} else if iw.ExactWhere == nil && iw.AddWhere == nil {
//		return &Error{Message: "ExactWhere OR AddWhere must be inclueded in a query"}
//	}
//
//	if len(iw.ExactWhere) == 0 && len(iw.AddWhere) == 0 {
//		return &Error{Message: "Where clause is empty. Needs to be populated with at least 1 key value pair to find"}
//	}
//
//	if iw.join != nil {
//		switch *iw.join {
//		case WhereAnd, WhereOr:
//			if iw.joinWhere == nil && iw.joinWhereInclusive == nil {
//				return &Error{Message: "Have a join type, but no join clause. Need to have either ['joinWhere' | 'joinWhereInclusive']"}
//			} else if iw.joinWhere != nil && iw.joinWhereInclusive != nil {
//				return &Error{Message: "Have a join type, with multiple join clauses. Need to have only one of ['joinWhere' | 'joinWhereInclusive']"}
//			}
//
//			if iw.joinWhere != nil {
//				return iw.joinWhere.Validate()
//			} else {
//				return iw.joinWhereInclusive.Validate()
//			}
//		default:
//			return &Error{Message: "Invalid Join type. Must be ['and' | 'or']"}
//		}
//	}
//
//	if iw.joinWhere != nil || iw.joinWhereInclusive != nil {
//		return &Error{Message: "Have a join clause, but no join type. Need to have either ['and' | 'or']"}
//	}
//
//	return nil
//}
