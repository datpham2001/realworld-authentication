package repository

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Instance struct {
	db   *mongo.Database
	coll *mongo.Collection

	DBName         string
	ColName        string
	TemplateObject interface{}
}

func (m *Instance) ApplyDatabase(db *mongo.Database) *Instance {
	m.db = db
	m.coll = db.Collection(m.ColName)
	m.DBName = db.Name()

	return m
}

func (m *Instance) NewObject() interface{} {
	t := reflect.TypeOf(m.TemplateObject)

	v := reflect.New(t)
	return v.Interface()
}

func (m *Instance) NewList(limit int) interface{} {
	t := reflect.TypeOf(m.TemplateObject)
	return reflect.MakeSlice(reflect.SliceOf(t), 0, limit).Interface()
}

func (m *Instance) ConvertToBson(ent interface{}) (bson.M, error) {
	if ent == nil {
		return bson.M{}, nil
	}

	sel, err := bson.Marshal(ent)
	if err != nil {
		return nil, err
	}

	obj := bson.M{}
	_ = bson.Unmarshal(sel, &obj)

	return obj, nil
}

func (m *Instance) ConvertToObject(b bson.M) (interface{}, error) {
	obj := m.NewObject()

	if b == nil {
		return obj, nil
	}

	bytes, err := bson.Marshal(b)
	if err != nil {
		return nil, err
	}

	_ = bson.Unmarshal(bytes, obj)
	return obj, nil
}

func (m *Instance) Create(ent interface{}) (interface{}, error) {
	// check col
	if m.coll == nil {
		return nil, fmt.Errorf("DB error: Create - Collection " + m.ColName + " is not init.")
	}

	// convert to bson
	obj, err := m.ConvertToBson(ent)
	if err != nil {
		return nil, err
	}

	// init time
	if obj["created_time"] == nil {
		obj["created_time"] = time.Now()
	}

	// insert
	result, err := m.coll.InsertOne(context.TODO(), obj)
	if err != nil {
		return nil, err
	}

	obj["_id"] = result.InsertedID
	ent, _ = m.ConvertToObject(obj)

	list := m.NewList(1)
	listValue := reflect.Append(reflect.ValueOf(list), reflect.Indirect(reflect.ValueOf(ent)))

	return listValue.Interface(), nil
}

func (m *Instance) UpdateOne(query interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) (interface{}, error) {
	if m.coll == nil {
		return nil, fmt.Errorf("DB error: Collection " + m.ColName + "is not init")
	}

	bsonUpdate, err := m.ConvertToBson(update)
	if err != nil {
		return nil, err
	}
	bsonUpdate["last_updated_time"] = time.Now()

	converted, err := m.ConvertToBson(query)
	if err != nil {
		return nil, err
	}

	// do update
	if opts == nil {
		after := options.After
		opts = []*options.FindOneAndUpdateOptions{
			{
				ReturnDocument: &after,
			},
		}
	}

	result := m.coll.FindOneAndUpdate(context.TODO(), converted, bson.M{"$set": bsonUpdate}, opts...)
	if result.Err() != nil {
		detail := ""
		if result != nil {
			detail = result.Err().Error()
		}

		return nil, fmt.Errorf("Not found any matched " + m.ColName + ". Error detail: " + detail)
	}

	return m.ParseSingleResult(result, "UpdateOne")
}

func (m *Instance) ParseSingleResult(result *mongo.SingleResult, action string) (interface{}, error) {
	obj := m.NewObject()

	err := result.Decode(obj)
	if err != nil {
		return nil, err
	}

	list := m.NewList(1)
	listValue := reflect.Append(reflect.ValueOf(list), reflect.Indirect(reflect.ValueOf(obj)))

	return listValue.Interface(), nil
}

func (m *Instance) Query(query interface{}, offset, limit int64, sortFields *bson.M) (interface{}, error) {
	if m.coll == nil {
		return nil, fmt.Errorf("DB error: Collection " + m.ColName + " is not init.")
	}

	opt := &options.FindOptions{}
	k := int64(1000)
	if limit <= 0 {
		opt.Limit = &k
	} else {
		opt.Limit = &limit
	}
	if offset > 0 {
		opt.Skip = &offset
	}
	if sortFields != nil {
		opt.Sort = sortFields
	}

	converted, err := m.ConvertToBson(query)
	if err != nil {
		return nil, err
	}

	result, err := m.coll.Find(context.TODO(), converted, opt)
	if err != nil {
		return nil, err
	} else if result.Err() != nil {
		return nil, result.Err()
	}

	list := m.NewList(int(limit))
	err = result.All(context.TODO(), &list)
	_ = result.Close(context.TODO())
	if err != nil || reflect.ValueOf(list).Len() == 0 {
		return nil, err
	}

	return list, nil
}

func (m *Instance) QueryAll() (interface{}, error) {
	if m.coll == nil {
		return nil, fmt.Errorf("DB error: Collection " + m.ColName + " is not init.")
	}

	rs, err := m.coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}

	list := m.NewList(1000)
	_ = rs.All(context.TODO(), &list)
	_ = rs.Close(context.TODO())
	if reflect.ValueOf(list).Len() == 0 {
		return nil, fmt.Errorf("Not found any matched " + m.ColName + ".")
	}

	return list, nil
}

func (m *Instance) QueryOne(query interface{}) (interface{}, error) {
	if m.coll == nil {
		return nil, fmt.Errorf("DB error: Collection " + m.ColName + " is not init.")
	}

	// transform to bson
	converted, err := m.ConvertToBson(query)
	if err != nil {
		return nil, err
	}

	// do find
	result := m.coll.FindOne(context.TODO(), converted)

	if result == nil || result.Err() != nil {
		return nil, fmt.Errorf("Not found any matched " + m.ColName + ".")
	}

	return m.ParseSingleResult(result, "Query")
}

func (m *Instance) CreateIndex(keys bson.D, options *options.IndexOptions) error {
	_, err := m.coll.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    keys,
		Options: options,
	})

	return err
}

// Count Count object which matched with query.
func (m *Instance) Count(query interface{}) (interface{}, error) {
	// check col
	if m.coll == nil {
		return nil, fmt.Errorf("DB error: Collection " + m.ColName + " is not init.")
	}

	// convert query
	converted, err := m.ConvertToBson(query)
	if err != nil {
		return nil, err
	}

	count, err := m.coll.CountDocuments(context.TODO(), converted)
	if err != nil {
		return nil, err
	}

	return count, nil
}
