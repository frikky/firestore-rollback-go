package fsf

import (
	//"cloud.google.com/go/firestore"
	"fmt"
	"os"
	//"path/filepath"
	//"golang.org/x/tools/go/packages"
	"log"
	"reflect"
	"strconv"
	"strings"
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

// This is inside an array again. Always confuse.
type Value struct {
	MapValue     MapValue     `json:"mapValue,omitempty"`
	StringValue  StringValue  `json:"stringValue,omitempty"`
	IntegerValue IntegerValue `json:"integerValue,omitempty"`
	ArrayValue   ArrayValue   `json:"arrayValue,omitempty"`
}

type MapValue struct {
	Fields interface{} `json:"fields"`
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

	toBeRemoved := []int{}
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()
		fieldName := typeOfS.Field(i).Name

		curType := fmt.Sprintf("%s", reflect.TypeOf(values[i]))
		//if curType == "[]interface {}" {
		//	return values[0]
		//}

		log.Printf("TYPE: %s", curType)
		if curType == "fsf.ArrayValue" || strings.Contains(curType, "ArrayValue") {
			values[i] = iterate(values[i])
		} else if curType == "fsf.MapValue" || strings.Contains(curType, "MapValue") {
			values[i] = iterate(values[i])
		} else if curType == "[]fsf.Value" || strings.Contains(curType, "[]main.Value") {
			//log.Printf("GLOB: %#v",)
			//test := castType.Type()
			//log.Println(test)
			//log.Println(test.Elem())
			//test2 := test.Elem()
			//_ = test2

			//tmp := reflect.ValueOf(values[i])
			//switch tmp.Kind() {
			//case reflect.Slice, reflect.Array:
			//	log.Println(tmp.Kind())
			//	result := make([]interface{}, tmp.Len())
			//	for i := 0; i < tmp.Len(); i++ {
			//		result[i] = tmp.Index(i).Interface()
			//	}

			//	log.Println(result)
			//	log.Println(tmp.Len())
			//default:
			//	log.Println("NOT IN HERE")
			//}

			//tmpvalue := reflect.MakeSlice(reflect.ValueOf(values[i]), 1, 1)
			//log.Println(len(castType))
			//log.Printf("TYPE: %#v", castType)
			//test := castType
			//log.Printf("%#v", test)

			//vp := reflect.New(reflect.TypeOf(values[i]))
			//vp.Elem().Set(reflect.ValueOf(values[i]))

			//tmpvalues := make([]interface{}, tmp.Len())
			//log.Println(tmp.Interface().([]Value))
			//log.Printf("%#v", values[i].([]Value))
			//for iter, sub := range values[i].([]Value) {

			//log.Printf("%#v", vp)
			//for iter, sub := range tmp.Interface().([]Value) {
			//	log.Printf("%#v", sub)
			//	tmpvalues[iter] = iterate(sub, main)

			//	// FIXME - remove this
			//	break
			//	log.Printf("IN ITER: %#v", tmpvalues[iter])
			//}

			//values[i] = tmpvalues

			tmpvalues := make([]interface{}, len(values[i].([]Value)))
			//log.Printf("%#v", values[i].([]Value))
			for iter, sub := range values[i].([]Value) {
				log.Printf("%#v", sub)
				tmpvalues[iter] = iterate(sub)

				// FIXME - remove this
				break
				log.Printf("IN ITER: %#v", tmpvalues[iter])
			}

			values[i] = tmpvalues
		} else if curType == "fsf.Fields" || strings.Contains(curType, "Fields") {
			values[i] = iterate(values[i])
		} else if curType == "fsf.IntegerValue" || strings.Contains(curType, "IntegerValue") {
			normalSet = true

			if len(values[i].(IntegerValue).IntegerValue) == 0 {
				toBeRemoved = append(toBeRemoved, i)
				continue
			}

			tmp, err := strconv.Atoi(values[i].(IntegerValue).IntegerValue)
			if err != nil {
				toBeRemoved = append(toBeRemoved, i)
				log.Printf("Error handling integervalue for field %s. FieldValue: %s", fieldName, values[i].(IntegerValue).IntegerValue)
				continue
			}

			normalValues[fieldName] = tmp
		} else if curType == "fsf.StringValue" || strings.Contains(curType, "StringValue") {
			if len(values[i].(StringValue).StringValue) > 0 {
				toBeRemoved = append(toBeRemoved, i)
				normalSet = true
				normalValues[fieldName] = values[i].(StringValue).StringValue
			}
		} else {
			log.Printf("UNHANDLED TYPE: %s\nValue: %s, Fieldname: %s", curType, values[i], fieldName)
			normalValues[fieldName] = values[i]
		}
	}

	//log.Println(toBeRemoved)
	//if len(toBeRemoved) > 0 {
	//	log.Printf("%#v", values[toBeRemoved[0]])
	//}

	if normalSet {
		values[0] = normalValues
	}

	//log.Printf("Len: %d", len(values))
	if len(values) > 1 {
		log.Printf("Might be an error if I hit this one. Returning first value.\nItem: %#v", values)
	}

	return values[0]
}

// passedField takes arrayValue
func GetInterface(subValue interface{}) map[string]interface{} {
	//log.Println(packages.Load(main, "bytes", "unicode"))
	var err error
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(pwd)

	v := reflect.ValueOf(subValue)

	newValues := make(map[string]interface{})
	values := make([]interface{}, v.NumField())
	typeOfS := v.Type()

	// Didn't find a good way to do everything in the same iter..
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()
		fieldName := typeOfS.Field(i).Name

		curType := fmt.Sprintf("%s", reflect.TypeOf(values[i]))
		log.Printf("TYPE: %s", curType)
		if curType == "fsf.IntegerValue" {
			newValues[fieldName], err = strconv.Atoi(values[i].(IntegerValue).IntegerValue)
			if err != nil {
				log.Printf("Error handling integervalue for field %s", fieldName)
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
