// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/ligato/cn-infra/db/sql"
	"github.com/ligato/cn-infra/db/sql/cassandra"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/satori/go.uuid"
	"time"
	"github.com/smartystreets/assertions"
)

//connectivity used to verify connectivity with Cassandra
func connectivity(db *cassandra.BrokerCassa) (err error) {

	var insertTweet1 = &tweet{ID: uuid.NewV4().String(), Timeline: "me1", Text: "hello world1", User: "user1"}
	insertErr1 := insert(db, insertTweet1)
	if insertErr1 != nil {
		return insertErr1
	}
	var insertTweet2 = &tweet{ID: uuid.NewV4().String(), Timeline: "me2", Text: "hello world2", User: "user2"}
	insertErr2 := insert(db, insertTweet2)
	if insertErr2 != nil {
		return insertErr2
	}

	selectErr := selectByID(db, &insertTweet1.ID)
	if selectErr != nil {
		return selectErr
	}

	selectAllErr := selectAll(db)
	if selectAllErr != nil {
		return selectAllErr
	}

	return nil
}

//alterTable used to depict support for ALTER TABLE
func alterTable(db *cassandra.BrokerCassa) (err error) {

	err = db.Exec(`ALTER TABLE example.person ADD data text`)
	if err != nil {
		logroot.StandardLogger().Errorf("Error executing alter table %v", err)
		return err
	}

	var insertPerson = &person{ID: uuid.NewV1().String(), Name: "James Bond", Data: "new column added"}
	insertErr := insertPersonTable(db, insertPerson)
	if insertErr != nil {
		logroot.StandardLogger().Fatalf("Error executing insert %v", err)
		return insertErr
	}

	selectErr := selectPersonByID(db, &insertPerson.ID)
	if selectErr != nil {
		logroot.StandardLogger().Errorf("select error = %v", selectErr)
		return selectErr
	}

	return nil
}

//createKeySpaceIfNotExist used to depict support for IF NOT EXISTS clause while creating a keyspace
func createKeySpaceIfNotExist(db *cassandra.BrokerCassa) (err error) {

	//creating a non existing keyspace
	err1 := db.Exec(`CREATE KEYSPACE IF NOT EXISTS example2 with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }`)

	if err1 != nil {
		logroot.StandardLogger().Errorf("Error creating keyspace %v", err1)
		return err1
	}

	//creating a non existing keyspace
	err2 := db.Exec(`CREATE KEYSPACE IF NOT EXISTS example with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }`)

	if err2 != nil {
		logroot.StandardLogger().Errorf("Error creating keyspace %v", err2)
		return err2
	}

	//does not return error for existing key space, even though key space exists since using 'IF NOT EXISTS'
	err3 := db.Exec(`CREATE KEYSPACE IF NOT EXISTS example with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }`)

	if err3 != nil {
		logroot.StandardLogger().Errorf("Error creating existing keyspace %v", err3)
		return err3
	}

	//will return an error since key space exists and not using 'IF NOT EXISTS'
	err4 := db.Exec(`CREATE KEYSPACE example with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }`)

	if err4 != nil {
		logroot.StandardLogger().Errorf("Error %v", err4)
		assertions.ShouldEqual(err4.Error(), "Cannot add existing keyspace \"example\"")
	}

	return nil
}

