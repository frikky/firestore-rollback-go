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
	Values []Value `json:"values"`
}

type Value struct {
	MapValue MapValue `json:"mapValue"`
}

type MapValue struct {
	Fields Fields `json:"fields"`
}

type SelectedCharities struct {
	ArrayValue ArrayValue `json:"arrayValue"`
}

type Fields struct {
	Amount           IntegerValue `json:"amount"`
	Name             StringValue  `json:"name"`
	Id               StringValue  `json:"id"`
	Description      StringValue  `json:"description"`
	Image            StringValue  `json:"image"`
	ShortDescription StringValue  `json:"shortDescription"`
	SmallImage       StringValue  `json:"smallImage"`
	Reference        StringValue  `json:"reference"`
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

		if curType == "fsf.ArrayValue" {
			values[i] = iterate(values[i])
		} else if curType == "fsf.Value" {
			values[i] = iterate(values[i])
		} else if curType == "fsf.Values" {
			values[i] = iterate(values[i])
		} else if curType == "fsf.MapValue" {
			values[i] = iterate(values[i])
		} else if curType == "[]fsf.Value" {
			tmpvalues := make([]interface{}, len(values[i].([]Value)))
			for iter, sub := range values[i].([]Value) {
				tmpvalues[iter] = iterate(sub)
			}

			values[i] = tmpvalues
		} else if curType == "fsf.Fields" {
			values[i] = iterate(values[i])
		} else if curType == "fsf.IntegerValue" {
			normalSet = true
			tmp, err := strconv.Atoi(values[i].(IntegerValue).IntegerValue)
			if err != nil {
				log.Println("Error handling integervalue for field %s", fieldName)
				continue
			}

			normalValues[fieldName] = tmp
		} else if curType == "fsf.StringValue" {
			normalSet = true
			normalValues[fieldName] = values[i].(StringValue).StringValue
		} else if curType == "string" {
			normalSet = true
			normalValues[fieldName] = values[i]
		} else {
			log.Printf("UNHANDLED TYPE: %s\nValue: %s, Fieldname: %s", curType, values[i], fieldName)
			normalValues[fieldName] = values[i]
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

func GetInterface(subValue interface{}) map[string]interface{} {
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
		if curType == "fsf.IntegerValue" {
			newValues[fieldName], err = strconv.Atoi(values[i].(IntegerValue).IntegerValue)
			if err != nil {
				log.Println("Error handling integervalue for field %s", fieldName)
				continue
			}
		} else if curType == "fsf.StringValue" {
			newValues[fieldName] = values[i].(StringValue).StringValue
		} else {
			tmpItem := iterate(values[i])
			newValues[fieldName] = tmpItem
		}
	}

	return newValues
}
