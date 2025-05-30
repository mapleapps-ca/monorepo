// internal/domain/collection/constants.go
package collectionsharingdto

const (
	// Permission levels define the access rights users have to a collection.
	// These levels dictate what actions a user can perform within the collection (e.g., viewing, adding, deleting files/subcollections, managing members).

	// CollectionDTOPermissionReadOnly grants users the ability to view the contents of the collection (files and subcollections)
	// and their metadata, but not modify them or the collection itself.
	CollectionDTOPermissionReadOnly = "read_only"
	// CollectionDTOPermissionReadWrite grants users the ability to view, add, modify, and delete
	// files and subcollections within the collection. They cannot manage collection members or delete the collection itself.
	CollectionDTOPermissionReadWrite = "read_write"
	// CollectionDTOPermissionAdmin grants users full control over the collection, including
	// all read/write operations, managing collection members (sharing/unsharing), and deleting the collection.
	CollectionDTOPermissionAdmin = "admin"
)
