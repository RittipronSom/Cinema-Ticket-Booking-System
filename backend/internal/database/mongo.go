	package database

	import (
		"context"
		"time"
		"os"
		"go.mongodb.org/mongo-driver/mongo"
		"go.mongodb.org/mongo-driver/mongo/options"
	)

	var DB *mongo.Database

	func ConnectMongo() error {

		client, err := mongo.Connect(
			context.Background(),
			options.Client().ApplyURI(os.Getenv("MONGO_URI")),
		)

		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = client.Ping(ctx, nil)

		if err != nil {
			return err
		}

		DB = client.Database("cinema_booking")

		return nil
	}