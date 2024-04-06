package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Message struct {
	Message   string    `json:"message"`
	Sender    string    `json:"sender"`
	Timestamp time.Time `json:"timestamp"`
	Read      bool      `json:"read"`
}

type Mongo struct {
	Uri string
}

type Collection struct {
	Database   string
	Collection string
}

var MongoClient *mongo.Client

func (m *Mongo) Connection() error {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(m.Uri))
	if err != nil {
		log.Println(err)
		return err
	}
	MongoClient = client
	log.Println("MongoDB Connected.")
	return nil
}

func (c *Collection) Get() *mongo.Collection {
	return MongoClient.Database(c.Database).Collection(c.Collection)
}

func (m *Message) Insert() error {
	collection := &Collection{
		Database:   "GameVillages",
		Collection: "DM",
	}
	DMcollection := collection.Get()
	messageBSON, err := bson.Marshal(m)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	_, err = DMcollection.InsertOne(context.Background(), messageBSON)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	return nil
}
