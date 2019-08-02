DROP DATABASE IF EXISTS `isutrain`;
CREATE DATABASE `isutrain`;

DROP USER IF EXISTS 'isutrain'@'localhost';
CREATE USER 'isutrain'@'localhost' IDENTIFIED BY 'isutrain';
GRANT ALL PRIVILEGES ON `isutrain`.* TO 'isutrain'@'localhost';

DROP USER IF EXISTS 'isutrain'@'%';
CREATE USER 'isutrain'@'%' IDENTIFIED BY 'isutrain';
GRANT ALL PRIVILEGES ON `isutrain`.* TO 'isutrain'@'%';