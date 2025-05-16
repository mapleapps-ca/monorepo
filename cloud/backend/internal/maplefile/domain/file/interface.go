package file

import "go.mongodb.org/mongo-driver/bson/primitive"

type FileRepository interface {
	Create(file *File) error
	Get(id primitive.ObjectID) (*File, error)
	GetByCollection(collectionID string) ([]*File, error)
	Update(file *File) error
	Delete(id primitive.ObjectID) error
	StoreEncryptedData(fileID string, encryptedData []byte) error
	GetEncryptedData(fileID string) ([]byte, error)
	CreateMany(files []*File) error
	DeleteMany(ids []primitive.ObjectID) error
	GetByEncryptedFileID(encryptedFileID string) (*File, error)
	GetMany(ids []primitive.ObjectID) ([]*File, error)
}
