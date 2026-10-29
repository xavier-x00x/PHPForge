# Product Requirements Document (PRD)
## Multi-Version PHP Web Server Manager untuk Windows Server (Berbasis Delphi 10.3)

| Field | Detail |
|---|---|
| Nama Produk (sementara) | **PHPForge Server** *(codename, dapat diganti sesuai branding internal)* |
| Versi Dokumen | 1.0 |
| Tanggal | 23 Juli 2026 |
| Status | Draft untuk Review |
| Platform Target | Windows Server (2016/2019/2022) & Windows 10/11 (dev environment) |
| Bahasa Pemrograman | Delphi 10.3 (Rio) |
| Pemilik Dokumen | Tim Product/Engineering |

---

## 1. Ringkasan Eksekutif

PHPForge Server adalah aplikasi **Windows Service** yang dibangun menggunakan **Delphi 10.3** untuk menjalankan web server yang mendukung **banyak versi PHP secara bersamaan** (PHP 5.6 hingga PHP 8.x) dalam satu instance server produksi. Produk ini juga menyediakan **CLI tool** untuk menjalankan **Composer** dengan versi PHP yang dapat dipilih per proyek, serta **manajemen extension PHP secara dinamis** (aktif/nonaktif tanpa restart total service).

Kebutuhan utama: produk ini harus dapat **berjalan berdampingan dengan aman** bersama XAMPP dan/atau Laragon yang mungkin sudah terpasang di mesin yang sama, tanpa bentrok port, file, registry, atau environment variable.

---

## 2. Latar Belakang & Masalah

Saat ini tim development/operasional menghadapi kendala:

1. Aplikasi legacy menggunakan PHP versi lama (5.6/7.x), sementara aplikasi baru menggunakan PHP 8.x — sulit dijalankan dalam satu server tanpa konflik.
2. Tools seperti XAMPP/Laragon lebih ditujukan untuk **development**, bukan **production** (kurang kontrol proses, kurang logging terstruktur, tidak dirancang sebagai Windows Service yang tangguh untuk produksi).
3. Tidak ada cara mudah menjalankan **Composer** dengan versi PHP yang berbeda-beda per proyek dari satu CLI tool terpusat.
4. Konfigurasi extension PHP (misalnya `intl`, `gd`, `mbstring`, `pdo_sqlsrv`, dll.) sering harus diubah manual di `php.ini` dan butuh restart service — tidak efisien untuk lingkungan produksi multi-tenant.
5. Instalasi produk baru berisiko merusak instalasi XAMPP/Laragon yang sudah berjalan di server yang sama (bentrok port 80/443, PATH, registry Apache/Windows service).

---

## 3. Tujuan Produk (Goals)

- **G1** — Menyediakan web server produksi (Windows Service) yang stabil, ringan, dan native Windows menggunakan Delphi.
- **G2** — Mendukung eksekusi **multi-version PHP** secara paralel (berbeda situs/vhost bisa pakai versi PHP berbeda).
- **G3** — Menyediakan **CLI tool** (`phforge`) untuk menjalankan Composer dengan versi PHP yang dipilih per proyek/perintah.
- **G4** — Memungkinkan **manajemen extension PHP secara dinamis** (melalui GUI Admin Panel) per versi PHP dan per site, tanpa mengganggu site lain.
- **G5** — Memastikan **isolasi total** (port, path, service, registry) agar aman berdampingan dengan XAMPP/Laragon.
- **G6** — Menyediakan tingkat keamanan production-grade (proses terisolasi, permission minimal, logging audit).

### Non-Goals (Di Luar Tujuan)

- Tidak bertujuan menggantikan XAMPP/Laragon untuk kebutuhan development lokal casual.
- Tidak membangun ulang mesin PHP (tetap menggunakan PHP binary resmi dari php.net, bukan re-implementasi interpreter).
- Tidak menyediakan database engine sendiri (MySQL/PostgreSQL dianggap dependency eksternal yang dikelola terpisah).
- Versi awal tidak mencakup load balancing multi-server/cluster (dicatat sebagai roadmap masa depan).

---

## 4. Target Pengguna

| Persona | Kebutuhan |
|---|---|
| **DevOps / SysAdmin Windows Server** | Deploy banyak aplikasi PHP versi berbeda di satu server produksi, kontrol penuh, monitoring, keamanan. |
| **Backend Developer PHP** | Menjalankan `composer install/update` dengan versi PHP yang sesuai proyek tanpa ubah PATH manual. |
| **Tim Hosting Internal (perusahaan)** | Multi-tenant hosting internal untuk beberapa aplikasi/departemen dengan kebutuhan PHP version berbeda. |
| **QA/Tester** | Menjalankan environment mirip produksi tanpa merusak setup XAMPP/Laragon yang dipakai untuk keperluan lain. |