//insertCustomizedDataStructure used to depict support for customized data structure, creating and using user-defined types and storing/retrieval
func insertCustomizedDataStructure(db *cassandra.BrokerCassa) (customAddress map[string]address, err error) {

	var homePhone phone
	homePhone.CountryCode = 1
	homePhone.Number = "408-123-1234"

	var cellPhone phone
	cellPhone.CountryCode = 1
	cellPhone.Number = "408-123-1235"

	var workPhone phone
	workPhone.CountryCode = 1
	workPhone.Number = "408-123-1236"

	var homeAddr address
	homeAddr.City = "San Jose"
	homeAddr.Street = "123 Tasman Drive"
	homeAddr.Zip = "95135"
	phoneMap1 := make(map[string]phone)
	phoneMap1["home"] = homePhone
	phoneMap1["cell"] = cellPhone
	homeAddr.Phones = phoneMap1

	var workAddr address
	workAddr.City = "San Jose"
	workAddr.Street = "255 E Tasman Drive"
	workAddr.Zip = "95134"
	phoneMap2 := make(map[string]phone)
	phoneMap2["work"] = workPhone
	workAddr.Phones = phoneMap2

	addressMap := make(map[string]address)
	addressMap["home"] = homeAddr
	addressMap["work"] = workAddr

	err1 := db.Exec(`CREATE KEYSPACE IF NOT EXISTS example2 with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }`)
	if err1 != nil {
		logroot.StandardLogger().Errorf("Error creating keyspace %v", err1)
		return nil, err1
	}

	err2 := db.Exec(`CREATE TYPE IF NOT EXISTS example2.phone (
			countryCode int,
			number text,
		)`)

	if err2 != nil {
		logroot.StandardLogger().Errorf("Error creating user-defined type phone %v", err2)
		return nil, err2
	}

	err3 := db.Exec(`CREATE TYPE IF NOT EXISTS example2.address (
			street text,
			city text,
			zip text,
			phones map<text, frozen<phone>>
		)`)

	if err3 != nil {
		logroot.StandardLogger().Errorf("Error creating user-defined type address %v", err)
		return nil, err3
	}

	err4 := db.Exec(`CREATE TABLE IF NOT EXISTS example2.user (
			ID text PRIMARY KEY,
			addresses map<text, frozen<address>>
		)`)

	if err4 != nil {
		logroot.StandardLogger().Errorf("Error creating table user %v", err4)
		return nil, err4
	}

	var insertUser1 = &user{ID: "user1", Addresses: addressMap}
	insertErr1 := insertUserTable(db, insertUser1)
	if insertErr1 != nil {
		logroot.StandardLogger().Errorf("Insert error = %v", insertErr1)
		return nil, insertErr1
	}

	logroot.StandardLogger().Info("insert successful")

	var UserTable = &user{}
	users := []*user{}

	query1 := sql.FROM(UserTable, sql.WHERE(sql.Field(&UserTable.ID, sql.EQ("user1"))))

	it := db.ListValues(query1)
	for {
		user := &user{}
		stop := it.GetNext(user)
		if stop {
			break
		}
		users = append(users, user)
	}
	itErr := it.Close()

	if itErr != nil {
		logroot.StandardLogger().Errorf("Error closing iterator %v", itErr)
		return nil, itErr
	}

	logroot.StandardLogger().Infof("users = %v", users)

	logroot.StandardLogger().Infof("address = %v", users[0].Addresses)

	return users[0].Addresses, nil
}

//reconnectInterval used to depict redial_interval timeout behavior
//need to manually bring down cassandra during sleep interval, after bring it back up again we can retrieve results
func reconnectInterval(db *cassandra.BrokerCassa) (err error) {

	var insertTweet1 = &tweet{ID: uuid.NewV4().String(), Timeline: "me1", Text: "hello world1", User: "user1"}
	insertErr1 := insert(db, insertTweet1)
	if insertErr1 != nil {
		return insertErr1
	}
	var insertTweet2 = &tweet{ID: uuid.NewV4().String(), Timeline: "me2", Text: "hello world2", User: "user2"}
	insertErr2 := insert(db, insertTweet2)
	if insertErr2 != nil {
		return insertErr2
	}

	selectErr := selectByID(db, &insertTweet1.ID)
	if selectErr != nil {
		return selectErr
	}

	//sleep for 5 minutes (need to restart cassandra manually in the meantime)
	logroot.StandardLogger().Infof("Sleep for 5 min - start %v", time.Now())
	time.Sleep(5 * time.Minute)
	logroot.StandardLogger().Infof("Sleep for 5 min - end")

	selectAllErr := selectAll(db)
	if selectAllErr != nil {
		return selectAllErr
	}

	return nil
}

//queryTimeout used to depict op_timeout timeout behavior
//need to update the config to a very low op_timeout value (600ns) to get the expected timeout error
func queryTimeout(db *cassandra.BrokerCassa) (err error) {

	var insertTweet1 = &tweet{ID: uuid.NewV4().String(), Timeline: "me1", Text: "hello world1", User: "user1"}
	insertErr1 := insert(db, insertTweet1)
	if insertErr1 != nil {
		return insertErr1
	}
	var insertTweet2 = &tweet{ID: uuid.NewV4().String(), Timeline: "me2", Text: "hello world2", User: "user2"}
	insertErr2 := insert(db, insertTweet2)
	if insertErr2 != nil {
		return insertErr2
	}

	selectErr := selectByID(db, &insertTweet1.ID)
	if selectErr != nil {
		return selectErr
	}

	selectAllErr := selectAll(db)
	if selectAllErr != nil {
		return selectAllErr
	}

	return err
}

//connectTimeout used to depict dial_timeout timeout behavior
//need to update the config to a very low dial_timeout value (600ns) to get the expected timeout error
func connectTimeout(db *cassandra.BrokerCassa) (err error) {

	var insertTweet1 = &tweet{ID: uuid.NewV4().String(), Timeline: "me1", Text: "hello world1", User: "user1"}
	insertErr1 := insert(db, insertTweet1)
	if insertErr1 != nil {
		return insertErr1
	}
	var insertTweet2 = &tweet{ID: uuid.NewV4().String(), Timeline: "me2", Text: "hello world2", User: "user2"}
	insertErr2 := insert(db, insertTweet2)
	if insertErr2 != nil {
		return insertErr2
	}

	selectErr := selectByID(db, &insertTweet1.ID)
	if selectErr != nil {
		return selectErr
	}

	selectAllErr := selectAll(db)
	if selectAllErr != nil {
		return selectAllErr
	}

	return err
}

