# Laporan Perbandingan Teknis & Operasional: PHPForge vs Microsoft IIS + PHP

**Proyek:** PHPForge — High-Performance Portable Multi-PHP Web Server for Windows  
**Tanggal:** 29 Oktober 2026  
**Ruang Lingkup:** Evaluasi Arsitektur, Performa, Keamanan, Kemudahan Deployment, dan Pemeliharaan Aplikasi PHP/Laravel di Lingkungan Windows Server.

---

## 1. Ringkasan Eksekutif

Menjalankan aplikasi PHP modern (seperti **Laravel**) di lingkungan **Windows Server** secara tradisional mengandalkan tumpukan **Microsoft IIS (Internet Information Services) + FastCGI + PHP Manager**. Meskipun IIS adalah server web bawaan Windows yang matang untuk ekosistem .NET, penggunaannya untuk ekosistem PHP/Laravel sering kali membutuhkan konfigurasi berlapis, modul tambahan pihak ketiga, serta rawan mengalami kendala portabilitas saat migrasi server.

**PHPForge** dirancang secara khusus sebagai alternatif modern berbasis **Go (Golang)** dan **Wails v2** yang menggabungkan HTTP Server berkinerja tinggi, orkestrator *FastCGI Worker Pool* dengan antrean *Channel Semaphore*, sistem keamanan berlapis otomatis, serta antarmuka GUI modern dalam satu paket mandiri (*100% portable*).

---

## 2. Perbandingan Arsitektur

```
+-----------------------------------------------------------------------------------+
|                        ARSITEKTUR MICROSOFT IIS + PHP                             |
+-----------------------------------------------------------------------------------+
  Request HTTP 
       │
       ▼
  [HTTP.sys (Kernel)] ──► [IIS W3SVC / AppPool] ──► [Modul URL Rewrite (web.config)]
                                                            │
                                                            ▼
                                               [IIS FastCGI Handler (iisfcgi.dll)]
                                                            │ (Spawn / Dynamic Pool)
                                                            ▼
                                                     [php-cgi.exe]
```

```
+-----------------------------------------------------------------------------------+
|                             ARSITEKTUR PHPFORGE                                   |
+-----------------------------------------------------------------------------------+
  Request HTTP 
       │
       ▼
  [Go HTTP Engine] ──► [Security Filter (.env/.git/NTFS)] ──► [Native Laravel Router]
                                                                       │
                                                                       ▼
                                                       [Channel Semaphore Queue (30s)]
                                                                       │ (Round-Robin)
                                                                       ▼
                                                    [Pre-Warmed FastCGI Pool (16-32x)]
                                                     ├─► php-cgi.exe (:9001)
                                                     ├─► php-cgi.exe (:9002)
                                                     └─► php-cgi.exe (:90xx)
```

### Perbedaan Mendasar Mekanisme Kerja:
1. **Routing & URL Rewrite:**
   * **IIS + PHP:** Wajib menginstal ekstensi terpisah **IIS URL Rewrite Module** dan membuat file `web.config` di setiap folder `public/` proyek Laravel agar rute URL bersih (*clean URL*) diarahkan ke `index.php`.
   * **PHPForge:** Memiliki logika *Laravel Front Controller* bawaan langsung di dalam kode mesin HTTP Go (`internal/httpengine/server.go`). Aplikasi Laravel dapat langsung berjalan tanpa memerlukan file `web.config` atau `.htaccess`.
2. **Manajemen Proses PHP (`php-cgi.exe`):**
   * **IIS + PHP:** Menggunakan proses dinamis yang dikelola oleh Application Pool. Ketika terjadi lonjakan trafik mendadak (*spike*), IIS harus memicu pembuatan proses `php-cgi.exe` baru (*cold-start*), yang memakan waktu inisialisasi CPU dan I/O disk.
   * **PHPForge:** Menggunakan pendekatan **Pre-Warmed Static Worker Pool** (misal: 16 hingga 32 worker aktif sejak awal) yang dipadukan dengan **Atomic Round-Robin Load Balancer** dan **Channel Queueing Semaphore**. Saat ratusan request masuk bersamaan, request tidak langsung ditolak dengan *502 Bad Gateway*, melainkan diantrekan secara efisien di memori Go hingga worker tersedia.

---

## 3. Tabel Perbandingan Head-to-Head

