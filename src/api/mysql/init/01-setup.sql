-- Create the database if it doesn't exist
CREATE DATABASE IF NOT EXISTS fedhathbt CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE fedhathbt;

-- Create function for UUID generation (if using UUID as primary keys)
DELIMITER $$
CREATE FUNCTION IF NOT EXISTS uuid_to_bin(_uuid CHAR(36))
RETURNS BINARY(16)
DETERMINISTIC
BEGIN
    RETURN UNHEX(REPLACE(_uuid, '-', ''));
END $$

CREATE FUNCTION IF NOT EXISTS bin_to_uuid(_bin BINARY(16))
RETURNS CHAR(36)
DETERMINISTIC
BEGIN
    DECLARE _hex CHAR(32);
    SET _hex = HEX(_bin);
    RETURN CONCAT(
        SUBSTR(_hex, 1, 8), '-',
        SUBSTR(_hex, 9, 4), '-',
        SUBSTR(_hex, 13, 4), '-',
        SUBSTR(_hex, 17, 4), '-',
        SUBSTR(_hex, 21)
    );
END $$
DELIMITER ;

-- Create trigger for timestamps
DELIMITER $$
CREATE TRIGGER IF NOT EXISTS set_timestamps_before_insert
BEFORE INSERT ON users FOR EACH ROW
BEGIN
    SET NEW.created_at = IFNULL(NEW.created_at, NOW());
    SET NEW.updated_at = IFNULL(NEW.updated_at, NOW());
END $$

CREATE TRIGGER IF NOT EXISTS set_timestamps_before_update
BEFORE UPDATE ON users FOR EACH ROW
BEGIN
    SET NEW.updated_at = NOW();
END $$
DELIMITER ;

-- Create admin user if needed
-- INSERT INTO users (id, name, email, password_hash, role, is_active, email_verified, created_at, updated_at)
-- VALUES (UUID_TO_BIN(UUID()), 'Admin User', 'admin@example.com', '$2a$10$JvXMqKLZrGKPBn5eCZfZOOliPRLa5GTQNj1YdkLFALKSNEsyjnLYO', 'ADMIN', 1, 1, NOW(), NOW()); 