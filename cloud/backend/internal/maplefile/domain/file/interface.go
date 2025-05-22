// cloud/backend/internal/maplefile/domain/file/interface.go
package file

import "go.mongodb.org/mongo-driver/bson/primitive"

type FileRepository interface {
	Create(file *File) error
	Get(id primitive.ObjectID) (*File, error)
	GetByCollection(collectionID string) ([]*File, error)
	Update(file *File) error
	Delete(id primitive.ObjectID) error
	StoreEncryptedData(fileID primitive.ObjectID, encryptedData []byte) error
	GetEncryptedData(fileID primitive.ObjectID) ([]byte, error)
	CreateMany(files []*File) error
	DeleteMany(ids []primitive.ObjectID) error
	GetMany(ids []primitive.ObjectID) ([]*File, error)
}