| Parameter Evaluasi | PHPForge (Go + FastCGI Pool) | Microsoft IIS + PHP (FastCGI) |
| :--- | :--- | :--- |
| **Instalasi & Deployment** | **Portable (Zero-Install)** — Cukup salin 1 folder & jalankan `install-service.bat` | **Kompleks** — Aktifkan fitur Windows Role, instal URL Rewrite MSI, atur FastCGI & AppPool |
| **Dukungan Multi-PHP** | **Bawaan (Built-in)** — Setiap situs bebas memilih PHP 7.4, 8.1, atau 8.4 dari GUI | Membutuhkan konfigurasi manual *Handler Mappings* per situs atau *PHP Manager for IIS* (proyek pihak ke-3 yang sudah tidak aktif) |
| **Konfigurasi Laravel Routing** | **Otomatis (Zero-Config)** — Langsung mengenali pola *Front Controller* `index.php` | Wajib membuat dan memelihara file `web.config` + modul *IIS URL Rewrite* |
| **Manajemen Lonjakan Trafik** | **Channel Semaphore Queue + Round-Robin** — Menahan lonjakan trafik tanpa *port exhaustion* | Rentan mengalami *502.3 Bad Gateway* atau *FastCGI Queue Full* jika batas `maxInstances` tidak dituning manual |
| **Keamanan File Sensitif** | **Aktif Secara Default** — Blokir otomatis `.env`, `.git`, `composer.json`, `artisan`, `.log`, `.sql`, dan NTFS `::$DATA` | Harus dikonfigurasi manual di bagian *Request Filtering* (`hiddenSegments` & `fileExtensions`) pada setiap situs |
| **Optimasi OPcache & PHP** | **1-Klik Preset GUI** — Tersedia preset *Rekomendasi Laravel* & *Enterprise* langsung dari Dashboard | Harus mencari dan mengedit baris `php.ini` secara manual menggunakan teks editor |
| **Manajemen Ekstensi PHP** | **Toggle Switch Instan di GUI** — Pencarian dan klik on/off langsung dari modal | Edit manual `php.ini` atau melalui modul ekstensi lawas di MMC |
| **Jejak Memori (Engine Footprint)** | **Sangat Ringan (~15–25 MB RAM)** untuk binary `phforge-core.exe` | **Menengah–Berat** — Melibatkan layanan `W3SVC`, `WAS`, `w3wp.exe` per Application Pool |
| **Kemudahan Pindah Server (Migrasi)** | **Sangat Mudah (Auto-Heal Path)** — Salin seluruh folder ke server baru, path otomatis menyesuaikan | **Sulit** — Harus mengekspor konfigurasi IIS (`applicationHost.config`), menyalin PHP, dan mengatur ulang *permissions* `IIS_IUSRS` |
| **Masalah Izin Folder (Permissions)** | **Minimal** — Berjalan konsisten di bawah konteks Service/User tanpa kerumitan grup akun virtual | Sering terkendala error *Permission Denied* pada folder `storage/` dan `bootstrap/cache` Laravel karena izin akun `IUSR` / `IIS_IUSRS` |

---

## 4. Analisis Performa & Konkurensi Tinggi

Pada skenario jaringan internal perusahaan / multi-cabang ritel (misalnya 10–50 cabang melakukan sinkronisasi transaksi POS atau Stock Opname secara bersamaan):

### A. Penanganan Antrean Konkurensi (Burst Traffic)
* **Pada IIS + PHP:**  
  Secara default, `php-cgi.exe` di Windows hanya mampu menangani 1 request per proses pada satu waktu (non-thread-safe dalam eksekusi skrip tunggal). Jika 200 request datang serentak sementara `maxInstances` di IIS belum siap, sebagian koneksi akan mengalami antrean di level named-pipe/TCP IIS atau terputus dengan kode `502 Bad Gateway`.
* **Pada PHPForge:**  
  Setiap worker `php-cgi.exe` memiliki port TCP lokal terdedikasi (`127.0.0.1:9001` s/d `9032`) yang sudah menyala sejak service dijalankan (*zero cold-start*). Di depannya, Go menggunakan `chan struct{}` (*buffered channel semaphore*) yang menahan dan mendistribusikan ribuan *goroutine* ringan (masing-masing hanya berukuran ~2 KB di RAM) secara *Round-Robin* ke worker yang selesai paling cepat.

### B. Dampak Aktivasi Zend OPcache Terpadu
Pada sistem file Windows (NTFS), operasi pembukaan ribuan file kecil (seperti folder `vendor/` pada Laravel) jauh lebih lambat dibandingkan di Linux (ext4/xfs) karena *overhead* penguncian metadata file Windows Defender dan NTFS.
* **PHPForge** mengatasi *bottleneck* utama Windows ini dengan menghadirkan fitur **1-Click Laravel Tuning Preset** yang memastikan `zend_extension=opcache`, `opcache.enable=1`, `opcache.memory_consumption=128M`, dan `opcache.max_accelerated_files=10000` terkonfigurasi sempurna.
* Berdasarkan hasil uji beban **k6** pada proyek ini, waktu eksekusi endpoint ringan dengan OPcache aktif di PHPForge mencapai **< 1 ms (hingga 0.009 ms di sisi PHP)** dengan tingkat kegagalan **0.00%**.

---

## 5. Analisis Keamanan (Security Posture)

Salah satu risiko terbesar saat men-deploy proyek Laravel di Windows adalah kesalahan penempatan *Document Root* (mengarahkan ke folder root proyek, bukan ke folder `public/`) atau lupa memblokir file tersembunyi, sehingga file `.env` yang berisi kredensial database dapat diunduh langsung melalui browser (`http://ip-server/.env`).

