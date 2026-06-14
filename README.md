# Netflix Central (Lokal)

Aplikasi desktop lokal untuk mengelola banyak akun Netflix berbasis profil Google Chrome yang terisolasi. Semua sesi login/cookies tersimpan secara lokal di folder profil Chrome PC masing-masing, sehingga data login tidak saling bercampur.

---

## Prasyarat
Sebelum menjalankan aplikasi, pastikan PC Anda sudah terpasang:
1. **Google Chrome** (sebagai browser utama).
2. **Go (Golang)** versi 1.21 ke atas (untuk backend API).
3. **Node.js & npm** (untuk frontend React + Vite).
4. **PostgreSQL Database Server** (berjalan secara lokal di port default `5432`).

---

## Langkah Setup Database (Pertama Kali)

1. Buka PostgreSQL Anda (bisa lewat pgAdmin/DBeaver/Terminal) lalu buat database kosong bernama **`netflixdb`**:
   ```sql
   CREATE DATABASE netflixdb;
   ```
2. *(Opsional)* Jika Anda memiliki database SQLite lama (`database/app.db`) dan ingin memindahkan datanya ke PostgreSQL lokal Anda, jalankan perintah migrasi ini di terminal:
   ```powershell
   go run scripts/migrate.go
   ```
   *Script ini secara otomatis membuat tabel-tabel di PostgreSQL dan menyalin data dari file `app.db`.*

---

## Cara Menjalankan Aplikasi

Aplikasi ini membutuhkan dua terminal yang aktif secara bersamaan (satu untuk Backend, satu untuk Frontend).

### 1) Menjalankan Backend (Terminal 1)
Buka terminal di folder utama proyek `S6-NetfliksCentral` dan jalankan:
```powershell
go run main.go
```
*   **Default PostgreSQL Credentials**: Secara default, backend akan mencoba terhubung ke PostgreSQL lokal Anda dengan username: `postgres`, password: `dina2004`, database: `netflixdb`, port: `5432`.
*   Jika kredensial PostgreSQL Anda berbeda, Anda bisa menjalankannya lewat file batch:
    ```powershell
    .\run_backend.bat
    ```
    *(Silakan edit berkas `run_backend.bat` terlebih dahulu untuk menyesuaikan password/username database Anda).*

### 2) Menjalankan Frontend (Terminal 2)
Buka terminal baru, masuk ke dalam subfolder `frontend`, lalu jalankan:
```powershell
cd frontend
npm install   # Jalankan ini hanya untuk pertama kali setup
npm run dev
```
Setelah aktif, buka alamat yang ditampilkan di browser Anda (biasanya **`http://localhost:5173`**).

---

## Cara Penggunaan Singkat
1. **Registrasi Akun Admin (Pertama Kali)**: 
   * Jika database PostgreSQL baru dibuat, Anda wajib melakukan **Register** terlebih dahulu di halaman login UI untuk membuat akun admin aplikasi.
2. **Menambah Akun Netflix**:
   * Klik tombol **Add Account**, isi label nama dan email Netflix. Sistem akan membuat profil Chrome lokal terisolasi secara otomatis.
3. **Membuka Akun Netflix**:
   * Klik kartu akun Netflix di dashboard untuk membuka jendela Google Chrome baru yang terisolasi. 
   * Login ke Netflix sekali, dan sesi login Anda akan aman tersimpan di PC Anda untuk seterusnya.

---

## Lokasi Penyimpanan Data Sesi
* **Database**: PostgreSQL (`netflixdb`).
* **Sesi Login Chrome**: Disimpan di folder lokal `chrome_profiles/<nama-profil>`. Folder ini tidak akan ikut ter-upload ke Git/GitHub demi alasan keamanan dan performa.
