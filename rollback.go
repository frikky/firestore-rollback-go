package rollback

import (
	"context"
	"errors"
	"fmt"
	"log"
	//"os"
	"reflect"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
)

type IntegerValue struct {
	IntegerValue string `json:"integerValue,omitempty"`
}
type StringValue struct {
	StringValue string `json:"stringValue"`
}
type NullValue struct {
	NulLValue string `json:"nullValue"`
}
type BooleanValue struct {
	BooleanValue string `json:"booleanValue"`
}
type DoubleValue struct {
	DoubleValue string `json:"doubleValue"`
}
type TimestampValue struct {
	TimestampValue string `json:"timestampValue"`
}
type BytesValue struct {
	BytesValue string `json:"bytesValue"`
}
type ReferenceValue struct {
	ReferenceValue string `json:"referenceValue"`
}

type ArrayValue struct {
	Values []Value `json:"values"`
}

// This is inside an array again. Always confuse.
type Value struct {
	MapValue     MapValue     `json:"mapValue,omitempty"`
	StringValue  StringValue  `json:"stringValue,omitempty"`
	IntegerValue IntegerValue `json:"integerValue,omitempty"`
	ArrayValue   ArrayValue   `json:"arrayValue,omitempty"`
	BooleanValue BooleanValue `json:"booleanValue,omitempty"`
	DoubleValue  DoubleValue  `json:"doubleValue,omitempty"`
}

//TimestampValue TimestampValue `json:"timestampValue,omitempty"`
//BytesValue     BytesValue     `json:"bytesValue,omitempty"`
//ReferenceValue ReferenceValue `json:"referenceValue,omitempty"`

type MapValue struct {
	Fields interface{} `json:"fields"`
}

func Rollback(ctx context.Context, client *firestore.Client, firestoreLocation string, subValue interface{}) (map[string]interface{}, *firestore.WriteResult, error) {
	checkNumber := 0
	startLocation := 5
	if !strings.HasPrefix(firestoreLocation, "project") {
		checkNumber = 1
		startLocation = 0
	}

	collections := []string{}
	docs := []string{}
	for cnt, item := range strings.Split(firestoreLocation, "/") {
		if cnt < startLocation {
			continue
		}

		if cnt%2 == checkNumber {
			docs = append(docs, item)
		} else {
			collections = append(collections, item)
		}
	}

	updateData := GetInterface(subValue)
	log.Printf("Ready data: %#v", updateData)
	updateData = GetInterface(updateData)
	log.Printf("Ready data2: %#v", updateData)

	if len(collections) == 0 || len(docs) == 0 {
		return updateData, nil, errors.New("No firestore location found.")
	}

	clientDoc := client.Collection(collections[0]).Doc(docs[0])
	for i := 1; i < len(collections); i += 2 {
		clientDoc = clientDoc.Collection(collections[i]).Doc(docs[i])
	}

	setter, err := clientDoc.Set(ctx, updateData)
	return updateData, setter, err

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

Special:
map[string]interface {}
[]interface {}
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
		log.Printf("TYPE: %s", curType)
		if curType == "rollback.ArrayValue" || strings.Contains(curType, "ArrayValue") {
			values[i] = iterate(values[i])
		} else if curType == "rollback.MapValue" || strings.Contains(curType, "MapValue") {
			values[i] = iterate(values[i])
		} else if curType == "[]rollback.Value" {
			tmpvalues := make([]interface{}, len(values[i].([]Value)))
			for iter, sub := range values[i].([]Value) {
				tmpvalues[iter] = iterate(sub)
			}

			values[0] = tmpvalues
		} else if curType == "rollback.Fields" || strings.Contains(curType, "Fields") {
			values[i] = iterate(values[i])
		} else if curType == "rollback.IntegerValue" || strings.Contains(curType, "IntegerValue") {
			if len(values[i].(IntegerValue).IntegerValue) == 0 {
				continue
			}

			tmp, err := strconv.Atoi(values[i].(IntegerValue).IntegerValue)
			if err != nil {
				continue
			}

			normalSet = true
			normalValues[fieldName] = tmp
		} else if curType == "rollback.DoubleValue" || strings.Contains(curType, "DoubleValue") {
			tmp, err := strconv.ParseFloat(values[i].(DoubleValue).DoubleValue, 64)
			if err != nil {
				continue
			}

			normalSet = true
			normalValues[fieldName] = tmp
		} else if curType == "rollback.StringValue" || strings.Contains(curType, "StringValue") {
			if len(values[i].(StringValue).StringValue) > 0 {
				normalSet = true
				normalValues[fieldName] = values[i].(StringValue).StringValue
			}
		} else if curType == "rollback.NullValue" || strings.Contains(curType, "NullValue") {
			normalValues[fieldName] = nil
		} else if curType == "rollback.BooleanValue" || strings.Contains(curType, "BooleanValue") {
			value := values[i].(BooleanValue).BooleanValue
			if len(value) > 0 {
				if value == "true" {
					normalValues[fieldName] = true
				} else if value == "false" {
					normalValues[fieldName] = true
				} else {
					continue
				}

				normalSet = true
			}
		} else if curType == "map[string]interface {}" {
			log.Printf("NEW: How do I handle map[string]interface? \nValue: %#v\nFieldname: %s", values[i], fieldName)
			val := handleMap(values[i].(map[string]interface{}))
			newType := fmt.Sprintf("%s", reflect.TypeOf(val))
			if newType == "map[string]interface {}" {
				log.Printf("MAP??")
				normalValues[fieldName] = val.(map[string]interface{})
				normalSet = true
			} else if newType == "string" {
				normalValues[fieldName] = val.(string)
				normalSet = true
			} else if newType == "int" {
				tmpVal, err := strconv.Atoi(val.(string))
				if err != nil {
					continue
				}

				normalValues[fieldName] = tmpVal
				normalSet = true
			} else {
				log.Printf("NEW UNHANDLED TYPE (TBD): %s", newType)
			}
		} else {
			log.Printf("UNHANDLED TYPE: %s\n Value: %#v\n Fieldname: %s", curType, values[i], fieldName)
			//values[i] = iterate(values[i])

			// THis is just as a test. Idk if it will work or not for weird values lol
			normalSet = true
			normalValues[fieldName] = values[i]
		}
	}

	if normalSet {
		return normalValues
	}

	return values[0]
}

