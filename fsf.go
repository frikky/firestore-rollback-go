package fsf

import (
	"context"
	"fmt"
	"log"
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
		} else if cnt == startLocation {
			collections = append(collections, item)
			continue
		}

		if cnt%2 == checkNumber {
			docs = append(docs, item)
		} else {
			collections = append(collections, item)
		}
		log.Printf("COUNTER: %d, item: %s", cnt, item)
	}

	clientDoc := client.Collection(collections[0]).Doc(docs[0])
	for i := 1; i < len(collections); i += 2 {
		clientDoc = clientDoc.Collection(collections[i]).Doc(docs[i])
	}

	updateData := GetInterface(subValue)
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
*/
func iterate(subValue interface{}) interface{} {
	v := reflect.ValueOf(subValue)

	// This will stop working if its a map
	values := make([]interface{}, v.NumField())
	typeOfS := v.Type()

	normalValues := make(map[string]interface{})
	normalSet := false

	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()
		fieldName := typeOfS.Field(i).Name

		curType := fmt.Sprintf("%s", reflect.TypeOf(values[i]))
		if curType == "fsf.ArrayValue" || strings.Contains(curType, "ArrayValue") {
			values[i] = iterate(values[i])
		} else if curType == "fsf.MapValue" || strings.Contains(curType, "MapValue") {
			values[i] = iterate(values[i])
		} else if curType == "[]fsf.Value" {
			tmpvalues := make([]interface{}, len(values[i].([]Value)))
			for iter, sub := range values[i].([]Value) {
				tmpvalues[iter] = iterate(sub)
			}

			values[0] = tmpvalues
		} else if curType == "fsf.Fields" || strings.Contains(curType, "Fields") {
			values[i] = iterate(values[i])
		} else if curType == "fsf.IntegerValue" || strings.Contains(curType, "IntegerValue") {

			tmp, err := strconv.Atoi(values[i].(IntegerValue).IntegerValue)
			if err != nil {
				continue
			}

			normalSet = true
			normalValues[fieldName] = tmp
		} else if curType == "fsf.DoubleValue" || strings.Contains(curType, "DoubleValue") {
			tmp, err := strconv.ParseFloat(values[i].(DoubleValue).DoubleValue, 64)
			if err != nil {
				continue
			}

			normalSet = true
			normalValues[fieldName] = tmp
		} else if curType == "fsf.StringValue" || strings.Contains(curType, "StringValue") {
			if len(values[i].(StringValue).StringValue) > 0 {
				normalSet = true
				normalValues[fieldName] = values[i].(StringValue).StringValue
			}
		} else if curType == "fsf.NullValue" || strings.Contains(curType, "NullValue") {
			normalValues[fieldName] = nil
		} else if curType == "fsf.BooleanValue" || strings.Contains(curType, "BooleanValue") {
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
			log.Printf("How do I handle map[string]interface? \nValue: %#v\nFieldname: %s", values[i], fieldName)
			normalValues[fieldName] = values[i]
		} else {
			log.Printf("UNHANDLED TYPE: %s\n Value: %#v\n Fieldname: %s", curType, values[i], fieldName)
			//values[i] = iterate(values[i])

			// THis is just as a test. Idk if it will work or not for weird values lol
			normalSet = true
			normalValues[fieldName] = values[i]
		}
	}

	if normalSet {
		//log.Printf("PRE: %#v", values)
		// FIXME - this might be wrong
		//log.Printf("NORMAL %#v", normalValues)
		return normalValues
	}

	return values[0]
}

// passedField takes arrayValue
func GetInterface(subValue interface{}) map[string]interface{} {
	var err error
	v := reflect.ValueOf(subValue)

	newValues := make(map[string]interface{})
	values := make([]interface{}, v.NumField())
	typeOfS := v.Type()

	// Didn't find a good way to do everything in the same iter..
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()
		fieldName := typeOfS.Field(i).Name

		curType := fmt.Sprintf("%s", reflect.TypeOf(values[i]))
		//log.Printf("TYPE: %s", curType)
		if curType == "fsf.IntegerValue" {
			newValues[fieldName], err = strconv.Atoi(values[i].(IntegerValue).IntegerValue)
			if err != nil {
				//log.Printf("Error handling integervalue for field %s", fieldName)
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
