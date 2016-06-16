ALTER TABLE user_service_list ADD FOREIGN KEY (user_service_id) REFERENCES user_service(id);
ALTER TABLE user_service_addon_list ADD FOREIGN KEY (user_service_id) REFERENCES user_service(id);

ALTER TABLE service ADD COLUMN estimated_time_fleet INT NOT NULL DEFAULT 0;

DROP TABLE IF EXISTS fleet_service_addon_list_id;
DROP TABLE IF EXISTS fleet_service_addon_list;
DROP TABLE IF EXISTS fleet_service_list_id;
DROP TABLE IF EXISTS fleet_service_list;
DROP TABLE IF EXISTS fleet_history;
DROP TABLE IF EXISTS fleet_service;
DROP TABLE IF EXISTS fleet;
DROP TRIGGER IF EXISTS INSERT_FLEET_TOKEN;
DROP TRIGGER IF EXISTS INSERT_FLEET_RESERVATIOM_ID;
DELETE FROM user WHERE type in ('FLEET', 'SALE');

CREATE TABLE fleet (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    token VARCHAR(5) NOT NULL DEFAULT '',
    name VARCHAR(128) NOT NULL DEFAULT '',
    setup INT NOT NULL DEFAULT 0,
    user_id INT NOT NULL,
    sale_user_id INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(id),
    FOREIGN KEY (sale_user_id) REFERENCES user(id)
);

CREATE TABLE fleet_service (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    reservation_id VARCHAR(8) NOT NULL DEFAULT '',
    user_id INT NOT NULL,
    report_id INT NULL,
    gap INT NOT NULL DEFAULT 0,
    types VARCHAR(32) NOT NULL,
    estimated_time INT NOT NULL,
    estimated_price FLOAT NOT NULL DEFAULT 0.0,
    note VARCHAR(2048) NOT NULL DEFAULT '',
    status ENUM('WAITING', 'RESERVED', 'IN_PROGRESS', 'DONE', 'CANCEL'),
    opening_id INT NOT NULL,
    assignee INT NOT NULL DEFAULT -1,
    reserved_start_timestamp TIMESTAMP NULL,
    start_timestamp TIMESTAMP NULL,
    end_timestamp TIMESTAMP NULL,
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id),
    FOREIGN KEY (opening_id) REFERENCES opening(id),
    INDEX(status)
);

CREATE TABLE fleet_service_list (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    fleet_service_id INT NOT NULL,
    order_id INT NOT NULL,
    car_count INT NOT NULL,
    FOREIGN KEY (fleet_service_id) REFERENCES fleet_service(id)
);

CREATE TABLE fleet_service_list_id (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    service_id INT NOT NULL,
    fleet_service_list_id INT NOT NULL,
    FOREIGN KEY (service_id) REFERENCES service(id),
    FOREIGN KEY (fleet_service_list_id) REFERENCES fleet_service_list(id)
);

CREATE TABLE fleet_service_addon_list (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    fleet_service_id INT NOT NULL,
    order_id INT NOT NULL,
    car_count INT NOT NULL,
    FOREIGN KEY (fleet_service_id) REFERENCES fleet_service(id)
);

CREATE TABLE fleet_service_addon_list_id (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    service_addon_id INT NOT NULL,
    fleet_service_addon_list_id INT NOT NULL,
    amount INT NOT NULL,
    FOREIGN KEY (service_addon_id) REFERENCES service_addon(id),
    FOREIGN KEY (fleet_service_addon_list_id) REFERENCES fleet_service_addon_list(id)
);

CREATE TABLE fleet_history (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    rating FLOAT NOT NULL DEFAULT 0,
    fleet_service_id INT NOT NULL,
    report_id INT NULL,
    note VARCHAR(2048) NOT NULL DEFAULT '',
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (fleet_service_id) REFERENCES fleet_service(id)
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

DELIMITER $$
CREATE TRIGGER INSERT_FLEET_RESERVATIOM_ID BEFORE INSERT ON fleet_service FOR EACH ROW
BEGIN
    DECLARE id INT DEFAULT 0;

    SELECT AUTO_INCREMENT INTO id FROM information_schema.tables
    WHERE TABLE_NAME = 'fleet_service' and TABLE_SCHEMA = database();

    SET NEW.reservation_id = UPPER(SUBSTRING(SHA2(id, 256), 32, 8));
END $$
DELIMITER ;

ALTER TABLE user CHANGE type type ENUM('RESIDENTIAL', 'BUSINESS', 'EGOBIE', 'SALE', 'FLEET');

INSERT INTO user (type, username, password, email, phone_number) VALUES
('SALE', 'bd1', 'bc254388680ed7c7e426b417e81f41b6af7ef319', 'bdsale1@egobie.com', '2019120383');

INSERT INTO service_addon (service_id, name, note, price, time) VALUES
(0, "Detailed Shampoo", "Seating & Mats & Carpets", 60, 60),
(0, "Engine Cleaning", "", 50, 30),
(0, "Hand Wax", "", 35, 60),
(0, "Headlight Reconditioning", "", 65, 60),
(0, "Hot Carpet Extraction", "", 15, 30),
(0, "Paint Protectant", "Multi-layer", 50, 60),
(0, "Wax & Polish", "Multi-layer", 75, 60),
(0, "Change Cabine Air Filter", "", 45, 0),
(0, "Change Engine Air Filter",  "", 45, 0),
(0, "Change Serpentine Belts",  "", 150, 0);

INSERT INTO service_addon (service_id, name, note, price, max, unit) VALUES
(0, 'Extra Conventional Oil', 'per quart', 4, 30, 'quart'),
(0, 'Extra Synthetic Oil', 'per quart', 8, 30, 'quart');