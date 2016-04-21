DROP DATABASE IF EXISTS egobie;

CREATE DATABASE egobie;

USE egobie;

CREATE TABLE user (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    type ENUM('RESIDENTIAL', 'BUSINESS', 'RUNNER'),
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
    sign_up DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sign_in DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX(type),
    INDEX(home_address_state),
    INDEX(home_address_zip),
    INDEX(home_address_city)
);

CREATE TABLE service (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    type ENUM('CAR_WASH', 'OIL_CHANGE', 'DETAILING', 'REPAIR'),
    items TEXT NOT NULL,
    description VARCHAR(1024) NOT NULL,
    estimated_price FLOAT NOT NULL,
    estimated_time INT NOT NULL,
    addons INT(1) NOT NULL DEFAULT 0,
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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
-- 1=8-9, 2=9-10, 3=10-11, 4=11-12, 5=12-13, 6=13-14
-- 7=14-15, 8=15-16, 9=16-17, 10=17-18, 11=18-19, 12=19-20
CREATE TABLE opening (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    day DATE NOT NULL,
    period INT NOT NULL,
    count INT NOT NULL DEFAULT 2,
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
    account_type ENUM('CREDIT', 'DEBIT'),
    code VARCHAR(128) NULL,
    expire_month VARCHAR(2) NOT NULL,
    expire_year VARCHAR(4) NOT NULL,
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE TABLE user_service (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    user_car_id INT NOT NULL,
    user_payment_id INT NOT NULL,
    report_id INT NULL,
    estimated_time INT NOT NULL,
    estimated_price FLOAT NOT NULL,
    note VARCHAR(512) NULL,
    status ENUM('RESERVED', 'IN_PROGRESS', 'DONE'),
    opening_id INT NOT NULL,
    start_timestamp TIMESTAMP NULL,
    end_timestamp TIMESTAMP NULL,
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id),
    FOREIGN KEY (user_car_id) REFERENCES user_car(id),
    FOREIGN KEY (user_payment_id) REFERENCES user_payment(id),
    FOREIGN KEY (report_id) REFERENCES report(id),
    FOREIGN KEY (opening_id) REFERENCES opening(id)
);

CREATE TABLE user_service_list (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    service_id INT NOT NULL,
    user_service_id INT NOT NULL,
    FOREIGN KEY (service_id) REFERENCES service(id),
    FOREIGN KEY (user_service_id) REFERENCES user_service(id)
);

CREATE TABLE user_history (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    ratting FLOAT NOT NULL,
    user_id INT NOT NULL,
    user_service_id INT NOT NULL,
    report_id INT NULL,
    note VARCHAR(512) NULL,
    create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id),
    FOREIGN KEY (user_service_id) REFERENCES user_service(id),
    INDEX(user_id)
);

CREATE TABLE user_notification (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    content VARCHAR(512) NOT NULL,
    type ENUM('MEMO'),
    FOREIGN KEY (user_id) REFERENCES user(id),
    INDEX(user_id)
);


