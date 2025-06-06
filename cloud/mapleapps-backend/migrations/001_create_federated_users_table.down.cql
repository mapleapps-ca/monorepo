-- Drop indexes first
DROP INDEX IF EXISTS mapleapps.federated_users_email_idx;
DROP INDEX IF EXISTS mapleapps.federated_users_verification_code_idx;
DROP INDEX IF EXISTS mapleapps.federated_users_status_idx;
DROP INDEX IF EXISTS mapleapps.federated_users_role_idx;
DROP INDEX IF EXISTS mapleapps.federated_users_created_at_idx;

-- Drop the table
DROP TABLE IF EXISTS mapleapps.federated_users;
