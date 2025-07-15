// cloud/backend/internal/maplefile/domain/collection/constants.go
package collection

const (
	CollectionTypeFolder = "folder"
	CollectionTypeAlbum  = "album"
)

const ( // Permission levels
	CollectionPermissionReadOnly  = "read_only"
	CollectionPermissionReadWrite = "read_write"
	CollectionPermissionAdmin     = "admin"
)

const (
	CollectionStateActive   = "active"
	CollectionStateDeleted  = "deleted"
	CollectionStateArchived = "archived"
)

const (
	CollectionAccessTypeOwner  = "owner"
	CollectionAccessTypeMember = "member"
)