---

## 5. Ruang Lingkup

### 5.1 Dalam Lingkup (In-Scope)

- Core web server engine (HTTP internal) berbasis Delphi, berjalan sebagai Windows Service.
- **Nginx** sebagai reverse-proxy front-facing (TLS termination, HTTP/2 ke client) di depan PHForge Core Service.
- Reverse-proxy/dispatcher ke FastCGI PHP (`php-cgi.exe`) multi-versi.
- Manajemen konfigurasi site/vhost (domain, port, document root, versi PHP, extension aktif).
- CLI tool `phforge` — **khusus** untuk menjalankan Composer dengan versi PHP tertentu (kebutuhan dependency management proyek PHP/Laravel).
- GUI Admin Panel (VCL) sebagai **satu-satunya antarmuka** untuk seluruh manajemen server: site, versi PHP, extension, service control (start/stop/restart/reload), dan monitoring.
- Mekanisme deteksi & pencegahan konflik dengan XAMPP/Laragon.
- Logging (access log, error log, audit log) dan monitoring dasar (CPU/RAM per worker, health check).
- Auto-restart worker PHP yang crash (self-healing).
- Dokumentasi instalasi & panduan migrasi dari XAMPP/Laragon.

### 5.2 Di Luar Lingkup (Out-of-Scope, versi 1.0)

- Container/Docker support.
- Load balancing multi-node/cluster.
- Built-in database server.
- Panel hosting berbayar/multi-user billing (cPanel-like commercial features).
- Dukungan Linux (fokus 100% Windows Server).

---

## 6. Persyaratan Fungsional

Prioritas: **M** = Must Have, **S** = Should Have, **C** = Could Have.

### 6.1 Web Server Core / HTTP Engine

| ID | Requirement | Prioritas | Acceptance Criteria |
|---|---|---|---|
| FR-01 | Service dapat menerima request HTTP pada port internal yang dapat dikonfigurasi (default bukan 80/443 agar tidak bentrok), dengan **Nginx** berperan sebagai reverse-proxy front-facing yang menangani koneksi publik (termasuk TLS termination) sebelum diteruskan ke PHForge Core Service. | M | Nginx menerima request client (HTTP/HTTPS, port dapat dikonfigurasi, default bukan 80/443), melakukan TLS termination (TLS 1.2/1.3), lalu meneruskan request ke PHForge Core Service via HTTP internal pada port terpisah. |
| FR-02 | Mendukung banyak virtual host (domain/subdomain) dalam satu instance. | M | Minimal 50 vhost aktif tanpa penurunan performa signifikan. |
| FR-03 | Static file serving (HTML, CSS, JS, gambar) langsung tanpa melalui PHP. | M | File statis disajikan dengan header cache yang benar (ETag/Last-Modified). |
| FR-04 | Dispatch request `.php` ke worker PHP sesuai versi yang dikonfigurasi per vhost via FastCGI. | M | Request ke site A (PHP 7.4) dan site B (PHP 8.3) berjalan bersamaan tanpa konflik. |
| FR-05 | Konfigurasi berbasis file (JSON/YAML/INI) yang dapat di-reload tanpa restart penuh service (hot-reload). | S | Perubahan config site baru langsung aktif dalam < 5 detik tanpa downtime site lain. |
| FR-06 | Mendukung URL rewriting sederhana (mirip `.htaccess` untuk framework seperti Laravel/CodeIgniter). | M | Rule rewrite dasar (`front controller pattern`) berfungsi. |
| FR-07 | Mendukung **HTTP/2** (selain HTTP/1.1) untuk koneksi HTTPS, ditangani oleh Nginx pada sisi client-facing; koneksi internal Nginx → PHForge Core Service tetap menggunakan HTTP/1.1. | M | Client yang mendukung HTTP/2 (browser modern) berhasil melakukan negosiasi ALPN ke HTTP/2 saat mengakses site via Nginx; fallback otomatis ke HTTP/1.1 untuk client yang tidak mendukung. |

### 6.2 Manajemen Multi-Version PHP

