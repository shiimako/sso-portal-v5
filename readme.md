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

### 2. Setup Google OAuth Credential

Aplikasi ini menggunakan Google OAuth untuk autentikasi pengguna.

Untuk mengaktifkan fitur ini, silakan lakukan langkah berikut:

1. Buka Google Cloud Console  
   https://console.cloud.google.com/

2. Pilih atau buat project baru

3. Aktifkan **Google Identity / OAuth Consent Screen**

4. Pergi ke menu `APIs & Services > Credentials` untuk membuat credentials baru

5. Buat OAuth Client ID:
   - Application type: **Web Application**
   - Authorized redirect URI:
     ```
     http://localhost:8080/auth/google/callback
     ```
     >⚠️ sesuaikan kembali redirect URI ini apabila base_url sistem mengalami perubahan.

6. Simpan nilai berikut ke file `.env`:
   ```env
   GOOGLE_CLIENT_ID=your_client_id
   GOOGLE_CLIENT_SECRET=your_client_secret
   ```

### 3. Clone Repository

```bash
git clone https://github.com/shiimako/sso-portal-v5.git
cd sso-portal-v5
```

### 4. Install Dependency

Download semua library yang dibutuhkan:

```bash
go mod tidy
```

### 5. Konfigurasi Environment (.env)

Copy file contoh `.env.example` menjadi `.env` :

```bash
cp .env.example .env
```

Lalu edit file .env dan sesuaikan isinya.

### 6. Setup Database

Karena menggunakan pendekatan Native SQL, silakan import file skema database secara manual:

1. Buat database baru di MySQL (misal: sso_providers).
2. Download dan import schema database:
   - [Download Schema Database](database/schema.sql)
3. (Opsional) Import data awal:
   - [Download Seeder / Data Awal](database/seeder.sql)
   > ⚠️ Jika menggunakan seeder, Ganti email admin dengan email anda, pastikan email tersebut aktif.

### 7. Generate Keys

Jalankan seluruh generator dengan langkah berikut:

#### 1. Masuk ke direktori `cmd`

```bash
cd cmd
```

#### 2. Jalankan setiap generator key (satu per satu)

```bash
go run cmd/RSA-Key-Generate/main.go

cd Session-Key-Generate
go run main.go
cd ..

cd WebPush-Key-Generate
go run main.go
cd ..

cd Webhook-Secret-Generate
go run main.go
cd ../..
```

#### 3. Simpan hasil generate

- Salin nilai key yang ditampilkan ke dalam file `.env`
   ```env
   SESSION_KEY="{{session key generated}}"
   
   WEBHOOK_SECRET= "{{webhook_secret generated}}"
   
   VAPID_PRIVATE_KEY="{{vapid_private_key generated}}"
   VAPID_PUBLIC_KEY="{{vapid_public_key generated}}"
   ```
- Khusus RSA Key: file akan otomatis tergenerate sebagai:
  - `public.pem`
  - `private.pem`

> ⚠️ Catatan: Proses generate key hanya perlu dijalankan satu kali.

### 8. Jalankan Aplikasi

```bash
go run .
```
