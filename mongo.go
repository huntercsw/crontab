package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDB struct {
	uri string
	cli *mongo.Client
}

type TimePoint struct {
	StartTime int64 `bson: startTime`
	EndTime   int64 `bson: endTime`
}

type MongoRecodeTest struct {
	TimePoint
	Name    string `bson: name`
	Command string `bson: command`
	StdErr  string `bson: stdErr`
	StdOut  string `bson: stdOut`
}

type FetchBy struct {
	Name   string `bson: name`
}

func (m *MongoDB) Init() (err error) {
	var (
		ctx context.Context
	)
	m.uri = "mongodb://" + "192.168.0.100:27017"
	ctx, _ = context.WithTimeout(context.TODO(), 5*time.Second)
	if m.cli, err = mongo.Connect(ctx, options.Client().ApplyURI(m.uri).SetMaxPoolSize(5).SetMinPoolSize(2)); err != nil {
		fmt.Println("mongo init create mongo client error:", err)
		return
	}
	return
}

func (m *MongoDB) InsertOneRecord(ctx context.Context, dbName, collectionName, name, command, stdErr, stdOut string) (objID primitive.ObjectID, err error) {
	var (
		rsp *mongo.InsertOneResult
	)
	document := MongoRecodeTest{
		TimePoint: TimePoint{time.Now().Unix(), time.Now().Unix() + 5},
		Name:      name,
		Command:   command,
		StdErr:    stdErr,
		StdOut:    stdOut,
	}
	collection := m.cli.Database(dbName).Collection(collectionName)
	if rsp, err = collection.InsertOne(ctx, document); err != nil {
		fmt.Println("mongo insert error:", err)
	} else {
		objID = rsp.InsertedID.(primitive.ObjectID)
	}
	return
}

func (m *MongoDB) InsertManyRecord(ctx context.Context, dbName, collectionName, name, command, stdErr, stdOut string) (objIDs []interface{}, err error) {
	var (
		rsp *mongo.InsertManyResult
	)
	document := MongoRecodeTest{
		TimePoint: TimePoint{time.Now().Unix(), time.Now().Unix() + 5},
		Name:      name,
		Command:   command,
		StdErr:    stdErr,
		StdOut:    stdOut,
	}
	docSlice := []interface{}{document, document, document}
	collection := m.cli.Database(dbName).Collection(collectionName)
	if rsp, err = collection.InsertMany(ctx, docSlice); err != nil {
		fmt.Println("mongo insert many error:", err)
	} else {
		objIDs = rsp.InsertedIDs
	}
	return
}

func (m *MongoDB) Fetch(ctx context.Context) (res []MongoRecodeTest, err error) {
	var (
		cursor *mongo.Cursor
		sk int64
		li int64
	)
	collection := m.cli.Database("test").Collection("test_collection")
	findBy := FetchBy{Name:"yinuo-02"}
	sk, li = 0, 3
	findOpts := &options.FindOptions{Skip:&sk, Limit:&li}
	if cursor, err = collection.Find(ctx, findBy, findOpts); err != nil {
		fmt.Println("mongo fetch error:", err)
	} else {
		for cursor.Next(context.TODO()) {
			r := new(MongoRecodeTest)
			if err1 := cursor.Decode(r); err1 != nil {
				fmt.Println("record decode error:", err1)
				return
			} else {
				res = append(res, *r)
			}
		}
	}
	return
}


