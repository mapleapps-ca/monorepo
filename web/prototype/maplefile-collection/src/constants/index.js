export const COLLECTION_TYPES = {
  FOLDER: "folder",
  ALBUM: "album",
};

export const COLLECTION_STATES = {
  ACTIVE: "active",
  DELETED: "deleted",
  ARCHIVED: "archived",
};

export const PERMISSION_LEVELS = {
  READ_ONLY: "read_only",
  READ_WRITE: "read_write",
  ADMIN: "admin",
};

export const ROUTES = {
  HOME: "/",
  COLLECTIONS: "/collections",
  COLLECTION_DETAIL: "/collections/:id",
  CREATE_COLLECTION: "/collections/new",
  EDIT_COLLECTION: "/collections/:id/edit",
  SHARED_COLLECTIONS: "/shared",
};

export const PERMISSION_LABELS = {
  [PERMISSION_LEVELS.READ_ONLY]: "Read Only",
  [PERMISSION_LEVELS.READ_WRITE]: "Read & Write",
  [PERMISSION_LEVELS.ADMIN]: "Admin",
};

export const COLLECTION_TYPE_LABELS = {
  [COLLECTION_TYPES.FOLDER]: "Folder",
  [COLLECTION_TYPES.ALBUM]: "Album",
};
