package file

type FileRepository interface {
	Create(file *File) error
	Get(id string) (*File, error)
	GetByCollection(collectionID string) ([]*File, error)
	Update(file *File) error
	Delete(id string) error
	StoreEncryptedData(fileID string, encryptedData []byte) error
	GetEncryptedData(fileID string) ([]byte, error)
}
