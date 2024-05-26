-- create database bookreview
CREATE Database If NOT EXIST `bookreview`

-- review table

CREATE TABLE `reviews` (
  `id` int(11) PRIMARY KEY AUTO_INCREMENT NOT NULL,
  `isbn` varchar(50) NOT NULL,
  `title` varchar(255) NOT NULL,
  `author` varchar(255) NOT NULL,
  `genre` varchar(50) NOT NULL,
  `descriptions` text NOT NULL,
  `price` decimal(5,2) NOT NULL,
  `created_at` datetime DEFAULT current_timestamp(),
  `uid` int(11) NOT NULL,
  `rating` int(11) NOT NULL,
  `is_deleted` tinyint(1) DEFAULT 0
)

-- change user info

CREATE TABLE `users` (
  `id` int(11) NOT NULL,
  `email` varchar(100) NOT NULL,
  `password` varchar(100) NOT NULL,
  `created_at` datetime DEFAULT current_timestamp(),
  `token` varchar(100) DEFAULT NULL,
  `active` tinyint(1) DEFAULT 0,
  `verified` varchar(100) DEFAULT NULL
)

-- to forget password

CREATE TABLE `forget_passw` (
  `id` int(11) PRIMARY KEY AUTO_INCREMENT NOT NULL,
  `uid` int(11) NOT NULL,
  `uri` varchar(100) NOT NULL,
  `created_at` datetime DEFAULT current_timestamp(),
  `superseded` tinyint(1) DEFAULT 0
) 

-- to log user activity

CREATE TABLE `user_log` (
  `id` int(11) PRIMARY KEY AUTO_INCREMENT NOT NULL,
  `activity` varchar(50) NOT NULL,
  `uid` int(11) NOT NULL,
  `created_at` datetime DEFAULT current_timestamp(),
  `superseded` tinyint(1) DEFAULT 0
) 