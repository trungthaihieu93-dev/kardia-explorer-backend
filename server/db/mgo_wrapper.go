/*
 *  Copyright 2018 KardiaChain
 *  This file is part of the go-kardia library.
 *
 *  The go-kardia library is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Lesser General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The go-kardia library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Lesser General Public License for more details.
 *
 *  You should have received a copy of the GNU Lesser General Public License
 *  along with the go-kardia library. If not, see <http://www.gnu.org/licenses/>.
 */
// Package db
package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type KaiMgo struct {
	DB  *mongo.Database
	col *mongo.Collection
}

func (w *KaiMgo) Database(db *mongo.Database) {
	w.DB = db
}

func (w *KaiMgo) C(name string) *KaiMgo {
	w.col = w.DB.Collection(name)
	return w
}

func (w *KaiMgo) Ping() error {
	return nil
}

func (w *KaiMgo) EnsureIndex() {

}

func (w *KaiMgo) Update(filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return w.col.UpdateOne(context.TODO(), filter, update, opts...)
}

func (w *KaiMgo) Upsert(filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	opts = append(opts, options.Update().SetUpsert(true))
	return w.col.UpdateOne(context.TODO(), filter, bson.M{"$set": update}, opts...)
}

func (w *KaiMgo) RemoveAll(filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return w.col.DeleteMany(context.TODO(), filter, opts...)
}

func (w *KaiMgo) Find(filter interface{},
	opts ...*options.FindOptions) (*mongo.Cursor, error) {
	return w.col.Find(context.TODO(), filter, opts...)
}

func (w *KaiMgo) FindOne(filter interface{},
	opts ...*options.FindOneOptions) *mongo.SingleResult {
	return w.col.FindOne(context.TODO(), filter, opts...)
}

func (w *KaiMgo) Select(filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return w.col.DeleteMany(context.TODO(), filter, opts...)
}

func (w *KaiMgo) Sort(filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return w.col.DeleteMany(context.TODO(), filter, opts...)
}

func (w *KaiMgo) One(filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return w.col.DeleteMany(context.TODO(), filter, opts...)
}

func (w *KaiMgo) BulkWrite(models []mongo.WriteModel,
	opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	opts = append(opts, options.BulkWrite().SetOrdered(false), options.BulkWrite().SetBypassDocumentValidation(true))
	return w.col.BulkWrite(context.TODO(), models, opts...)
}

func (w *KaiMgo) BulkInsert(models []mongo.WriteModel,
	opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	opts = append(opts, options.BulkWrite().SetOrdered(false), options.BulkWrite().SetBypassDocumentValidation(true))
	return w.col.BulkWrite(context.TODO(), models, opts...)
}

func (w *KaiMgo) BulkUpsert(models []mongo.WriteModel,
	opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	opts = append(opts, options.BulkWrite().SetOrdered(false), options.BulkWrite().SetBypassDocumentValidation(true))
	return w.col.BulkWrite(context.TODO(), models, opts...)
}

func (w *KaiMgo) Count(filter interface{},
	opts ...*options.CountOptions) (int64, error) {
	return w.col.CountDocuments(context.TODO(), filter, opts...)
}

func (w *KaiMgo) Insert(document interface{},
	opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return w.col.InsertOne(context.TODO(), document, opts...)
}

func (w *KaiMgo) FindSetSort(data string) *options.FindOptions {
	if data[0:1] == "-" {
		return options.Find().SetSort(bson.M{data[1:]: -1})
	} else {
		return options.Find().SetSort(bson.M{data: 1})
	}

}

func (w *KaiMgo) FindOneSetSort(data string) *options.FindOneOptions {
	if data[0:1] == "-" {
		return options.FindOne().SetSort(bson.M{data[1:]: -1})
	} else {
		return options.FindOne().SetSort(bson.M{data: 1})
	}

}

func (w *KaiMgo) DropDatabase(ctx context.Context) {
	if err := w.DB.Drop(ctx); err != nil {
		return
	}
}