func iterateSlice(subValue []interface{}) interface{} {
	values := make([]interface{}, len(subValue))
	for i := 0; i < len(subValue); i++ {
		values[i] = subValue[i]

		curType := fmt.Sprintf("%s", reflect.TypeOf(values[i]))
		if curType == "map[string]interface {}" {
			values[i] = handleMap(values[i].(map[string]interface{}))
		} else {
			log.Printf("Missing []interface handler for %s", curType)
		}
	}

	return values
}

func handleMap(subValue map[string]interface{}) interface{} {
	newValues := make(map[string]interface{})
	for key, value := range subValue {
		curType := fmt.Sprintf("%s", reflect.TypeOf(value))
		// FIXME - add all the types
		if key == "stringValue" || key == "integerValue" {
			return value
		}

		log.Printf("curtype: %s, %#v", curType, value)
		if curType == "string" {
			tmpVal, err := strconv.Atoi(value.(string))
			if err == nil {
				newValues[key] = tmpVal
				continue
			}

			newValues[key] = value
			continue
		} else if curType == "int" {

			newValues[key] = value
			continue
		}

		// Array inside interface
		if curType == "[]interface {}" {
			val := iterateSlice(value.([]interface{}))
			newValues[key] = val
		} else if curType == "map[string]interface {}" {
			val := handleMap(value.(map[string]interface{}))
			newType := fmt.Sprintf("%s", reflect.TypeOf(val))

			if newType == "map[string]interface {}" {
				newValues[key] = val.(map[string]interface{})
			} else if newType == "string" {
				newValues[key] = val.(string)
			} else if newType == "int" {
				newValues[key] = val.(int)
			} else {
				log.Printf("UNHANDLED TYPE (TBD): %s", newType)
			}
		} else {
			log.Printf("UNHANDLED TYPE (TBD OUTER): %s", curType)
		}
	}

	return newValues
}

// passedField takes arrayValue
func GetInterface(subValue interface{}) map[string]interface{} {
	v := reflect.ValueOf(subValue)

	// The type should basically never be this.
	if (fmt.Sprintf("%s", reflect.TypeOf(subValue))) == "map[string]interface {}" {
		return handleMap(subValue.(map[string]interface{})).(map[string]interface{})
	}

	//values := make([]interface{}, v.NumField())
	newValues := make(map[string]interface{})
	values := make([]interface{}, v.NumField())
	typeOfS := v.Type()

	// Didn't find a good way to do everything in the same iter..
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()
		fieldName := typeOfS.Field(i).Name

		curType := fmt.Sprintf("%s", reflect.TypeOf(values[i]))
		if curType == "rollback.IntegerValue" {
			tmpVal, err := strconv.Atoi(values[i].(IntegerValue).IntegerValue)
			if err != nil {
				continue
			}
			newValues[fieldName] = tmpVal
		} else if curType == "rollback.StringValue" {
			newValues[fieldName] = values[i].(StringValue).StringValue
		} else {
			tmpItem := iterate(values[i])
			newValues[fieldName] = tmpItem
		}
	}

	return newValues
}
