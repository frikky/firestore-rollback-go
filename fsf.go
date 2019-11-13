package fsf

import (
	//"cloud.google.com/go/firestore"
	"fmt"
	"log"
	"reflect"
	"strconv"
)

type IntegerValue struct {
	IntegerValue string `json:"integerValue"`
}

type StringValue struct {
	StringValue string `json:"stringValue"`
}

type ArrayValue struct {
	Values Value `json:"values"`
}

type Value struct {
	MapValue []MapValue `json:"mapValue"`
}

type MapValue struct {
	Fields Fields `json:"fields"`
}

type SelectedCharities struct {
	ArrayValue ArrayValue `json:"arrayValue"`
}

// This is fucking stupid - firestore values
type Contribution struct {
	AmountGiven       IntegerValue      `json:"amountGiven"`
	SelectedCharities SelectedCharities `json:"selectedCharities"`
}

type Fields struct {
	Amount IntegerValue `json:"amount"`
	Id     StringValue  `json:"id"`
}

/*
Function that loops over and finds the integervalue and stringvalues before putting them in
their allocated spaces

All unhandled values from https://cloud.google.com/firestore/docs/reference/rest/v1/Value:
booleanvalue,
nullvalue,
doublevalue,
timestampvalue,
bytesvalue,
referencevalue,
geopointvalue
*/
func iterate(subValue interface{}) interface{} {
	v := reflect.ValueOf(subValue)
	values := make([]interface{}, v.NumField())
	typeOfS := v.Type()

	normalValues := make(map[string]interface{})
	normalSet := false
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()
		fieldName := typeOfS.Field(i).Name

		curType := fmt.Sprintf("%s", reflect.TypeOf(values[i]))
		if curType == "[]interface {}" {
			return values[0]
		}

		if curType == "main.ArrayValue" {
			values[i] = iterate(values[i])
		} else if curType == "main.Value" {
			values[i] = iterate(values[i])
		} else if curType == "[]main.MapValue" {
			tmpvalues := make([]interface{}, len(values[i].([]MapValue)))
			for iter, sub := range values[i].([]MapValue) {
				tmpvalues[iter] = iterate(sub)
			}

			values[i] = tmpvalues
		} else if curType == "main.Fields" {
			values[i] = iterate(values[i])
		} else if curType == "main.IntegerValue" {
			normalSet = true
			tmp, err := strconv.Atoi(values[i].(IntegerValue).IntegerValue)
			if err != nil {
				panic("Error handling integervalue")
			}

			normalValues[fieldName] = tmp
		} else if curType == "main.StringValue" {
			normalSet = true
			normalValues[fieldName] = values[i].(StringValue).StringValue
		} else {
			log.Printf("UNHANDLED TYPE: %s", curType)
		}

		if normalSet {
			values[i] = normalValues
		}
	}

	if len(values) == 1 {
		return values[0]
	} else {
		return values
	}
}

func getInterface(subValue interface{}) map[string]interface{} {
	v := reflect.ValueOf(subValue)

	newValues := make(map[string]interface{})
	values := make([]interface{}, v.NumField())
	var err error
	typeOfS := v.Type()

	// Didn't find a good way to do everything in the same iter..
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()
		fieldName := typeOfS.Field(i).Name

		curType := fmt.Sprintf("%s", reflect.TypeOf(values[i]))
		if curType == "main.IntegerValue" {
			newValues[fieldName], err = strconv.Atoi(values[i].(IntegerValue).IntegerValue)
			if err != nil {
				panic("Error handling integervalue")
			}
		} else if curType == "main.StringValue" {
			newValues[fieldName] = values[i].(StringValue).StringValue
		} else {
			tmpItem := iterate(values[i])
			newValues[fieldName] = tmpItem
		}
	}

	return newValues
}