//insert used to insert data in tweet table
func insert(db *cassandra.BrokerCassa, insertTweet *tweet) (err error) {
	//inserting a record (runs update behind the scene)
	start1 := time.Now()
	err1 := db.Put(sql.FieldEQ(&insertTweet.ID), insertTweet)
	if err1 != nil {
		elapsed1 := time.Since(start1)
		logroot.StandardLogger().Infof("Time taken for insert : %v", elapsed1)
		logroot.StandardLogger().Errorf("Error executing insert %v", err1)
		return err1
	}

	elapsed := time.Since(start1)
	logroot.StandardLogger().Infof("Time taken for insert : %v", elapsed)
	return nil
}

//insertPersonTable used to insert data in person table
func insertPersonTable(db *cassandra.BrokerCassa, insertPerson *person) (err error) {
	//inserting a record (runs update behind the scene)
	start1 := time.Now()
	err1 := db.Put(sql.FieldEQ(&insertPerson.ID), insertPerson)
	if err1 != nil {
		elapsed1 := time.Since(start1)
		logroot.StandardLogger().Infof("Time taken for insert : %v", elapsed1)
		logroot.StandardLogger().Errorf("Error executing insert %v", err1)
		return err1
	}

	elapsed := time.Since(start1)
	logroot.StandardLogger().Infof("Time taken for insert : %v", elapsed)
	return nil
}

//insertUserTable used to insert data in user table
func insertUserTable(db *cassandra.BrokerCassa, insertUser *user) (err error) {
	//inserting a record (runs update behind the scene)
	start1 := time.Now()
	err1 := db.Put(sql.FieldEQ(&insertUser.ID), insertUser)
	if err1 != nil {
		elapsed1 := time.Since(start1)
		logroot.StandardLogger().Infof("Time taken for insert : %v", elapsed1)
		logroot.StandardLogger().Errorf("Error executing insert %v", err1)
		return err1
	}

	elapsed := time.Since(start1)
	logroot.StandardLogger().Infof("Time taken for insert : %v", elapsed)
	return nil
}

//selectByID used to retrieve data from tweet table
func selectByID(db *cassandra.BrokerCassa, id *string) (err error) {
	start2 := time.Now()
	var TweetTable = &tweet{}
	tweets := &[]tweet{}

	query1 := sql.FROM(TweetTable, sql.WHERE(sql.Field(&TweetTable.ID, sql.EQ(id))))
	err = sql.SliceIt(tweets, db.ListValues(query1))

	if err != nil {
		elapsed2 := time.Since(start2)
		logroot.StandardLogger().Infof("Time taken for select : %v", elapsed2)
		logroot.StandardLogger().Errorf("Error executing select %v", err)
		return err
	}

	elapsed2 := time.Since(start2)
	logroot.StandardLogger().Infof("Time taken for select : %v", elapsed2)
	logroot.StandardLogger().Info("Tweet:", tweets)
	return nil
}

//selectPersonByID used to retrieve data from person table
func selectPersonByID(db *cassandra.BrokerCassa, id *string) (err error) {
	start2 := time.Now()
	var PersonTable = &person{}
	people := &[]person{}

	query1 := sql.FROM(PersonTable, sql.WHERE(sql.Field(&PersonTable.ID, sql.EQ(id))))
	err = sql.SliceIt(people, db.ListValues(query1))

	if err != nil {
		elapsed2 := time.Since(start2)
		logroot.StandardLogger().Infof("Time taken for select : %v", elapsed2)
		logroot.StandardLogger().Errorf("Error executing select %v", err)
		return err
	}

	elapsed2 := time.Since(start2)
	logroot.StandardLogger().Infof("Time taken for select : %v", elapsed2)
	logroot.StandardLogger().Info("People:", people)
	return nil
}

//selectAll used to retrieve all records from tweet table
func selectAll(db *cassandra.BrokerCassa) (err error) {
	start3 := time.Now()
	var TweetTable = &tweet{}
	query2 := sql.FROM(TweetTable, nil)
	iterator := db.ListValues(query2)
	for {
		tweetItem := &tweet{}
		stop := iterator.GetNext(tweetItem)
		if stop {
			break
		} else {
			logroot.StandardLogger().Info("Tweet Item: ", tweetItem)
		}
	}
	iterator.Close()
	elapsed3 := time.Since(start3)
	logroot.StandardLogger().Infof("Time taken for select all : %v", elapsed3)
	return nil
}

// SchemaName schema name for tweet table
func (entity *tweet) SchemaName() string {
	return "example"
}

// SchemaName schema name for person table
func (entity *person) SchemaName() string {
	return "example"
}

// SchemaName schema name for user table
func (entity *user) SchemaName() string {
	return "example2"
}
