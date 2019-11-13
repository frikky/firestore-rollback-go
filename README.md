# Firestore rollback helper for golang
This has been built to handle the rollback a firestore event from the new value to and old value.
The reason this now exists is because of the horrendous design feeling of data you retrieve from firestore. 

See more about values: https://cloud.google.com/firestore/docs/reference/rest/v1/Value
An example is the following:
```go
// This struct:
type Tmp struct {
	Id string `json:"id"`
}

// would become
type Tmp struct {
	Id struct {
		StringValue string `json:"stringValue"`
	} `json:"stringValue"`
}

// TBD: Add all fields (bool etc). Arrays and maps are solved.
// Don't even get me started on arrays and maps...
```

## Setup and use
```bash
go get github.com/frikky/firestore-rollback-go
```

Define according to sample: https://github.com/GoogleCloudPlatform/golang-samples/blob/master/functions/firebase/upper/upper.go
```go
import (
	"github.com/frikky/firestore-rollback-go"
)

// Data from cloud function comes here
type FirestoreEvent struct {
	OldValue   FirestoreValue `json:"oldValue"`
	Value      FirestoreValue `json:"value"`
	UpdateMask struct {
		FieldPaths []string `json:"fieldPaths"`
	} `json:"updateMask"`
}

// FirestoreValue holds Firestore fields.
type FirestoreValue struct {
	CreateTime time.Time    `json:"createTime"`
	UserId     string       `json:"userId"`
	Name       string       `json:"name"`
	UpdateTime time.Time    `json:"updateTime"`
	Fields     MyData 		`json:"fields"`
}

// MyData represents a value from Firestore. The type definition depends on the
// format of your database.
type MyData struct {
	Original struct {
		StringValue string `json:"stringValue"`
	} `json:"original"`
}

// Now that you got it parsed into MyData, you will have to fix the data to rollback
// This will give you back a map[string]interface that can be sent back to backend

data := fsf.GetInterface(oldValue.Fields)
// ret, err := client.Collection("THISCOLLECTION").Doc("THISDOC").Set(ctx, data)
// if err != nil {
// ...
// }
```

## TBD: Make transformer for normal struct -> firestore struct
## TBD: Make a rollback(name, data) event to handle everything for you

## Firestore cloud functions
How cloud functions are/can be used:
1. Frontend -> firestore -> cloud function -> firestore
2. Frontend -> cloud function -> firestore

This project is made specifically for handling the first case.

