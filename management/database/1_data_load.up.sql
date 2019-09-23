-- MySQL dump 10.13  Distrib 8.0.16, for macos10.14 (x86_64)
--
-- Host: localhost    Database: restaurant
-- ------------------------------------------------------
-- Server version	8.0.16

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
 SET NAMES utf8mb4 ;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `admins`
--

DROP TABLE IF EXISTS `admins`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `admins` (
  `id` varchar(50) NOT NULL,
  `email_id` varchar(30) NOT NULL,
  `Name` varchar(25) NOT NULL,
  `password` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email_id` (`email_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `admins`
--

LOCK TABLES `admins` WRITE;
/*!40000 ALTER TABLE `admins` DISABLE KEYS */;
INSERT INTO `admins` VALUES ('268ae631-8ab6-4be6-bf4a-ac0414e15774','admin1@gmail.com','admin1','$2a$10$lIEk/TjgMVeZwLWmpEJ.M.XyyO.ZvK70fBParBXGxLTV4R.zT8GGC'),('2e4076de-f932-41b5-b18b-17f0f7234f14','admin5@gmail.com','admin5','$2a$10$iGODKka6TABPEmkJyW40SO/Efls1WOn7v9XHZBcWL2opyQrzgbJzm'),('6f05da4f-462f-4e8b-b352-1fff2d63ad38','admin7@gmail.com','admin7','$2a$10$M74zGnQvA6uSK2gGd6NQMuPPZNqdjWVW0IkjD30qFWdxHnZAqn5uy'),('7e1e78c6-a4c8-4483-a98a-c613347b310f','admin2@gmail.com','admin2','$2a$10$.tR7oXXKr78BXQ/EpzkDS.5ty8hHSwO6sC/XEjU2HzwTJ88geOChG'),('c758d0a0-7922-40d9-8a92-e8da21d4c0ce','admin6@gmail.com','admin6','$2a$10$BqSX9.oQLvwjW28E3o4Ii.IjRknSMFGPjuC8.8TtyyJptN/TiGzuy'),('d1110719-a2e2-42bc-9a8d-8f693dbea262','admin3@gmail.com','admin3','$2a$10$pNTMeF6Yvws8XeF8J6ASAucK7bev.nTgmTZzf/wrOmL4SfCevYFEC'),('d45556d0-72b5-437a-9812-aa1e2531f0b7','admin4@gmail.com','admin4','$2a$10$TnRbxUOF6kD4kAq0a4kIR.S2oN4eNOcjzEGJoQOHjCg20SpwlLeLS');
/*!40000 ALTER TABLE `admins` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `dishes`
--

DROP TABLE IF EXISTS `dishes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `dishes` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(30) NOT NULL,
  `price` float(7,2) NOT NULL,
  `res_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_restaurant` (`res_id`),
  CONSTRAINT `fk_restaurant` FOREIGN KEY (`res_id`) REFERENCES `restaurants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=55 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `dishes`
--

LOCK TABLES `dishes` WRITE;
/*!40000 ALTER TABLE `dishes` DISABLE KEYS */;
INSERT INTO `dishes` VALUES (9,'Dish5',104.10,12),(14,'Dish8',102.10,13),(15,'Dish9',103.10,13),(20,'',100.10,13),(21,'Dish7',101.10,13),(22,'Dish8',102.10,13),(23,'Dish9',103.10,13),(24,'Dish10',104.10,13),(25,'',100.10,13),(26,'Dish7',101.10,13),(27,'Dish8',102.10,13),(28,'Dish9',103.10,13),(29,'Dish10',104.10,13),(30,'',100.10,13),(31,'Dish7',101.10,13),(32,'Dish8',102.10,13),(37,'Dish8',102.10,13),(38,'Dish9',103.10,13),(39,'Dish10',104.10,13),(40,'',100.10,13),(41,'Dish7',101.10,13),(42,'Dish8',102.10,13),(43,'Dish9',103.10,13),(44,'Dish10',104.10,13),(45,'',100.10,13),(46,'Dish7',101.10,13),(47,'Dish8',102.10,13),(48,'Dish9',103.10,13),(49,'Dish10',104.10,13),(50,'',100.10,13),(51,'Dish7',101.10,13),(52,'Dish8',102.10,13),(53,'Dish9',103.10,13),(54,'Dish10',104.10,13);
/*!40000 ALTER TABLE `dishes` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `invalid_tokens`
--

DROP TABLE IF EXISTS `invalid_tokens`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `invalid_tokens` (
  `token` varchar(200) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `invalid_tokens`
--

LOCK TABLES `invalid_tokens` WRITE;
/*!40000 ALTER TABLE `invalid_tokens` DISABLE KEYS */;
INSERT INTO `invalid_tokens` VALUES ('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6ImY0NjhkOTkyLWM0NjUtNDJiZC04MDVhLTM5ODU3YWNkYmVjMyIsIlJvbGUiOiJzdXBlckFkbWluIiwiZXhwIjoxNTY1MzYzNDc0fQ.4pbJD35rrOqtqhukLbh-8T_DWqHgKPbpX6_5oAz0QCg'),('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6ImY0NjhkOTkyLWM0NjUtNDJiZC04MDVhLTM5ODU3YWNkYmVjMyIsIlJvbGUiOiJzdXBlckFkbWluIiwiZXhwIjoxNTY1MzYzNDc0fQ.4pbJD35rrOqtqhukLbh-8T_DWqHgKPbpX6_5oAz0QCg'),('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6ImY0NjhkOTkyLWM0NjUtNDJiZC04MDVhLTM5ODU3YWNkYmVjMyIsIlJvbGUiOiJzdXBlckFkbWluIiwiZXhwIjoxNTY1MzYzNDc0fQ.4pbJD35rrOqtqhukLbh-8T_DWqHgKPbpX6_5oAz0QCg'),('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6IjJlNDA3NmRlLWY5MzItNDFiNS1iMThiLTE3ZjBmNzIzNGYxNCIsIlJvbGUiOiJhZG1pbiIsImV4cCI6MTU2NTY4NjE1NX0.DjL0Ri0OXtfBOILgVs5uIGtsNBELrD-SK_LxZsNh_EQ'),('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6IjJlNDA3NmRlLWY5MzItNDFiNS1iMThiLTE3ZjBmNzIzNGYxNCIsIlJvbGUiOiJhZG1pbiIsImV4cCI6MTU2NTY5ODY4Nn0.IGzX7l2O6rcppFkBcwjLOGWbpCFeMfkeSIltsUekjDY'),('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6IjJlNDA3NmRlLWY5MzItNDFiNS1iMThiLTE3ZjBmNzIzNGYxNCIsIlJvbGUiOiJhZG1pbiIsImV4cCI6MTU2NTcxMTIzNn0.lELe3ezhjWFqIyQrERP-FVkcwnbBKfvGACsAQUPAtco'),('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6ImY1NGEwNzhiLWI0M2QtNDliNi1iZTFiLWU3ZGYzNGQ5YWJlNSIsIlJvbGUiOiJzdXBlckFkbWluIiwiZXhwIjoxNTY1NzEyMzkyfQ.T-iFuT0yjQqevycWGjJjA5_y9lnWCAWshndl3whIC4o'),('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6ImY1NGEwNzhiLWI0M2QtNDliNi1iZTFiLWU3ZGYzNGQ5YWJlNSIsIlJvbGUiOiJzdXBlckFkbWluIiwiZXhwIjoxNTY1NzE0MjAyfQ.cl8KM_-v89XaSSk_gCEkk8Y4kJXlErXhL5CJ6_j2dwI'),('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6ImY1NGEwNzhiLWI0M2QtNDliNi1iZTFiLWU3ZGYzNGQ5YWJlNSIsIlJvbGUiOiJzdXBlckFkbWluIiwiZXhwIjoxNTY1NzE0Mjk3fQ.TXoYtPxPgSC7l4rAvdTB2B14kOl03aEFqWuwHwdH__w');
/*!40000 ALTER TABLE `invalid_tokens` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `owners`
--

DROP TABLE IF EXISTS `owners`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `owners` (
  `id` varchar(50) NOT NULL,
  `email_id` varchar(30) DEFAULT NULL,
  `name` varchar(25) NOT NULL,
  `password` varchar(100) NOT NULL,
  `creator_id` varchar(50) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email_id` (`email_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `owners`
--

LOCK TABLES `owners` WRITE;
/*!40000 ALTER TABLE `owners` DISABLE KEYS */;
INSERT INTO `owners` VALUES ('0bb6ab77-0d25-40ba-a989-c8ee31e10862','Verynewowner@gmail. com','VeryNew Owner','$2a$10$PtyerDTNzzXRzKil22SEYOD1soJ1jcJivbeGLmQJAvcgqOkztfLR.','f54a078b-b43d-49b6-be1b-e7df34d9abe5'),('155914aa-1ae9-4db3-8479-3acb725835cc','Owner7@gmail.com','Ownerseven','$2a$10$EGvIfdceG8w5B4NZjB/ifOPItVe5GCdUsE.vhNwPVM2yfuy7u8rWu','6f05da4f-462f-4e8b-b352-1fff2d63ad38'),('2df89e8f-43fa-424e-b929-12193dc0a559','owner6@gmail.com','owner6','$2a$10$V0x/aDnIgzHVSmTUkM5sEeI9mG8dElTUySmjVYSQlgeEYRjzEQb0W','c758d0a0-7922-40d9-8a92-e8da21d4c0ce'),('6b64417e-ac98-4f2d-a44f-de10b2e533be','owner5@gmail.com','owner5','$2a$10$4wF4mUap7ulZRqdgF5m/xOnX/iQnDDeYLxLTeHHoozxOjWE96sbvu','c758d0a0-7922-40d9-8a92-e8da21d4c0ce'),('7015d53a-1d26-4306-971d-051930bbb34b','owner114@gmail.com','owner14','$2a$10$JPpYbF3UT6IejPOIOCK.OufB/RuTacLaZdAoq6wXOgJcKAbu1OwKa','2e4076de-f932-41b5-b18b-17f0f7234f14'),('7bd2ced9-2707-4602-a794-2623bca93083','owner8@gmail.com','owner8','$2a$10$rjZ1mq0AfzT6VIePJOOa2OWxPeTJcCLUgcvL.exD6E3obptxzhYPa','6f05da4f-462f-4e8b-b352-1fff2d63ad38'),('80e756dc-2f5b-4435-bf92-f7d8faaebfef','owner12@gmail.com','owner12','$2a$10$tRR4qd4W8cXsKOP.snj2zeO79cXekCgSqh5FayviS0P/JZ5e.FteS','d079e1da-f224-4bdf-8dcb-fb92ebb75f1a'),('904587f0-9f28-4a82-9694-274bfae0b323','owner11@gmail.com','owner11','$2a$10$nbmcflBZolsKyP9QX5ArCOtn34nHh2BXhcgAyFpeFwT/SAkm.Kaxq','f468d992-c465-42bd-805a-39857acdbec3'),('c0b8bdae-82b9-4f5a-91bf-bd7cc6210e12','owner4@gmail.com','owner4','$2a$10$TQce0o54T7HHM8fBZ4h61.sZFwlox/oeywJb/GuAFGg0.gCmZIX5i','c758d0a0-7922-40d9-8a92-e8da21d4c0ce'),('f2e2580a-72e9-4ee3-adb3-d728ea7ad68d','owner9@gmail.com','owner9','$2a$10$OS43oEPwgL2FlqnaP7NZzOQfpj2NX1ICwDRQjESZ28dU2rHVpE8zq','6f05da4f-462f-4e8b-b352-1fff2d63ad38');
/*!40000 ALTER TABLE `owners` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `restaurants`
--

DROP TABLE IF EXISTS `restaurants`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `restaurants` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL,
  `lat` float(10,6) NOT NULL,
  `lng` float(10,6) NOT NULL,
  `creator_id` varchar(50) DEFAULT NULL,
  `owner_id` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=46 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `restaurants`
--

LOCK TABLES `restaurants` WRITE;
/*!40000 ALTER TABLE `restaurants` DISABLE KEYS */;
INSERT INTO `restaurants` VALUES (12,'res1',9.000000,11.000000,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(13,'res2',5.000000,4.000000,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(14,'res3',3.000000,9.000000,'2e4076de-f932-41b5-b18b-17f0f7234f14','c0b8bdae-82b9-4f5a-91bf-bd7cc6210e12'),(15,'res4',10.000000,11.000000,'c758d0a0-7922-40d9-8a92-e8da21d4c0ce','c0b8bdae-82b9-4f5a-91bf-bd7cc6210e12'),(16,'res5',12.000000,14.000000,'c758d0a0-7922-40d9-8a92-e8da21d4c0ce','c0b8bdae-82b9-4f5a-91bf-bd7cc6210e12'),(17,'res6',15.000000,16.000000,'c758d0a0-7922-40d9-8a92-e8da21d4c0ce','c0b8bdae-82b9-4f5a-91bf-bd7cc6210e12'),(18,'res7',17.000000,18.000000,'6f05da4f-462f-4e8b-b352-1fff2d63ad38','c0b8bdae-82b9-4f5a-91bf-bd7cc6210e12'),(19,'res8',20.000000,21.000000,'6f05da4f-462f-4e8b-b352-1fff2d63ad38','c0b8bdae-82b9-4f5a-91bf-bd7cc6210e12'),(20,'res9',22.000000,24.000000,'6f05da4f-462f-4e8b-b352-1fff2d63ad38','c0b8bdae-82b9-4f5a-91bf-bd7cc6210e12'),(21,'res10',23.000000,25.000000,'d079e1da-f224-4bdf-8dcb-fb92ebb75f1a','c0b8bdae-82b9-4f5a-91bf-bd7cc6210e12'),(22,'res11',25.000000,27.000000,'d079e1da-f224-4bdf-8dcb-fb92ebb75f1a',NULL),(23,'res12',27.000000,29.000000,'f468d992-c465-42bd-805a-39857acdbec3',NULL),(24,'res13',28.000000,30.000000,'f468d992-c465-42bd-805a-39857acdbec3',NULL),(25,'res50',28.000000,30.000000,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(26,'Baba Hotel',24.500000,30.700001,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(27,'Chacha Chicken',30.500000,24.799999,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(28,'Rasooi',10.400000,35.599998,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(29,'Res11',43.900002,23.500000,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(30,'Res12',15.600000,25.799999,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(31,'Res13',10.500000,24.600000,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(32,'Res5',12.900000,23.500000,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(33,'Res34',19.000000,12.500000,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(34,'Res50',17.799999,24.400000,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(35,'Res10',1.000000,2.000000,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(36,'Res36',34.500000,24.500000,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(37,'Res45',34.599998,34.599998,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(38,'Res30',23.000000,45.000000,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(39,'Res30',23.000000,45.000000,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(40,'Res100',23.400000,23.500000,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(41,'Nikhildhaba',23.000000,34.000000,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(42,'Anshudhaba',23.400000,25.700001,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(43,'Res90',34.599998,78.599998,'2e4076de-f932-41b5-b18b-17f0f7234f14',NULL),(44,'Hotcocoa',33.439999,33.439999,'f54a078b-b43d-49b6-be1b-e7df34d9abe5',NULL),(45,'Food with Fire',34.000000,81.900002,'f54a078b-b43d-49b6-be1b-e7df34d9abe5',NULL);
/*!40000 ALTER TABLE `restaurants` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `super_admins`
--

DROP TABLE IF EXISTS `super_admins`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `super_admins` (
  `id` varchar(50) NOT NULL,
  `email_id` varchar(30) NOT NULL,
  `Name` varchar(25) NOT NULL,
  `password` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email_id` (`email_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `super_admins`
--

LOCK TABLES `super_admins` WRITE;
/*!40000 ALTER TABLE `super_admins` DISABLE KEYS */;
INSERT INTO `super_admins` VALUES ('072bd9cd-646c-4384-bf24-32b45de406f6','superAdmin5@gmail.com','superAdmin5','$2a$10$zgoYPsH7OJmWUZ4C20yg2ubHIXJ9KU6ko/uALK3jB27FNk44FIzzy'),('97790ad8-1c28-4868-8f3d-e7cd67e7ce33','superAdmin2@gmail.com','superAdmin2','$2a$10$UPl00qprVQf/Z/aFn8vMbOzIKbRMYGHZveWijHhVJfG2/1ty4EenO'),('d079e1da-f224-4bdf-8dcb-fb92ebb75f1a','superAdmin4@gmail.com','superAdmin4','$2a$10$rSIq6v89YZXPBCjEbTStjuGzvZkdJkYQXKygSxEUEo.ObEpiEzGGu'),('f468d992-c465-42bd-805a-39857acdbec3','superAdmin3@gmail.com','superAdmin3','$2a$10$VAo9KjjuHMew66Sh/FgSQ.DPNpj.obz5gxEdUsFS9U9CqVtsb6iLW'),('f54a078b-b43d-49b6-be1b-e7df34d9abe5','superAdmin1@gmail.com','superAdmin1','$2a$10$NrdDy/WsUfOzHwz4UqY8lu4vj90qo6flccDAM1bGkLEjJNCqAKhnu');
/*!40000 ALTER TABLE `super_admins` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2019-08-14 11:51:09
