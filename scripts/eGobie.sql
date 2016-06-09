DROP DATABASE IF EXISTS egobie;

CREATE DATABASE egobie;

USE egobie;

-- SET GLOBAL time_zone = '-04:00';

-- SET GLOBAL log_bin_trust_function_creators = 1;

CREATE TABLE user (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    type ENUM('RESIDENTIAL', 'BUSINESS', 'EGOBIE', 'SALE', 'FLEET'),
    first_name VARCHAR(32) NULL DEFAULT '',
    last_name VARCHAR(32) NULL DEFAULT '',
    middle_name VARCHAR(32) NULL DEFAULT '',
    username VARCHAR(16) NOT NULL DEFAULT '' UNIQUE KEY,
    email VARCHAR(64) NOT NULL DEFAULT '' UNIQUE KEY,
    phone_number VARCHAR(10) NOT NULL DEFAULT '',
    password VARCHAR(128) NOT NULL,
    home_address_state VARCHAR(2) NULL DEFAULT '',
    home_address_zip VARCHAR(8) NULL DEFAULT '',
    home_address_city VARCHAR(32) NULL DEFAULT '',
    home_address_street VARCHAR(128) NULL DEFAULT '',
    work_address_state VARCHAR(2) NULL DEFAULT '',
    work_address_zip VARCHAR(8) NULL DEFAULT '',
    work_address_city VARCHAR(32) NULL DEFAULT '',
    work_address_street VARCHAR(128) NULL DEFAULT '',
    sign INT NOT NULL DEFAULT 0,
    coupon VARCHAR(5) NOT NULL DEFAULT '',
    referred VARCHAR(5) NOT NULL DEFAULT '',
    discount INT NOT NULL DEFAULT 0,
    invitation INT NOT NULL DEFAULT 0,
    sign_up DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sign_in DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX(type),
    INDEX(coupon),
    INDEX(home_address_state),
    INDEX(home_address_zip),
    INDEX(home_address_city)
);

