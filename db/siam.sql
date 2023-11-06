-- phpMyAdmin SQL Dump
-- version 5.1.1
-- https://www.phpmyadmin.net/
--
-- Host: localhost
-- Generation Time: Nov 06, 2023 at 06:40 PM
-- Server version: 10.4.21-MariaDB
-- PHP Version: 8.1.2

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `siam`
--

-- --------------------------------------------------------

--
-- Table structure for table `siam_login`
--

CREATE TABLE `siam_login` (
  `id` int(11) NOT NULL,
  `username` varchar(100) NOT NULL,
  `password` varchar(100) NOT NULL,
  `nama` varchar(100) NOT NULL,
  `role` varchar(100) DEFAULT NULL,
  `status` int(11) NOT NULL,
  `count_block` int(11) DEFAULT 0,
  `tgl_input` datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- Dumping data for table `siam_login`
--

INSERT INTO `siam_login` (`id`, `username`, `password`, `nama`, `role`, `status`, `count_block`, `tgl_input`) VALUES
(1, 'admin_siam', '7c467e39c8cf77223dc7a3a67150414b', 'Admin Sistem Akademik Methodist 1', 'ADMINISTRATOR', 1, 0, '2023-11-03 00:03:05');

-- --------------------------------------------------------

--
-- Table structure for table `siam_login_session`
--

CREATE TABLE `siam_login_session` (
  `id` int(11) NOT NULL,
  `username` varchar(100) DEFAULT NULL,
  `paramkey` varchar(500) DEFAULT NULL,
  `tgl_input` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- Dumping data for table `siam_login_session`
--

INSERT INTO `siam_login_session` (`id`, `username`, `paramkey`, `tgl_input`) VALUES
(1, 'admin_siam', 'MTY5OTA4ODA0OTA2NTA2NQ==EryYm1M6DqPws0l6i7m6Ie6zizZOzg', '2023-11-04 16:14:09'),
(2, 'admin_siam', 'MTY5OTA4ODA3NDA2MzQ4NQ==mc4oAa1qdtFq6OTuxnFcc6HrS1i1Rd', '2023-11-04 16:14:34'),
(3, 'admin_siam', 'MTY5OTE2NzYwOTA4MzMwNA==oUBAoAiuNfRqThVWb44GpFADLbjLOY', '2023-11-05 14:20:09'),
(4, 'admin_siam', 'MTY5OTE2NzY0MDAxNzI1NA==wv1GPnO4Js9MSKdO40Eg46RGmsRco6', '2023-11-05 14:20:40'),
(5, 'admin_siam', 'MTY5OTE2NzY1NzI4MjM1Nw==uZ0Sx53SZv8RBav3LkZqzPyehCRcJm', '2023-11-05 14:20:57'),
(6, 'admin_siam', 'MTY5OTE2NzY4NjIxMTQ4OA==pttyXCZcgNJ9Wpvxa7M3Dv7di6fKsX', '2023-11-05 14:21:26'),
(7, 'admin_siam', 'MTY5OTE2ODE1OTEwODI0MQ==t2Y0TpJqdPyozqj9j0PKWueGjs3pRq', '2023-11-05 14:29:19'),
(8, 'admin_siam', 'MTY5OTE2ODE5NjgzODkzNw==gbi7Z7mlPf0BMaDN1eVGSlrNoEof6O', '2023-11-05 14:29:56');

-- --------------------------------------------------------

--
-- Table structure for table `siam_log_activity`
--

CREATE TABLE `siam_log_activity` (
  `id` int(11) NOT NULL,
  `username` varchar(100) DEFAULT NULL,
  `page` varchar(100) DEFAULT NULL,
  `ip` varchar(100) DEFAULT NULL,
  `json_request` text DEFAULT NULL,
  `method` varchar(100) DEFAULT NULL,
  `log` varchar(100) DEFAULT NULL,
  `log_status` varchar(100) DEFAULT NULL,
  `role` varchar(100) DEFAULT NULL,
  `tgl_input` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- Dumping data for table `siam_log_activity`
--

INSERT INTO `siam_log_activity` (`id`, `username`, `page`, `ip`, `json_request`, `method`, `log`, `log_status`, `role`, `tgl_input`) VALUES
(1, 'admin_siam', 'LOGIN', '::1', '{\n    \"UserName\": \"admin_siam\",\n    \"Password\": \"7c467e39c8cf77223dc7a3a67150414b\"\n}', 'POST', 'Login Pukul 15:54 PM', 'Sukses', 'ADMINISTRATOR', '2023-11-04 15:54:09'),
(2, 'admin_siam', 'LOGIN', '::1', '{\n    \"UserName\": \"admin_siam\",\n    \"Password\": \"7c467e39c8cf77223dc7a3a67150414b\"\n}', 'POST', 'Login Pukul 15:54 PM', 'Sukses', 'ADMINISTRATOR', '2023-11-04 15:54:34'),
(3, 'admin_siam', 'LOGIN', '::1', '{\n    \"UserName\": \"admin_siam\",\n    \"Password\": \"7c467e39c8cf77223dc7a3a67150414b\"\n}', 'POST', 'Login Pukul 14:00 PM', 'Sukses', 'ADMINISTRATOR', '2023-11-05 14:00:09'),
(4, 'admin_siam', 'LOGIN', '::1', '{\n    \"UserName\": \"admin_siam\",\n    \"Password\": \"7c467e39c8cf77223dc7a3a67150414b\"\n}', 'POST', 'Login Pukul 14:00 PM', 'Sukses', 'ADMINISTRATOR', '2023-11-05 14:00:40'),
(5, 'admin_siam', 'LOGIN', '::1', '{\n    \"UserName\": \"admin_siam\",\n    \"Password\": \"7c467e39c8cf77223dc7a3a67150414b\"\n}', 'POST', 'Login Pukul 14:00 PM', 'Sukses', 'ADMINISTRATOR', '2023-11-05 14:00:57'),
(6, 'admin_siam', 'LOGIN', '::1', '{\n    \"UserName\": \"admin_siam\",\n    \"Password\": \"7c467e39c8cf77223dc7a3a67150414b\"\n}', 'POST', 'Login Pukul 14:01 PM', 'Sukses', 'ADMINISTRATOR', '2023-11-05 14:01:26'),
(7, 'admin_siam', 'LOGIN', '::1', '{\n    \"UserName\": \"admin_siam\",\n    \"Password\": \"7c467e39c8cf77223dc7a3a67150414b\"\n}', 'POST', 'Login Pukul 14:09 PM', 'Sukses', 'ADMINISTRATOR', '2023-11-05 14:09:19'),
(8, 'admin_siam', 'LOGIN', '::1', '{\n    \"UserName\": \"admin_siam\",\n    \"Password\": \"7c467e39c8cf77223dc7a3a67150414b\"\n}', 'POST', 'Login Pukul 14:09 PM', 'Sukses', 'ADMINISTRATOR', '2023-11-05 14:09:56');

-- --------------------------------------------------------

--
-- Table structure for table `siam_log_error`
--

CREATE TABLE `siam_log_error` (
  `id` int(11) NOT NULL,
  `username` varchar(100) DEFAULT NULL,
  `page` varchar(100) DEFAULT NULL,
  `error_log` varchar(2000) DEFAULT NULL,
  `json_request` text DEFAULT NULL,
  `json_response` text DEFAULT NULL,
  `error_code` int(11) DEFAULT NULL,
  `header_auth` text DEFAULT NULL,
  `method_request` varchar(100) DEFAULT NULL,
  `path` varchar(500) DEFAULT NULL,
  `IP` varchar(100) DEFAULT NULL,
  `tgl_input` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- Dumping data for table `siam_log_error`
--

INSERT INTO `siam_log_error` (`id`, `username`, `page`, `error_log`, `json_request`, `json_response`, `error_code`, `header_auth`, `method_request`, `path`, `IP`, `tgl_input`) VALUES
(1, 'admin_siam', 'LOGIN', 'Akun Anda terblokir, harap hubungi Admin SIAM', '', '', 1, '\"Header=>Postman-Token:c2d1205b-480a-4c3e-8291-3c6d10e7a332 | Accept-Encoding:gzip, deflate, br | Signature:340cbd8e1edeb0fdb612cff26f965f46 | Accept:*/* | Content-Length:55 | Content-Type:application/json | User-Agent:PostmanRuntime/7.33.0 | Cache-Control:no-cache | Connection:keep-alive\"', 'POST', '/api/v1/Login', '::1', '2023-11-04 15:52:56'),
(2, 'admin_siam', 'LOGIN', 'Akun Anda terblokir, harap hubungi Admin SIAM', '', '', 1, '\"Header=>Accept-Encoding:gzip, deflate, br | Content-Length:84 | Signature:e5d0a7056bfdfdd21ba1150e6f0db4f1 | User-Agent:PostmanRuntime/7.33.0 | Cache-Control:no-cache | Content-Type:application/json | Accept:*/* | Postman-Token:44add85e-07ed-4557-9740-24b98f3c453a | Connection:keep-alive\"', 'POST', '/api/v1/Login', '::1', '2023-11-04 15:53:49'),
(3, 'admin_siam', 'LOGIN', '', '', '', 1, '\"Header=>Signature:e5d0a7056bfdfdd21ba1150e6f0db4f1 | User-Agent:PostmanRuntime/7.33.0 | Accept:*/* | Content-Length:84 | Content-Type:application/json | Cache-Control:no-cache | Accept-Encoding:gzip, deflate, br | Connection:keep-alive | Postman-Token:39e82288-68b0-41de-abc2-bae4eb0238c0\"', 'POST', '/api/v1/Login', '::1', '2023-11-04 15:54:09');

--
-- Indexes for dumped tables
--

--
-- Indexes for table `siam_login`
--
ALTER TABLE `siam_login`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `siam_login_session`
--
ALTER TABLE `siam_login_session`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `siam_log_activity`
--
ALTER TABLE `siam_log_activity`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `siam_log_error`
--
ALTER TABLE `siam_log_error`
  ADD PRIMARY KEY (`id`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `siam_login`
--
ALTER TABLE `siam_login`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT for table `siam_login_session`
--
ALTER TABLE `siam_login_session`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=9;

--
-- AUTO_INCREMENT for table `siam_log_activity`
--
ALTER TABLE `siam_log_activity`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=9;

--
-- AUTO_INCREMENT for table `siam_log_error`
--
ALTER TABLE `siam_log_error`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