| ID | Requirement | Prioritas | Acceptance Criteria |
|---|---|---|---|
| FR-10 | Mendukung instalasi dan pendaftaran banyak versi PHP (contoh: 7.4, 8.0, 8.1, 8.2, 8.3) secara independen (folder terpisah, tidak share `php.ini` global). | M | Minimal 5 versi PHP terdaftar & dapat dipilih per site. |
| FR-11 | Setiap versi PHP dijalankan sebagai pool worker FastCGI terpisah (proses terisolasi). | M | Crash pada worker PHP 7.4 tidak mempengaruhi worker PHP 8.3. |
| FR-12 | Admin dapat menambah/menghapus versi PHP via **GUI Admin Panel** tanpa mengedit file manual. | M | Melalui form GUI, admin berhasil mendaftarkan versi PHP baru (versi + path folder instalasi) tanpa command line maupun edit file manual. |
| FR-13 | Auto-scaling jumlah worker per versi berdasarkan beban (min/max worker configurable). | S | Worker bertambah otomatis saat load tinggi, berkurang saat idle. |
| FR-14 | Versi PHP default dapat ditentukan per vhost, dan fallback jika versi tidak ditemukan menghasilkan error yang jelas (bukan crash). | M | Site dengan versi PHP yang belum terdaftar menampilkan error 500 dengan log jelas, bukan crash service. |

### 6.3 CLI Tool untuk Composer Multi-Version

> **Catatan Cakupan**: CLI `phforge` **hanya** digunakan untuk menjalankan Composer (dependency management proyek PHP/Laravel). Seluruh fungsi administrasi lain — site, versi PHP, extension, service control, dan monitoring — dilakukan melalui **GUI Admin Panel** (lihat §6.7), bukan CLI.

| ID | Requirement | Prioritas | Acceptance Criteria |
|---|---|---|---|
| FR-20 | CLI tool (`phforge`) dapat menjalankan Composer dengan versi PHP tertentu: `phforge composer install --php=8.2 --path=D:\sites\myapp`. | M | Composer dijalankan menggunakan `php.exe` versi 8.2 yang terisolasi, tidak menyentuh PHP default sistem. |
| FR-21 | Mendukung semua perintah Composer standar (`install`, `update`, `require`, `remove`, `dump-autoload`, dll.) diteruskan apa adanya ke Composer. | M | Output Composer identik dengan menjalankan langsung, hanya PHP binary yang dikontrol. |
| FR-22 | Auto-deteksi versi PHP dari file konfigurasi proyek (misal `.phforge.json` atau `composer.json` custom field) jika `--php` tidak disebutkan. | S | Jika ada file config proyek, versi PHP terpilih otomatis sesuai file tsb. |
| FR-23 | Composer binary (`composer.phar`) dikelola terpusat, dapat di-update via `phforge composer self-update`. | S | Versi Composer dapat di-cek dan diperbarui dari CLI. |
| FR-24 | CLI tidak mengubah environment variable **PATH** global sistem (agar tidak bentrok dengan PHP XAMPP/Laragon yang mungkin ada di PATH). | M | Setelah instalasi, `php -v` global (jika dijalankan user di CMD biasa) tetap menunjukkan PHP XAMPP/Laragon, bukan berubah ke PHP produk ini. |
| FR-25 | CLI `phforge` tidak menyediakan subperintah administrasi selain `composer` (manajemen site/versi PHP/extension/service dilakukan eksklusif via GUI). | M | CLI `phforge` hanya mengekspos subperintah `composer ...`; tidak ada subperintah `site`, `php`, `ext`, `status`, atau `reload`. |

### 6.4 Manajemen Extension PHP Dinamis

| ID | Requirement | Prioritas | Acceptance Criteria |
|---|---|---|---|
| FR-30 | Admin dapat mengaktifkan/menonaktifkan extension PHP (`.dll`) per versi PHP dan/atau per site melalui **GUI Admin Panel**. | M | Melalui toggle/checkbox di GUI, admin mengaktifkan extension (misal `intl`) tanpa edit manual `php.ini` maupun command line. |
| FR-31 | Perubahan extension untuk suatu versi PHP hanya me-restart worker pool versi tersebut (tidak mempengaruhi versi PHP lain). | M | Restart worker PHP 8.2 tidak menyebabkan downtime pada site yang pakai PHP 7.4. |
| FR-32 | Validasi otomatis sebelum apply: cek apakah file `.dll` extension ada & kompatibel dengan versi PHP (Thread Safe/Non-Thread Safe, x64/x86). | M | Jika extension tidak kompatibel, sistem menolak perubahan dengan pesan error jelas, bukan menyebabkan PHP gagal start. |
| FR-33 | Daftar extension per site dapat berbeda-beda meski menggunakan versi PHP yang sama (melalui `php.ini` override per site/pool). | S | Site A dan Site B sama-sama pakai PHP 8.2 tapi extension aktifnya berbeda. |
| FR-34 | Rollback otomatis jika perubahan extension menyebabkan worker gagal start berkali-kali. | S | Jika worker gagal start 3x berturut-turut setelah perubahan extension, sistem otomatis revert ke konfigurasi sebelumnya dan mengirim notifikasi/log. |

