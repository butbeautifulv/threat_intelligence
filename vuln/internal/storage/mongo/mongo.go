package mongo

import (
	"context"
	"vuln/internal/domain"
	"vuln/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoVulnRepository struct {
	collection *mongo.Collection
}

func NewVulnerabilityRepository(db *mongo.Database) repository.VulnerabilityRepository {
	return &MongoVulnRepository{
		collection: db.Collection("vulnerabilities"),
	}
}

func (r *MongoVulnRepository) Save(ctx context.Context, v *domain.Vulnerability) error {
	_, err := r.collection.InsertOne(ctx, v)
	return err
}

func (r *MongoVulnRepository) FindByCVE(ctx context.Context, id string) (*domain.Vulnerability, error) {
	var v domain.Vulnerability
	err := r.collection.FindOne(ctx, bson.M{"cve": id}).Decode(&v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *MongoVulnRepository) Upsert(ctx context.Context, v *domain.Vulnerability) error {
	filter := bson.M{"cve": v.CVE}
	update := bson.M{"$set": v}
	_, err := r.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}