CREATE TABLE user_action (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    action VARCHAR(512) NOT NULL,
    data VARCHAR(128) NULL,
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE service (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    type ENUM('CAR_WASH', 'OIL_CHANGE', 'DETAILING', 'REPAIR'),
    items TEXT NOT NULL,
    description VARCHAR(1024) NOT NULL,
    note VARCHAR(128) NOT NULL DEFAULT '',
    estimated_price FLOAT NOT NULL,
    estimated_time INT NOT NULL,
    demand INT NOT NULL DEFAULT 0,
    reading INT NOT NULL DEFAULT 0,
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE service_addon (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    service_id INT NOT NULL,
    name VARCHAR(128) NOT NULL,
    note VARCHAR(128) NOT NULL DEFAULT "",
    price FLOAT NOT NULL,
    time INT NOT NULL DEFAULT 0,
    max INT NOT NULL DEFAULT 1,
    unit VARCHAR(32) NOT NULL DEFAULT '',
    demand INT NOT NULL DEFAULT 0
);

CREATE TABLE report (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    content VARCHAR(1024) NOT NULL,
    type ENUM('SERVICE', 'HISTORY', 'NOTIFICATION', 'CAR'),
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE car_maker (
  id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(64) NOT NULL,
  title VARCHAR(64) NOT NULL
);

CREATE TABLE IF NOT EXISTS car_model (
  id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  car_maker_id INT NOT NULL,
  name VARCHAR(128) NOT NULL,
  title VARCHAR(128) NOT NULL,
  FOREIGN KEY (car_maker_id) REFERENCES car_maker(id)
);

-- 8.am - 20.pm
CREATE TABLE opening (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    day DATE NOT NULL,
    period INT NOT NULL,
    count_wash INT NOT NULL DEFAULT 1,
    count_oil INT NOT NULL DEFAULT 1,
    demand INT NOT NULL DEFAULT 0,
    UNIQUE KEY (day, period)
);

CREATE TABLE user_car (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    report_id INT NULL,
    plate VARCHAR(16) NOT NULL UNIQUE KEY,
    state VARCHAR(2) NOT NULL,
    year INT NOT NULL,
    color ENUM('WHITE', 'BLACK', 'SILVER', 'GRAY', 'RED', 'BLUE', 'BROWN', 'YELLOW', 'GOLD', 'GREEN', 'PINK', 'OTHERS') NOT NULL,
    car_maker_id INT NOT NULL,
    car_model_id INT NOT NULL,
    reserved INT NOT NULL DEFAULT 0,
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id),
    FOREIGN KEY (car_maker_id) REFERENCES car_maker(id),
    FOREIGN KEY (car_model_id) REFERENCES car_model(id),
    INDEX(year),
    INDEX(state),
    INDEX(car_maker_id),
    INDEX(car_model_id)
);

CREATE TABLE user_payment (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    account_name VARCHAR(64) NOT NULL,
    account_number VARCHAR(128) NOT NULL UNIQUE KEY,
    account_zip VARCHAR(16) NOT NULL,
    account_type ENUM('CREDIT', 'DEBIT'),
    card_type VARCHAR(32) NOT NULL DEFAULT 'Visa',
    code VARCHAR(128) NULL,
    expire_month VARCHAR(2) NOT NULL,
    expire_year VARCHAR(4) NOT NULL,
    reserved INT NOT NULL DEFAULT 0,
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE TABLE user_service (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    reservation_id VARCHAR(8) NOT NULL DEFAULT '',
    user_id INT NOT NULL,
    user_car_id INT NOT NULL,
    user_payment_id INT NOT NULL,
    report_id INT NULL,
    gap INT NOT NULL DEFAULT 0,
    types VARCHAR(32) NOT NULL,
    estimated_time INT NOT NULL,
    estimated_price FLOAT NOT NULL,
    note VARCHAR(2048) NOT NULL DEFAULT '',
    status ENUM('RESERVED', 'IN_PROGRESS', 'DONE', 'CANCEL'),
    opening_id INT NOT NULL,
    assignee INT NOT NULL DEFAULT -1,
    reserved_start_timestamp TIMESTAMP NULL,
    start_timestamp TIMESTAMP NULL,
    end_timestamp TIMESTAMP NULL,
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id),
    INDEX(status)
);

CREATE TABLE user_service_list (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    service_id INT NOT NULL,
    user_service_id INT NOT NULL,
    FOREIGN KEY (service_id) REFERENCES service(id)
);

CREATE TABLE user_service_addon_list (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    service_addon_id INT NOT NULL,
    user_service_id INT NOT NULL,
    amount INT NOT NULL,
    FOREIGN KEY (service_addon_id) REFERENCES service_addon(id)
);

CREATE TABLE user_history (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    rating FLOAT NOT NULL DEFAULT 0,
    user_id INT NOT NULL,
    user_service_id INT NOT NULL,
    car_plate VARCHAR(16) NOT NULL,
    car_state VARCHAR(2) NOT NULL,
    car_maker VARCHAR(32) NOT NULL,
    car_model VARCHAR(64) NOT NULL,
    car_year INT NOT NULL,
    car_color VARCHAR(16) NOT NULL,
    payment_holder VARCHAR(64) NOT NULL,
    payment_number VARCHAR(128) NOT NULL,
    payment_type ENUM('CREDIT', 'DEBIT') NOT NULL,
    payment_price FLOAT NOT NULL,
    report_id INT NULL,
    note VARCHAR(2048) NOT NULL DEFAULT '',
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id),
    INDEX(user_id)
);

CREATE TABLE user_notification (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    content TEXT NOT NULL,
    type ENUM('MEMO'),
    FOREIGN KEY (user_id) REFERENCES user(id),
    INDEX(user_id)
);

CREATE TABLE user_feedback (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR(128) NOT NULL,
    feedback TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(id),
    INDEX(user_id)
);

CREATE TABLE user_opening (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    day DATE NOT NULL,
    user_id INT NOT NULL,
    user_schedule INT NOT NULL DEFAULT 16777215,
    mixed INT NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES user(id),
    UNIQUE KEY (day, user_id),
    INDEX(user_id)
);

