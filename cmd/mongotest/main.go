package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	dburl, ok := os.LookupEnv("DB_URL")
	if !ok {
		log.Fatal("DB_URL must be set in environment or in .env file")
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dburl))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	coll := client.Database("testdb").Collection("messages")
	defer coll.Drop(context.TODO())

	mq := newMessageQueue(ctx, coll, 5)

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Println("starting to watch on stream")
		defer log.Println("stopping stream watching")

		return mq.watchQueue()
	})

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case msg := <-mq.outchan:
				log.Println("Entry found: ", msg.Body)
				mq.markMessage(msg)
			}
		}
	})

	g.Go(func() error {
		log.Println("starting to send messages")
		defer log.Println("done sending messages")
		return sendTestMessages(mq)
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

}

func sendTestMessages(mq *messagequeue) error {
	type sendmessage struct {
		body  string
		delay time.Duration
	}

	messages := []sendmessage{
		{"uninitialized", 750 * time.Millisecond},
		{"starting", 500 * time.Millisecond},
		{"started", 1250 * time.Millisecond},
		{"processing 1", 250 * time.Millisecond},
		{"processing 2", 250 * time.Millisecond},
		{"processing 3", 250 * time.Millisecond},
		{"processing 4", 50 * time.Millisecond},
		{"processing 5", 50 * time.Millisecond},
		{"processing 6", 50 * time.Millisecond},
		{"processing 7", 50 * time.Millisecond},
		{"processing 8", 50 * time.Millisecond},
		{"processing 9", 50 * time.Millisecond},
		{"processing 10", 50 * time.Millisecond},
		{"processing 11", 50 * time.Millisecond},
		{"processing 12", 50 * time.Millisecond},
		{"processing 13", 50 * time.Millisecond},
		{"processing 14", 50 * time.Millisecond},
		{"stopping", 750 * time.Millisecond},
		{"stopped", 1000 * time.Millisecond},
	}

	for _, sendmsg := range messages {
		time.Sleep(sendmsg.delay)
		if err := mq.sendMessage(sendmsg.body); err != nil {
			return err
		}
		// log.Println("Inserted at ", result.InsertedID)
	}
	return nil
}

type message struct {
	Id        string    `bson:"uuid"`
	Timestamp time.Time `bson:"timestamp"`
	Body      string    `bson:"body"`
	Processed bool      `bson:"processed"`
}

func newMessage(body string) message {
	return message{
		Id:        uuid.New().String(),
		Body:      body,
		Processed: false,
	}
}

type messagequeue struct {
	ctx      context.Context
	coll     *mongo.Collection
	outchan  chan message
	pageSize int
}

func newMessageQueue(ctx context.Context, coll *mongo.Collection, pageSize int) *messagequeue {
	return &messagequeue{
		ctx:      ctx,
		coll:     coll,
		outchan:  make(chan message),
		pageSize: pageSize,
	}
}

func (mq *messagequeue) sendMessage(body string) error {
	msg := newMessage(body)
	msg.Timestamp = time.Now()
	_, err := mq.coll.InsertOne(mq.ctx, msg)
	return err
}

func (mq *messagequeue) markMessage(msg message) error {
	return mq.coll.FindOneAndUpdate(mq.ctx, bson.D{{"uuid", msg.Id}}, bson.D{{"$set", bson.D{{"processed", true}}}}).Err()
}

func (mq *messagequeue) watchQueue() error {

	opts := options.Find().SetLimit(int64(mq.pageSize))
	for {
		select {
		case <-mq.ctx.Done():
			return nil
		default:
			cursor, err := mq.coll.Find(mq.ctx, bson.D{{"processed", false}}, opts)
			if err != nil {
				return err
			}

			for cursor.Next(mq.ctx) {
				msg := message{}
				if err := cursor.Decode(&msg); err != nil {
					return err
				}
				mq.outchan <- msg
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// TODO: this only works on replica sets :(
// func watchCollection(ctx context.Context, coll *mongo.Collection) error {
// 	stream, err := coll.Watch(ctx, mongo.Pipeline{})
// 	if err != nil {
// 		return err
// 	}
// 	defer stream.Close(ctx)

// 	for stream.Next(ctx) {
// 		val := bson.D{}
// 		if err := stream.Decode(&val); err != nil {
// 			return err
// 		}
// 		log.Println("Stream received value: ", val)
// 	}
// 	return nil
// }
