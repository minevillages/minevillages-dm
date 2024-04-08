package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 메시지 구조체
type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Content  string `json:"content"`
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

func (m *Message) Find() ([]Message, error) {
	collection := &Collection{
		Database:   "GameVillages",
		Collection: "DM",
	}
	DMcollection := collection.Get()

	filter := bson.M{"receiver": bson.M{"$exists": true}}
	cursor, err := DMcollection.Find(context.Background(), filter)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var messages []Message
	for cursor.Next(context.Background()) {
		var msg Message
		if err := cursor.Decode(&msg); err != nil {
			log.Fatalln(err)
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err := cursor.Err(); err != nil {
		log.Fatalln(err)
		return nil, err
	}

	return messages, nil
}
