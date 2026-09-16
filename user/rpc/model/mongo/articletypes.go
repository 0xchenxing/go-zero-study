package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Article struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title     string        `c:"标题" o:"find,get,set"`
	Content   string        `c:"内容"`
	Published bool          `c:"是否发布"`
	UpdateAt  time.Time     `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt  time.Time     `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
