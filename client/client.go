package client

import (
	"database/sql"

	"time"
	//"time"

	_ "github.com/ibmdb/go_ibm_db"
)

type Config struct {

	LastBindPassword         string    `json:"last_bind_password"`
	LastBindPasswordRotation time.Time `json:"last_bind_password_rotation"`
	ConnectionString		string `json:"connection_string"`
}

func NewClient() Client {
	return Client{}
}

type Client struct {
	Username string
	Password string
	OldPasword string
	ConnectionString  string
	RotateString	string
}



// UpdatePassword uses a Modify call under the hood instead of LDAP change password function.
// This allows AD and OpenLDAP secret engines to use the same api without changes to
// the interface.
func (c *Client) UpdatePassword(connectionString, username, oldpassword, newpassword string) error {

	//con1 := "HOSTNAME=localhost;DATABASE=dojo;PORT=50000;UID=moayad;PWD=T7S7d95d"
	//con := "HOSTNAME=localhost;DATABASE=dojo;PORT=50000;UID=moayad;PWD=T7S7d95d;NEWPWD=T7S7d95e"
	c.ConnectionString = connectionString + ";UID=" + username + ";PWD=" + oldpassword
	c.RotateString = connectionString + ";UID=" + username + ";PWD=" + oldpassword + ";NEWPWD=" + newpassword


	//Connect and change the password, with the NEWPWD parameter
	db, err := sql.Open("go_ibm_db", c.ConnectionString)
	db.Exec("DROP table rocket")
	_, err = db.Exec("create table rocket(a int)")
	if err != nil {
		println("error dropping table")
	} else {
		println("success creating and dropping table")
	}
	db.Close()


	// start another connect with the old password, using con
	//this should return an error since it shouldn't  be able to connect.

	db1, err := sql.Open("go_ibm_db", c.RotateString)
	db1.Exec("DROP table rocket")
	_, err = db1.Exec("create table rocket(a int)")
	if err != nil {
		println("error dropping table")
	} else {
		println("success creating and dropping table")
	}
	db1.Close()
	return err
}


