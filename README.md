# Firestore rollback for golang
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

## How to use it (TBD)
```
go get github.com/frikky/firestore-rollback-go
```

## Firestore cloud functions
How cloud functions are/can be used:
1. Frontend -> firestore -> cloud function -> firestore
2. Frontend -> cloud function -> firestore

This project is made specifically for handling the first case.
