package db2se

import (
	"db2se/client"
)

// hashiCupsClient creates an object storing
// the client.
type Client struct {
	db2 client.Client
}

//// newClient creates a new client to access HashiCups
//// and exposes it for any secrets or roles to use.
//func newClient() (*db2Client, error) {
//	return &db2Client{nil}, nil
//}

func NewClient() *Client {
	return &Client{}
}