### 6.5 Kompatibilitas & Isolasi dengan XAMPP/Laragon

| ID | Requirement | Prioritas | Acceptance Criteria |
|---|---|---|---|
| FR-40 | Instalasi mendeteksi apakah XAMPP/Laragon terpasang & sedang berjalan (cek service Apache/Nginx, proses `httpd.exe`, `laragon.exe`, port 80/443/3306). | M | Installer menampilkan peringatan jika terdeteksi konflik port sebelum instalasi selesai. |
| FR-41 | Produk berjalan sebagai **Windows Service dengan nama unik sendiri** (bukan menggantikan service Apache/Nginx yang ada). | M | `services.msc` menampilkan service baru terpisah dari "Apache2.4"/service Laragon. |
| FR-42 | Semua file produk (binary, config, log) disimpan di direktori terisolasi sendiri (misal `C:\PHForge\`), tidak menimpa folder XAMPP (`C:\xampp`) atau Laragon (`C:\laragon`). | M | Tidak ada file yang di-overwrite di direktori XAMPP/Laragon selama instalasi/uninstall. |
| FR-43 | Port HTTP/HTTPS default produk **berbeda** dari default XAMPP/Laragon (misal default 8080/8443, dikonfigurasi ulang oleh admin), dengan pengecekan port bentrok saat startup. | M | Service gagal start dengan pesan error jelas jika port yang dikonfigurasi sudah dipakai proses lain, bukan crash diam-diam. |
| FR-44 | Tidak memodifikasi file `hosts` Windows kecuali disetujui eksplisit oleh admin (opt-in). | M | Tidak ada perubahan `C:\Windows\System32\drivers\etc\hosts` tanpa konfirmasi user. |
| FR-45 | Tidak mengubah environment variable global (`PATH`, `PHPRC`) yang dipakai XAMPP/Laragon; jika perlu, gunakan environment variable **scoped** (khusus proses child yang dijalankan produk ini). | M | `echo %PATH%` di sesi CMD baru tetap sama sebelum & sesudah instalasi. |
| FR-46 | Registry key (jika ada) ditulis di path terpisah (`HKLM\SOFTWARE\PHForge\...`), tidak menimpa key milik XAMPP/Laragon. | M | Uninstall produk tidak menghapus/mengubah registry key milik aplikasi lain. |
| FR-47 | Uninstaller membersihkan seluruh komponen produk (service, file, registry, firewall rule) tanpa menyisakan residu yang mengganggu XAMPP/Laragon. | S | Setelah uninstall, XAMPP/Laragon tetap berjalan normal tanpa perlu reinstall. |

### 6.6 Logging & Monitoring

| ID | Requirement | Prioritas | Acceptance Criteria |
|---|---|---|---|
| FR-50 | Access log & error log per site, format standar (mirip Apache Combined Log Format) agar mudah diintegrasikan tools lain (misal parsing log ke ELK). | M | Log dapat di-parse otomatis oleh tools analisis log umum. |
| FR-51 | Audit log untuk setiap perubahan konfigurasi (siapa, kapan, apa yang diubah) — penting untuk lingkungan produksi. | M | Setiap perubahan via CLI/GUI tercatat dengan timestamp & identitas user OS. |
| FR-52 | Health check endpoint internal (misal `/phforge-health`) untuk monitoring uptime. | S | Endpoint mengembalikan status JSON (uptime, jumlah worker aktif per versi PHP, memory usage). |
| FR-53 | Log rotation otomatis (harian/berdasarkan ukuran) agar tidak memenuhi disk. | M | Log lama otomatis di-arsipkan/dihapus sesuai retention policy yang dikonfigurasi. |

### 6.7 GUI Admin Panel (Antarmuka Utama Administrasi)

GUI Admin Panel adalah **satu-satunya antarmuka** untuk seluruh fungsi administrasi server (site, versi PHP, extension, service control, monitoring). CLI `phforge` **tidak** digunakan untuk fungsi-fungsi ini — lihat catatan cakupan di §6.3.

| ID | Requirement | Prioritas | Acceptance Criteria |
|---|---|---|---|
| FR-60 | Aplikasi desktop (VCL) untuk mengelola site (add/edit/remove/list) secara visual. | M | Admin dapat menambah, mengedit, menghapus, dan melihat daftar site tanpa command line maupun edit file manual. |
| FR-61 | GUI untuk menambah/menghapus/mendaftarkan versi PHP. | M | Admin dapat mendaftarkan versi PHP baru (pilih folder instalasi) melalui form GUI. |
| FR-62 | GUI untuk mengaktifkan/menonaktifkan extension PHP per versi dan/atau per site. | M | Perubahan extension diterapkan melalui toggle/checkbox di GUI, dengan validasi kompatibilitas (lihat FR-32). |
| FR-63 | GUI untuk kontrol service (start/stop/restart/reload/status). | M | Admin dapat start/stop/restart/reload service dari GUI tanpa perlu `services.msc` atau command line. |
| FR-64 | Dashboard monitoring real-time (grafik CPU/RAM per worker, jumlah request, status tiap versi PHP). | M | Update data setiap ≤5 detik. |

---

## 7. Persyaratan Non-Fungsional

### 7.1 Performa

- Mampu menangani minimal **500 concurrent connections** per instance server dengan spesifikasi hardware standar (4 vCPU/8GB RAM) tanpa penurunan response time signifikan.
- Overhead dispatching request ke worker PHP (di luar waktu eksekusi PHP itu sendiri) harus < 5ms.
- Startup service penuh (semua versi PHP + semua site) < 15 detik pada kondisi normal.

### 7.2 Keamanan

- Proses worker PHP berjalan dengan **least privilege** (bukan akun Administrator/SYSTEM), menggunakan akun service khusus dengan permission terbatas ke folder yang relevan saja.
- Isolasi antar site (site A tidak bisa mengakses file site B) melalui restriksi `open_basedir` per pool PHP.
- Validasi input pada semua endpoint CLI/GUI/API konfigurasi untuk mencegah command injection/path traversal saat mendaftarkan versi PHP/extension.
- Semua komunikasi admin (CLI ke service, GUI ke service) menggunakan named pipe/local socket dengan otentikasi lokal, bukan port network terbuka tanpa proteksi.
- Dukungan TLS untuk HTTPS dengan kemampuan load sertifikat per vhost (SNI).
- Tidak menyimpan credential (jika ada) dalam plaintext; gunakan Windows DPAPI untuk enkripsi data sensitif konfigurasi.
- Rutin security audit terhadap dependency (Composer packages, PHP versi yang sudah EOL) — beri peringatan jika versi PHP yang digunakan sudah tidak mendapat security update resmi.

### 7.3 Reliabilitas & Stabilitas (Production-Grade)

- Auto-restart worker yang crash, dengan circuit breaker (jika crash berulang > N kali dalam waktu singkat, worker di-disable sementara & admin diberi notifikasi, bukan restart-loop tanpa henti).
- Graceful reload konfigurasi (tanpa memutus koneksi yang sedang berjalan).
- Service harus dapat di-set auto-start saat Windows boot, dengan recovery action (restart otomatis) dikonfigurasi di Windows Service Control Manager.
- Backup otomatis konfigurasi sebelum perubahan besar (versi PHP/extension) untuk memudahkan rollback manual.

### 7.4 Kompatibilitas Sistem

- Windows Server 2016, 2019, 2022 (x64) — wajib.
- Windows 10/11 (x64) untuk keperluan development/testing — didukung.
- .NET/VC++ Redistributable dependency didokumentasikan jelas (karena PHP Windows build butuh VC++ Redistributable tertentu per versi).
- Kompatibel berjalan berdampingan dengan XAMPP, Laragon, IIS (asalkan port tidak bentrok).

### 7.5 Maintainability

- Kode Delphi modular: pemisahan jelas antara Core HTTP Engine, PHP Process Manager, Config Manager, CLI, dan GUI.
- Dokumentasi API internal (jika modul-modul saling berkomunikasi lewat RPC/IPC).
- Automated testing minimal untuk modul kritikal (dispatch request, proses spawning PHP, hot-reload config).

---

## 8. Arsitektur Teknis (High-Level)

### 8.1 Komponen Utama

0. **Nginx Reverse Proxy (Front-Facing)**
   Menangani seluruh koneksi publik dari client: TLS termination, negosiasi HTTP/2, dan meneruskan (proxy_pass) request ke PHForge Core Service melalui HTTP internal. Dikelola sebagai proses/service terpisah (bukan bagian dari kode Delphi), dibundel bersama instalasi PHForge dengan port dan konfigurasi yang terisolasi dari instalasi Nginx/Apache milik XAMPP atau Laragon.

1. **PHForge Core Service** (Windows Service, Delphi)
   Menangani listener HTTP internal (dari Nginx), routing berdasarkan vhost, dan dispatch ke pool FastCGI PHP yang sesuai.

2. **PHP Process Manager**
   Modul yang mengelola lifecycle proses `php-cgi.exe` per versi PHP (spawn, monitor, restart, scaling worker), menggunakan FastCGI protocol untuk komunikasi (bukan CGI klasik, demi performa).

3. **Config Manager**
   Membaca/menulis file konfigurasi (format JSON disarankan untuk kemudahan parsing & versioning), mendukung hot-reload dengan file-watcher.

4. **CLI Tool (`phforge.exe`)** — khusus untuk menjalankan Composer dengan versi PHP tertentu.
   Berkomunikasi dengan Core Service via **Named Pipe** lokal hanya untuk mendapatkan info path binary PHP yang sesuai; tidak menyediakan perintah administrasi lain (site/versi PHP/extension/service dikelola via GUI).

5. **Extension Manager**
   Submodul Config Manager yang memvalidasi & menerapkan perubahan `php.ini` per versi/per site secara terisolasi (menggunakan direktori `php.ini` override, bukan edit file global).

6. **Isolation/Conflict Guard**
   Modul yang dijalankan saat instalasi & startup untuk mendeteksi konflik dengan XAMPP/Laragon (cek port, cek proses berjalan, cek service terdaftar) — termasuk memeriksa apakah Laragon sedang dikonfigurasi menggunakan Nginx sebagai web server engine-nya, agar instance Nginx milik PHForge tidak bentrok port/proses dengan Nginx milik Laragon.

7. **GUI Admin Panel** — antarmuka utama administrasi server (site, versi PHP, extension, service control, monitoring).
   Aplikasi VCL terpisah yang berkomunikasi dengan Core Service melalui Named Pipe/local API.

### 8.2 Alur Request (Contoh)

```
Browser/Client
      │  (HTTPS, HTTP/2, port publik dikonfigurasi)
      ▼
Nginx Reverse Proxy (Front-Facing)
      │  (TLS termination, ALPN/HTTP2 negotiation)
      │  proxy_pass → HTTP internal
      ▼
PHForge Core Service (listener internal, mis. :8080)
      │  (baca Host header → cocokkan ke konfigurasi vhost)
      ▼
Router / Vhost Matcher
      │
      ├── Jika file statis → serve langsung dari disk
      │
      └── Jika *.php → forward via FastCGI ke:
                 PHP Process Manager
                        │
                        ▼
             Pool Worker PHP versi X.Y
             (dengan php.ini + extension sesuai site)
                        │
                        ▼
             Response dikembalikan ke Core Service → Client
```

### 8.3 Teknologi Delphi yang Digunakan

- **Nginx untuk Windows** (binary resmi) sebagai reverse-proxy front-facing, dibundel dan dikelola sebagai Windows Service terpisah oleh installer PHForge (konfigurasi `nginx.conf` di-generate/disinkronkan otomatis dari Config Manager berdasarkan `sites.json`).
- **Indy (Internet Direct)** atau **Synapse** untuk komponen socket/HTTP server tingkat rendah.
- Implementasi **FastCGI client** custom (karena Delphi tidak punya library FastCGI resmi) untuk komunikasi ke `php-cgi.exe`.
- **Windows Service Application** bawaan Delphi (`TService`) sebagai basis Core Service.
- **Named Pipes API** (via Windows API binding) untuk komunikasi CLI/GUI ↔ Core Service.
- Format konfigurasi: JSON (menggunakan library seperti `System.JSON` bawaan Delphi 10.3).
- Logging: modul custom dengan buffered writer + rotation, atau integrasi dengan library logging pihak ketiga yang kompatibel Delphi 10.3.

---

## 9. Instalasi & Deployment

1. **Installer** (disarankan berbasis Inno Setup atau InstallAware, bukan MSI kompleks) yang:
   - Menjalankan **Conflict Guard** sebelum instalasi (cek port, cek service XAMPP/Laragon).
   - Meminta konfirmasi port yang akan digunakan (default bukan 80/443).
   - Mendaftarkan Windows Service dengan nama unik (misal `PHForgeCoreSvc`).
   - Tidak mengubah PATH sistem secara default (opsional, dengan peringatan jelas jika user memilih untuk menambahkannya).
2. Struktur folder disarankan:
   ```
   C:\PHForge\
   ├── bin\                  (Core Service, CLI, GUI binaries)
   ├── php-versions\
   │   ├── php-7.4\
   │   ├── php-8.1\
   │   └── php-8.3\
   ├── composer\
   │   └── composer.phar
   ├── sites\
   │   ├── site-a\
   │   └── site-b\
   ├── config\
   │   ├── server.json
   │   ├── sites.json
   │   └── php-versions.json
   └── logs\
   ```
3. Dokumentasi migrasi: panduan bagi user yang ingin memindahkan site dari XAMPP/Laragon ke PHForge Server (copy dokumen root, sesuaikan konfigurasi vhost & versi PHP).

---

## 10. Use Case Utama (Ringkas)

| # | Use Case | Aktor |
|---|---|---|
| UC-1 | Admin menambahkan site baru dengan PHP 8.2 dan extension `pdo_sqlsrv` aktif (melalui GUI Admin Panel). | SysAdmin |
| UC-2 | Developer menjalankan `phforge composer install --php=7.4` untuk proyek legacy. | Developer |
| UC-3 | Admin mengubah versi PHP suatu site dari 7.4 ke 8.1 tanpa downtime site lain (melalui GUI Admin Panel). | SysAdmin |
| UC-4 | Admin menonaktifkan extension `xdebug` di lingkungan produksi untuk PHP 8.3 demi performa (melalui GUI Admin Panel). | SysAdmin |
| UC-5 | Sistem mendeteksi port 80 sudah dipakai Laragon saat instalasi, dan menyarankan port alternatif. | Sistem (otomatis) |
| UC-6 | Admin memonitor status seluruh worker PHP melalui **GUI Admin Panel** (dashboard monitoring). | SysAdmin |

---

## 11. Metrik Keberhasilan (Success Metrics)

- **Zero konflik** dengan instalasi XAMPP/Laragon yang sudah berjalan (0 insiden bentrok port/file selama UAT).
- Waktu setup site baru (termasuk pemilihan versi PHP & extension) **< 5 menit** oleh admin.
- Downtime saat pergantian versi PHP/extension untuk 1 site **< 2 detik** (site lain 0 downtime).
- Uptime service produksi **≥ 99.9%** dalam periode observasi 30 hari pertama pasca rilis.
- Composer install/update via CLI berhasil dijalankan dengan tingkat error **< 1%** dibanding menjalankan langsung.

---

## 12. Risiko & Mitigasi

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Tidak tersedia library FastCGI resmi untuk Delphi | Perlu develop custom, risiko bug protokol | Alokasikan waktu R&D khusus, buat unit test protokol FastCGI secara ekstensif. |
| Konflik port tidak terdeteksi sempurna (misal aplikasi lain non-XAMPP/Laragon) | Service gagal start di production | Tambahkan pengecekan port generik (bukan hanya deteksi nama proses XAMPP/Laragon), beri pesan error yang actionable. |
| Extension PHP tidak kompatibel (TS/NTS, x86/x64) menyebabkan PHP worker gagal start | Downtime site | Validasi kompatibilitas sebelum apply (FR-32), mekanisme rollback otomatis (FR-34). |
| Perusahaan/tim tidak familiar Delphi untuk maintenance jangka panjang | Sulit cari talent maintenance | Dokumentasi arsitektur lengkap, kode modular dengan komentar jelas, pertimbangkan training internal. |
| Perubahan besar di versi PHP mendatang (PHP 9) mengubah cara build Windows binary | Perlu update Process Manager | Desain Process Manager agar versi-agnostic (baca path binary dari config, bukan hardcode). |

---

## 13. Asumsi & Ketergantungan

- Binary PHP resmi (dari windows.php.net) untuk setiap versi yang didukung tersedia dan sudah diuji kompatibel dengan mode FastCGI (`php-cgi.exe`).
- Binary **Nginx untuk Windows** resmi tersedia dan dapat dibundel bersama installer PHForge sebagai reverse-proxy front-facing.
- Composer (`composer.phar`) versi terbaru kompatibel dijalankan oleh semua versi PHP yang didukung (catatan: Composer versi terbaru mensyaratkan PHP minimal tertentu — perlu dicek per rilis).
- Server production memiliki akses untuk membuka port custom (tidak diblokir firewall internal tanpa konfigurasi ulang).
- Tim memiliki lisensi Delphi 10.3 yang sah beserta komponen tambahan yang dibutuhkan (Indy sudah termasuk bawaan; jika pakai komponen pihak ketiga lain perlu dicek lisensinya).

---

## 14. Roadmap (Fase Pengembangan)

| Fase | Cakupan | Estimasi |
|---|---|---|
| **Fase 1 — Core MVP** | Core HTTP Engine, Nginx reverse-proxy, multi-version PHP dispatch dasar, konfigurasi manual via file, isolasi dasar dari XAMPP/Laragon | 6–8 minggu |
| **Fase 2 — GUI Admin Panel & Extension Management** | GUI Admin Panel (VCL) untuk manajemen site, versi PHP, extension, service control, dan monitoring; dynamic extension management; hot-reload | 6–8 minggu |
| **Fase 3 — CLI Composer Tool** | CLI tool (`phforge`) khusus untuk menjalankan Composer multi-version | 2–3 minggu |
| **Fase 4 — Hardening Produksi** | Auto-restart, circuit breaker, audit log, health check, security hardening | 4 minggu |
| **Fase 5 — UAT & Rilis** | User Acceptance Testing di environment yang sudah ada XAMPP/Laragon terpasang, dokumentasi final | 2–3 minggu |

---

## 15. Pertanyaan Terbuka (Open Questions)

1. Apakah ada kebutuhan **multi-server/cluster** di masa depan (load balancing lintas mesin), yang perlu dipertimbangkan dalam desain arsitektur sekarang agar tidak perlu rewrite besar nanti?
2. Siapa yang akan menjadi **Composer proxy repository** internal (jika ada kebijakan tidak boleh akses internet langsung dari server produksi)?
3. Apakah dibutuhkan dukungan **Windows Authentication (NTLM/Kerberos)** untuk beberapa aplikasi internal?
4. Berapa lama **retention policy** log yang dibutuhkan oleh kebijakan perusahaan/compliance?

---

## 16. Lampiran

### 16.1 Contoh Struktur Konfigurasi Site (`sites.json`)

```json
{
  "sites": [
    {
      "name": "app-legacy",
      "domain": "legacy.internal.local",
      "port": 8080,
      "document_root": "C:/PHForge/sites/app-legacy/public",
      "php_version": "7.4",
      "extensions": ["mysqli", "mbstring", "gd"],
      "tls": false
    },
    {
      "name": "app-modern",
      "domain": "modern.internal.local",
      "port": 8443,
      "document_root": "C:/PHForge/sites/app-modern/public",
      "php_version": "8.3",
      "extensions": ["pdo_sqlsrv", "intl", "opcache", "redis"],
      "tls": true,
      "cert_path": "C:/PHForge/certs/app-modern.pfx"
    }
  ]
}
```

### 16.2 Contoh Perintah CLI (Khusus Composer)

> CLI `phforge` hanya digunakan untuk kebutuhan Composer. Manajemen site, versi PHP, extension, service control, dan monitoring seluruhnya dilakukan melalui **GUI Admin Panel** (lihat §6.7).

```bash
# Menjalankan composer install dengan versi PHP tertentu (mis. proyek Laravel)
phforge composer install --php=8.2 --path="D:\sites\myapp"

# Menjalankan composer update
phforge composer update --php=8.2 --path="D:\sites\myapp"

# Menambah dependency baru
phforge composer require vendor/package --php=8.2 --path="D:\sites\myapp"

# Menghapus dependency
phforge composer remove vendor/package --php=8.2 --path="D:\sites\myapp"

# Regenerasi autoload
phforge composer dump-autoload --php=8.2 --path="D:\sites\myapp"

# Memperbarui composer.phar itu sendiri
phforge composer self-update
```

### 16.3 Checklist Keamanan Berdampingan dengan XAMPP/Laragon

- [ ] Port HTTP/HTTPS berbeda dari default XAMPP/Laragon.
- [ ] Instance Nginx milik PHForge menggunakan port & nama service terpisah dari Nginx milik Laragon (jika Laragon dikonfigurasi memakai Nginx, bukan Apache).
- [ ] Direktori instalasi terpisah, tidak menimpa `C:\xampp` atau `C:\laragon`.
- [ ] Nama Windows Service unik, tidak bentrok dengan service Apache/Nginx yang sudah ada.
- [ ] Tidak mengubah PATH/PHPRC environment variable global.
- [ ] Tidak memodifikasi `hosts` file tanpa konfirmasi eksplisit.
- [ ] Registry key ditulis di path terpisah.
- [ ] Uninstaller tidak menyisakan residu yang mengganggu aplikasi lain.
- [ ] Firewall rule yang dibuat diberi nama jelas & dapat diidentifikasi sebagai milik produk ini.

---

*Dokumen ini adalah draft awal (v1.0) dan perlu direview bersama tim engineering, security, dan operasional sebelum masuk tahap desain teknis rinci.*
