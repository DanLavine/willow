package btreeassociated

// TODO revisit this
// Find a value based on the query
//func (ct *compositeTree) FindInclusive(query *v1.InclusiveWhere, onFind datastructures.OnFind) []any {
//	return nil

//if query == nil {
//	return ct.findAll(onFind)
//}

//var values []any

//if query.ExactWhere != nil {
//	values = ct.findExact(query, onFind)
//} else {

//}

//return values
//}

//func (ct *compositeTree) findAll(onFind datastructures.OnFind) []any {
//	var values []any
//
//	iterate := func(key datatypes.CompareType, item any) {
//		values = append(values, item)
//	}
//
//	ct.idTree.Iterate(iterate)
//
//	return values
//}
//
//func (ct *compositeTree) findExact(query *v1.InclusiveWhere, onFind datastructures.OnFind) []any {
//	var values []any
//
//	castableCompositeKeyValues := ct.compositeKeyValues.Find(datatypes.Int(len(query.ExactWhere)), nil)
//	if castableCompositeKeyValues == nil {
//		return values
//	}
//	compositeKeyValues := castableCompositeKeyValues.(*keyValues)
//
//	idSet := set.New()
//	firstLoop := true
//
//	for searchKey, searchValue := range query.ExactWhere {
//		castableValues := compositeKeyValues.values.Find(searchKey, nil)
//		if castableValues == nil {
//			return values
//		}
//		valuesForKey := castableValues.(*keyValues)
//
//		if firstLoop {
//			if item := valuesForKey.values.Find(searchValue, onFindIDHolderAdd(idSet)); item == nil {
//				return values
//			}
//			firstLoop = false
//		} else {
//			if item := valuesForKey.values.Find(searchValue, onFindIDHolderKeep(idSet)); item == nil {
//				return values
//			}
//		}
//	}
//
//	if idSet.Len() == 1 {
//		exactItem := ct.idTree.Get(idSet.Values()[0])
//		if onFind != nil {
//			onFind(exactItem)
//		}
//
//		values = append(values, exactItem)
//	}
//
//	return values
//}
//
//// might be useful if we need the key value pais, but for now, we can just iterate on all ids
////func (ct *compositeTree) findAll(onFind datastructures.OnFind) []any {
////	values := []any{}
////
////	iterateIDHolders := func(key datatypes.CompareType, item any) {
////	}
////
////	iterateKeyValues := func(key datatypes.CompareType, item any) {
////		valuesForKey := item.(*keyValues)
////		valuesForKey.values.Iterate(iterateIDHolders)
////	}
////
////	iterateCompositeKeyValues := func(key datatypes.CompareType, item any) {
////		compositeKeyValues := item.(*keyValues)
////		compositeKeyValues.values.Iterate(iterateKeyValues)
////	}
////
////	ct.compositeKeyValues.Iterate(iterateCompositeKeyValues)
////
////	return values
////}
