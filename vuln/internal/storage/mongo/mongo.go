package mongo

import (
	"context"
	"vuln/internal/domain"
)

type MongoVulnRepository struct {
	collection *mongo.Collection
}

func (r *MongoVulnRepository) Save(ctx context.Context, v *domain.Vulnerability) error {
	_, err := r.collection.InsertOne(ctx, v)
	return err
}

func (r *MongoVulnRepository) Upsert(ctx context.Context, v *domain.Vulnerability) error {
	filter := bson.M{"_id": v.ID}
	update := bson.M{"$set": v}
	_, err := r.col.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}
