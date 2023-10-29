package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Location struct {
	ID          string    `bson:"_id,omitempty"`
	Name        string    `bson:"name"`
	Coordinates GeoPoint  `bson:"coordinates"`
}

type GeoPoint struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}

func insertLocation(collection *mongo.Collection, location Location) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, location)
	return result, err
}

func createGeoIndex(collection *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys:    bson.M{"coordinates": "2dsphere"},
		Options: options.Index().SetBackground(true),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Fatalf("Erro ao criar índice: %v", err)
	}
}

func updateLocationByName(collection *mongo.Collection, name string, updatedLocation Location) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"name": name}
	update := bson.M{"$set": updatedLocation}

	result, err := collection.UpdateOne(ctx, filter, update)
	return result, err
}

func findLocationByName(collection *mongo.Collection, name string) (Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"name": name}
	var location Location
	err := collection.FindOne(ctx, filter).Decode(&location)
	return location, err
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Erro ao conectar ao MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("geoApp")
	collection := db.Collection("locations")

	// 1. Inserir um registro
	loc := Location{
		Name: "Ponto A",
		Coordinates: GeoPoint{
			Type:        "Point",
			Coordinates: []float64{-46.6584693, -23.5630994},
		},
	}
	res, err := insertLocation(collection, loc)
	if err != nil {
		log.Fatalf("Erro ao inserir localização: %v", err)
	}
	fmt.Printf("Inserido com ID: %v\n", res.InsertedID)

	// 2. Atualizar esse registro
	updatedLocation := Location{
		Name: "Ponto A Atualizado",
		Coordinates: GeoPoint{
			Type:        "Point",
			Coordinates: []float64{-46.6600000, -23.5600000},
		},
	}
	updateResult, err := updateLocationByName(collection, "Ponto A", updatedLocation)
	if err != nil {
		log.Fatalf("Erro ao atualizar localização: %v", err)
	}
	fmt.Printf("Atualizadas %v localizações\n", updateResult.ModifiedCount)

	// 3. Exibir o registro atualizado
	foundLocation, err := findLocationByName(collection, "Ponto A Atualizado")
	if err != nil {
		log.Fatalf("Erro ao buscar localização pelo nome: %v", err)
	}
	fmt.Printf("Local encontrado: %+v\n", foundLocation)

	// Inserir um segundo registro e mostrar a leitura
	loc2 := Location{
		Name: "Ponto B",
		Coordinates: GeoPoint{
			Type:        "Point",
			Coordinates: []float64{-46.6700000, -23.5700000},
		},
	}
	_, err = insertLocation(collection, loc2)
	if err != nil {
		log.Fatalf("Erro ao inserir localização 2: %v", err)
	}

	foundLocation2, err := findLocationByName(collection, "Ponto B")
	if err != nil {
		log.Fatalf("Erro ao buscar localização 2 pelo nome: %v", err)
	}
	fmt.Printf("Local 2 encontrado: %+v\n", foundLocation2)

	createGeoIndex(collection)
}