| Vektor Serangan | Perilaku di PHPForge | Perilaku di IIS + PHP |
| :--- | :--- | :--- |
| Akses `/.env` atau `/.git/config` | **Diblokir Otomatis (403 Forbidden)** + Tercatat di `audit.log` dengan label `[SECURITY]` | Terbuka ke publik jika *Request Filtering* belum dikonfigurasi atau `DocumentRoot` salah diarahkan |
| Akses `/composer.json` & `/artisan` | **Diblokir Otomatis (403 Forbidden)** | Dapat terunduh sebagai file teks jika ekstensi/nama file tidak masuk daftar blokir `web.config` |
| Akses File Log & Backup (`*.log`, `*.sql`, `*.bak`, `*.sqlite`) | **Diblokir Otomatis (403 Forbidden)** | Dapat diakses publik jika berada di dalam jangkauan direktori web |
| Exploit Windows NTFS Stream (`index.php::$DATA`) | **Diblokir Otomatis (403 Forbidden)** di level parser URL | Bergantung pada aturan *Request Filtering* IIS |
| Serangan *Path Traversal* (`../../`) | **Diblokir Otomatis (403 Forbidden)** dengan validasi prefix `filepath.Clean` | Diblokir oleh IIS |

---

## 6. Kemudahan Operasional & Manajemen (DevOps / SysAdmin)

### Skenario: Memindahkan Server ke Mesin Windows Server Baru
* **Menggunakan IIS + PHP (Estimasi Waktu: 1 – 3 Jam):**
  1. Buka *Server Manager* -> *Add Roles and Features* -> Pilih *Web Server (IIS)* beserta fitur *CGI*.
  2. Unduh dan instal *VC++ Redistributable*.
  3. Unduh dan instal *IIS URL Rewrite Module 2.1*.
  4. Unduh binary PHP, ekstrak ke `C:\PHP`, salin dan edit `php.ini` secara manual.
  5. Buka IIS Manager -> Daftarkan *FastCGI Application* dan *Handler Mappings* (`*.php`).
  6. Buat situs baru, arahkan *Physical Path*, lalu atur izin keamanan folder (*Edit Permissions*) untuk grup `IIS_IUSRS` dan `IUSR` pada folder `storage/` Laravel.
  7. Pastikan file `web.config` tersedia dan tidak rusak.

* **Menggunakan PHPForge (Estimasi Waktu: < 3 Menit):**
  1. Pastikan *VC++ Redistributable* terpasang.
  2. Salin folder `PHPForge-Deploy` ke lokasi mana pun di server (misal `C:\PHPForge` atau `D:\PHPForge`).
  3. Klik kanan `install-service.bat` -> **Run as administrator**.
  4. Server beserta seluruh konfigurasi PHP, OPcache, ekstensi, dan daftar situs langsung berjalan otomatis (*Auto-Heal DocumentRoot Path* akan menyesuaikan lokasi folder secara otomatis).

---

## 7. Kapan Memilih PHPForge vs IIS + PHP?

Agar evaluasi bersifat objektif, berikut adalah panduan pemilihan berdasarkan kebutuhan infrastruktur:

### Pilih **PHPForge** apabila:
* Fokus utama server adalah menjalankan aplikasi **PHP / Laravel** (baik satu aplikasi maupun banyak proyek dengan versi PHP berbeda).
* Anda menginginkan **kemudahan distribusi ke server cabang atau on-premise** tanpa perlu mengonfigurasi ulang IIS di setiap mesin.
* Anda membutuhkan **Dashboard GUI ringkas dan modern** yang memudahkan pengaturan *PHP Workers*, *Zend OPcache*, *Memory/Upload Limit*, dan *Extensions* tanpa menyentuh file teks konfigurasi yang rumit.
* Anda mengutamakan **keamanan bawaan (*secure-by-default*)** terhadap kebocoran file `.env`, `.git`, dan `.log` pada aplikasi Laravel.

### Pilih **Microsoft IIS + PHP** apabila:
* Server yang sama wajib menjalankan aplikasi **.NET Framework / ASP.NET Core** berdampingan dengan PHP di port yang sama menggunakan *IIS Application Request Routing (ARR)*.
* Perusahaan mewajibkan penggunaan **Windows Authentication (Active Directory / Kerberos NTLM)** yang terintegrasi langsung di level modul IIS.
* Tim infrastruktur sudah memiliki automasi *PowerShell DSC / Ansible* khusus yang sepenuhnya terikat pada manajemen `applicationHost.config` milik IIS.

---

## 8. Kesimpulan

**PHPForge** memberikan keunggulan signifikan dibandingkan **IIS + PHP** dalam hal **kesederhanaan operasional, portabilitas lintas server, keamanan bawaan untuk struktur proyek Laravel, dan kemudahan tuning performa**. 

Dengan menghilangkan ketergantungan pada modul eksternal seperti *IIS URL Rewrite*, *file `web.config`*, pengaturan izin *`IIS_IUSRS`*, serta menyediakan manajemen *Pre-Warmed FastCGI Worker Pool* dan *1-Click Laravel OPcache Tuning*, PHPForge mampu memangkas waktu deployment dari hitungan jam menjadi hitungan menit sekaligus menjaga stabilitas tinggi pada beban konkurensi besar.

