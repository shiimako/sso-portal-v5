# SSO Portal (Standalone Edition)

Sistem Single Sign-On (SSO) dan Manajemen Pengguna Terpusat untuk lingkungan akademik. Aplikasi ini dibangun dengan pendekatan **Go Native** (Standard Library + SQLX) untuk performa tinggi dan kemudahan deployment.

## 🚀 Fitur Utama

- **Autentikasi & Otorisasi**:
  - Login dengan Google OAuth (Gmail Kampus (@pnc.ac.id)).
  - Role-Based Access Control (RBAC): Admin, Dosen, Mahasiswa.
- **Manajemen Pengguna (CRUD Lengkap)**:
  - Input data profil, kontak, dan alamat.
  - **Fitur Spesifik Role**: Input NIM untuk Mahasiswa, NIP/NUPTK untuk Dosen.
  - **Riwayat Jabatan**: Input jabatan struktural/fungsional dosen (Dekan, Kaprodi, dll) dengan lingkup unit (Jurusan/Prodi).
- **Manajemen Data Master**:
  - CRUD Jurusan (Majors).
  - CRUD Program Studi (Study Programs).
  - CRUD Jabatan (Positions).
  - CRUD Role.
- **Keamanan**:
  - Webhook Receiver dengan validasi Signature (HMAC-SHA256).
  - CSRF Protection & Secure Session Management.
- **Notifikasi**:
  - Sistem Push Notification Realtime (via Webhook).

## 🛠️ Teknologi yang Digunakan

- **Backend**: Go (Golang) 1.25.1
- **Database**: MySQL
- **Database Driver**: `jmoiron/sqlx` & `go-sql-driver/mysql`
- **Routing**: `gorilla/mux`
- **Logging**: `lumberjack` (Log rotation)
- **Frontend**: HTML5, Tailwind CSS (CDN), Alpine.js (Interaktivitas)

---

## ⚙️ Cara Instalasi & Menjalankan

### 1. Prasyarat

Pastikan sudah terinstall:

- Go (Golang)
- MySQL
- Git

### 2. Clone Repository

```bash
git clone https://github.com/shiimako/sso-portal-v5.git
cd sso-portal-v5
```

### 3. Install Dependency

Download semua library yang dibutuhkan:

```bash
go mod tidy
```

### 4. Konfigurasi Environment (.env)

Copy file contoh `.env.example` menjadi `.env` :

```bash
cp .env.example .env
```

Lalu edit file .env dan sesuaikan isinya.

### 5. Setup Database

Karena menggunakan pendekatan Native SQL, silakan import file skema database secara manual:

1. Buat database baru di MySQL (misal: sso_providers).
2. Download dan import schema database:
   - [Download Schema Database](https://raw.githubusercontent.com/shiimako/sso-portal-v5/main/database/schema.sql)
3. (Opsional) Import data awal:
   - [Download Seeder / Data Awal](https://raw.githubusercontent.com/shiimako/sso-portal-v5/main/database/seeder.sql)
   > ⚠️ Jika menggunakan seeder, Ganti email admin dengan email anda, pastikan email tersebut aktif.

### 6. Generate Keys

Jalankan seluruh generator dengan langkah berikut:

#### 1. Masuk ke direktori `cmd`

```bash
cd cmd
```

#### 2. Jalankan setiap generator key (satu per satu)

```bash
cd RSA-Key-Generate
go run main.go
cd ..

cd Session-Key-Generate
go run main.go
cd ..

cd WebPush-Key-Generate
go run main.go
cd ..
```

#### 3. Simpan hasil generate

- Salin nilai key yang ditampilkan ke dalam file `.env`
- Khusus RSA Key: file akan otomatis tergenerate sebagai:
  - `public.pem`
  - `private.pem`

> ⚠️ Catatan: Proses generate key hanya perlu dijalankan satu kali.

### 7. Jalankan Aplikasi

```bash
go run .
```
