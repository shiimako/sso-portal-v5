-- phpMyAdmin SQL Dump
-- version 5.2.2
-- https://www.phpmyadmin.net/
--
-- Host: localhost:3306
-- Generation Time: Dec 23, 2025 at 07:51 AM
-- Server version: 8.4.3
-- PHP Version: 8.3.16

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `sso_providers`
--

--
-- Dumping data for table `applications`
--

INSERT INTO `applications` (`id`, `name`, `description`, `slug`, `target_url`, `icon_url`, `category_id`) VALUES
(1, 'Attendix', 'Aplikasi untuk Absensi ', 'attendix', 'http://127.0.0.1:8000/portal/login', '/uploads/icons/app-attendix-icon.png', 1),
(2, 'SIWALI', 'Aplikasi Perwalian', 'siwali', 'http://127.0.0.1:8001/portal/login', '/uploads/icons/app-siwali-icon.png', 1);


--
-- Dumping data for table `application_notifications`
--

INSERT INTO `application_notifications` (`id`, `user_id`, `app_id`, `message`, `updated_at`) VALUES
(1, 1, 2, 'Dosen Wali memberikan solusi untuk bimbingan tanggal 17-12-2025', '2025-12-16 07:34:05');

--
-- Dumping data for table `application_position_access`
--

INSERT INTO `application_position_access` (`application_id`, `position_id`) VALUES
(2, 3);
--
-- Dumping data for table `application_role_access`
--

INSERT INTO `application_role_access` (`application_id`, `role_id`) VALUES
(1, 1),
(1, 2),
(1, 3),
(2, 3),
(2, 2),
(2, 1);

--
-- Dumping data for table `categories`
--

INSERT INTO `categories` (`id`, `name`, `sort_order`) VALUES
(1, 'Akademik', 1),
(2, 'Administrasi & Keuangan', 2),
(3, 'Kemahasiswaan', 3),
(4, 'Lainnya', 99);

--
-- Dumping data for table `lecturers`
--

INSERT INTO `lecturers` (`id`, `user_id`, `nip`, `nuptk`) VALUES
(1, 2, '198001018', '20309998');

--
-- Dumping data for table `lecturer_positions`
--

INSERT INTO `lecturer_positions` (`id`, `lecturer_id`, `position_id`, `major_id`, `study_program_id`, `start_date`, `end_date`, `updated_at`, `created_at`, `deleted_at`) VALUES
(1, 1, 1, 1, NULL, '2025-12-01', NULL, '2025-12-16 08:51:55', '2025-12-16 08:51:55', NULL);

--
-- Dumping data for table `majors`
--

INSERT INTO `majors` (`id`, `major_name`, `created_at`, `updated_at`, `deleted_at`) VALUES
(1, 'Komputer dan Bisnis', '2025-12-08 04:02:41', '2025-12-16 08:51:54', NULL),
(2, 'Rekayasa Elektro dan Mekatronika', '2025-12-08 04:02:41', '2025-12-16 08:51:54', NULL),
(3, 'Teknologi Pertanian dan Tumbuhan', '2025-12-08 04:02:42', '2025-12-22 02:56:29', NULL);

--
-- Dumping data for table `positions`
--

INSERT INTO `positions` (`id`, `position_name`, `created_at`, `updated_at`, `deleted_at`) VALUES
(1, 'Ketua Jurusan', '2025-12-08 06:36:41', '2025-12-16 08:51:56', NULL),
(2, 'Koordinator Program Studi', '2025-12-08 06:36:42', '2025-12-16 08:51:56', NULL);

--
-- Dumping data for table `roles`
--

INSERT INTO `roles` (`id`, `role_name`, `description`, `created_at`, `updated_at`, `deleted_at`) VALUES
(1, 'admin', 'Super Administrator kan?', '2025-12-08 05:39:41', '2025-12-22 02:12:54', NULL),
(2, 'dosen', 'Tenaga Pengajar', '2025-12-08 05:39:41', '2025-12-16 08:51:55', NULL),
(3, 'mahasiswa', 'Peserta Didik', '2025-12-08 05:39:41', '2025-12-16 08:51:55', NULL);

--
-- Dumping data for table `students`
--

INSERT INTO `students` (`id`, `user_id`, `nim`) VALUES
(38, 3, '230202003');

--
-- Dumping data for table `study_programs`
--

INSERT INTO `study_programs` (`id`, `study_program_name`, `created_at`, `updated_at`, `deleted_at`, `major_id`) VALUES
(1, 'D3 Teknik Informatika', '2025-12-08 05:11:49', '2025-12-16 08:51:54', NULL, 1),
(2, 'D4 Rekayasa Keamanan Siber', '2025-12-08 05:11:49', '2025-12-16 08:51:54', NULL, 1),
(3, 'D4 Akuntansi Lembaga Keuangan Syariah', '2025-12-08 05:11:50', '2025-12-16 08:51:54', NULL, 1),
(4, 'D3 Teknik Mesin', '2025-12-08 05:11:50', '2025-12-16 08:51:54', NULL, 2),
(5, 'D4 Teknologi Rekayasa Multimedia', '2025-12-08 05:11:50', '2025-12-16 08:51:54', NULL, 1);


--
-- Dumping data for table `users`
--

INSERT INTO `users` (`id`, `name`, `email`, `avatar`, `google_avatar`, `status`, `address`, `phone_number`, `created_at`, `updated_at`, `deleted_at`) VALUES
(1, 'User Dummy 1 (admin)', 'kayazmixf617@gmail.com', NULL, NULL, 'aktif', 'Jl. Admin Pusat No. 1', '081111111', '2025-09-25 03:35:27', '2025-12-16 07:29:51', NULL),
(2, 'User Dummy 2 (dosen)', 'kayazmixf617.stu@pnc.ac.id', '/uploads/avatars/user-12-avatar-1766454040644103200.jpg', NULL, 'aktif', 'Jl. Dosen No. 12', '0822222212', '2025-11-04 07:43:34', '2025-12-23 04:33:39', NULL),
(3, 'User Dummy 3 (mahasiswa)', 'dummy3@pnc.ac.id', NULL, NULL, 'aktif', 'Jl. Mahasiswa No. 3', '0833333333', '2025-11-04 07:43:34', '2025-12-16 07:30:21', NULL);


--
-- Dumping data for table `user_roles`
--

INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES
(1, 1),
(2, 2),
(3, 3);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
