# Panduan Deployment PHPForge ke Windows Server

Dokumen ini berisi daftar periksa (*checklist*) dan langkah-langkah yang diperlukan untuk memindahkan (deploy) aplikasi PHPForge beserta aplikasi Laravel Anda ke mesin Windows Server.

---

## 1. Kebutuhan File (Build Artifacts)
Anda **tidak perlu** menginstal XAMPP, Laragon, Nginx, atau Apache di Windows Server. Anda hanya perlu menyalin folder proyek ini secara utuh. 

Pastikan folder yang Anda salin memuat struktur berikut:
- `phforge-core.exe` (Aplikasi utama / Web Server & Process Manager)
- `phforge-gui.exe` (Aplikasi antarmuka pengguna / Dashboard GUI - opsional)
- `cli.exe` (Aplikasi Command Line Interface - opsional)
- Folder `dist/` beserta seluruh isinya:
  - `dist/php-versions/` (Berisi *binary* PHP yang sudah Anda unduh seperti PHP 7.4 / 8.4)
  - `dist/config/` (Berisi konfigurasi `server.json` dan `sites.json`)
  - `dist/sites/` (Tempat Anda meletakkan folder *source code* Laravel/aplikasi Anda)
  - `dist/composer/` (Berisi `composer.phar` - diperlukan jika Anda ingin menggunakan fitur Composer pada `cli.exe`)

---

## 2. Kebutuhan Sistem (Windows Server)
Meskipun PHPForge bersifat mandiri (*portable*), *binary* PHP (`php-cgi.exe`) bawaan Windows tetap membutuhkan beberapa pustaka OS.

* [ ] **Microsoft Visual C++ Redistributable (2015-2022)**
  Sangat penting! Jika tidak ada, PHP akan *error* saat dijalankan karena kehilangan file pustaka (misal: `vcruntime140.dll`). 
  *Download dari: [Situs Resmi Microsoft](https://learn.microsoft.com/en-us/cpp/windows/latest-supported-vc-redist)*
* [ ] **Database Server (MySQL / MariaDB / SQL Server)**
  PHPForge hanya bertugas sebagai Web Server dan pengelola PHP. Anda tetap harus menginstal *database server* pilihan Anda di Windows Server tersebut (atau menggunakan *database server* eksternal).

---

## 3. Konfigurasi Jaringan & Firewall
Agar aplikasi dapat diakses oleh 10 cabang retail Anda melalui intranet, Anda harus membuka jalur komunikasi.

* [ ] **Windows Defender Firewall (Inbound Rules)**
  Buka akses port untuk aplikasi web Anda. Jika situs Anda diatur untuk berjalan di port `80`, buat *New Inbound Rule* di firewall untuk mengizinkan trafik TCP pada port `80`.
* [ ] **IP Statis**
  Pastikan mesin Windows Server menggunakan *Static IP* (misal: `192.168.1.100` atau IP intranet yang telah ditentukan) agar tidak berubah-ubah saat server *restart*.

---

## 4. Cara Menjalankan Aplikasi

Ada dua metode untuk menjalankan PHPForge di server Anda: sebagai layanan latar belakang (direkomendasikan untuk Server) atau dijalankan langsung melalui konsol.

### Opsi A: Menjalankan Sebagai "Windows Service" (Direkomendasikan)
Di lingkungan *server*, aplikasi harus tetap hidup di latar belakang meskipun Anda me-*restart* server atau keluar (logout) dari *Remote Desktop* (RDP). 

**Cara Praktis (1-Klik):**
1. Klik kanan pada file **`install-service.bat`**.
2. Pilih **"Run as administrator"**.
3. Script akan otomatis mendaftarkan `PHForgeCoreSvc`, menyalakan service, dan mengaturnya ke **Automatic Startup**.

*(Opsional - Cara Manual via Command Prompt Administrator):*
```cmd
cd C:\path\to\PHPForge
phforge-core.exe --install
sc start PHForgeCoreSvc
```

Jika Anda ingin mencopot layanan di masa depan:
* Cukup klik kanan **`uninstall-service.bat`** -> **Run as administrator**, atau jalankan `phforge-core.exe --uninstall`.

### Opsi B: Menjalankan Langsung Melalui Console (Tanpa Service)
Jika Anda hanya ingin melakukan *testing*, *debugging*, atau tidak memiliki hak Administrator penuh untuk memasang *Service*, Anda bisa menjalankannya seperti aplikasi biasa.

Untuk mempermudah:
1. Buka folder aplikasi Anda.
2. Klik dua kali pada file **`start-console.bat`**.
3. Aplikasi akan menyala dalam mode *console*. Pastikan Anda **tidak menutup jendela terminal tersebut**, karena jika ditutup, server Anda akan ikut mati.

*(Sebagai alternatif, Anda juga bisa membuka terminal secara manual dan mengeksekusi `phforge-core.exe` secara langsung).*

---

## 5. Cara Membuka Dashboard Admin (GUI)
Jika Anda perlu melakukan konfigurasi seperti menambah situs baru, mengubah port, atau menambah *PHP Workers*, Anda dapat menggunakan antarmuka grafis (GUI) yang telah disediakan.

1. Buka folder `C:\path\to\PHPForge`.
2. Klik dua kali pada file **`phforge-gui.exe`**.
3. Aplikasi GUI (*Dashboard*) akan terbuka. 
4. **Penting:** Pastikan `phforge-core.exe` (Web Server) sedang menyala (baik lewat *Service* maupun *Console*) agar GUI ini bisa terhubung dan menampilkan data situs Anda.

---

> **Catatan Optimasi Traffic Tinggi:** 
> Server ini sudah dioptimasi menggunakan mekanisme *Channel Queueing Semaphore* dan *Connection Pooling* dengan konkurensi (default) 16 pekerja (*workers*). Arsitektur ini dirancang khusus untuk memproses ribuan permintaan bersamaan dari cabang ritel tanpa menimbulkan masalah TCP Port Exhaustion atau 502 Bad Gateway.
