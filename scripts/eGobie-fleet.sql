CREATE TABLE fleet (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    token VARCHAR(5) NOT NULL DEFAULT '',
    name VARCHAR(128) NOT NULL DEFAULT '',
    user_id INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE TABLE fleet_price (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    price FLOAT NOT NULL,
    type ENUM('WASH', 'OIL', 'WASH_OIL'),
    fleet_id INT NOT NULL,
    FOREIGN KEY (fleet_id) REFERENCES fleet(id)
);

DELIMITER $$
CREATE TRIGGER INSERT_FLEET_TOKEN BEFORE INSERT ON fleet FOR EACH ROW
BEGIN
    DECLARE id INT DEFAULT 0;

    SELECT AUTO_INCREMENT INTO id FROM information_schema.tables
    WHERE TABLE_NAME = 'fleet' and TABLE_SCHEMA = database();

    SET NEW.token = UPPER(SUBSTRING(SHA2(id, 256), 32, 5));
END $$
DELIMITER ;
