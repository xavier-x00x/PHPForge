# PHPForge k6 Benchmark & Testing Suite (Trial 3)

Proyek ini dirancang khusus untuk menguji performa, konkurensi, dan keandalan antrean FastCGI Semaphore pada **PHPForge Web Server**.

Situs ini berjalan pada port **`8084`** (dapat dikonfigurasi melalui GUI Admin atau `dist/config/sites.json`).

---

## Daftar Endpoint Pengujian

| Endpoint | Tujuan Pengujian | Karakteristik Beban |
|---|---|---|
| **`index.php`** | Dashboard landing page & runtime telemetry | HTML rendering, OPcache inspection, Memory stats |
| **`api_simple.php`** | Throughput murni & latensi FastCGI baseline | Payload JSON minimal, CPU minimal (uji I/O murni) |
| **`api_transaction.php`** | Simulasi POS checkout & sinkronisasi kasir | Delay I/O database acak (25–75ms), kalkulasi PPN & struk |
| **`api_heavy.php`** | Pemrosesan logika berat (CPU intensive) | 2.500 kalkulasi SHA-256, array sorting, dan serialisasi JSON |
| **`api_post.php`** | HTTP POST Request handler | Parsing payload JSON, echo back, dan validasi header |

---

## Skenario Load Testing k6

Di dalam folder ini telah disediakan skrip pengujian k6 siap pakai:

1. **`k6_smoke.js` (Smoke Test)**
   - **Tujuan:** Verifikasi awal bahwa seluruh endpoint bekerja normal (200 OK) tanpa error.
   - **Beban:** 5 Virtual Users (VU) selama 15 detik.

2. **`k6_load.js` (Load Test Standar)**
   - **Tujuan:** Menguji performa dalam kondisi beban kerja kasir harian.
   - **Beban:** Naik hingga 50 Virtual Users selama 50 detik.

3. **`k6_stress.js` (High Concurrency Stress Test)**
   - **Tujuan:** Menguji kapasitas antrean *Channel Semaphore* dan *Connection Pool* FastCGI PHPForge.
   - **Beban:** Naik bertahap dari 100 &rarr; 200 &rarr; 300 Virtual Users.
   - **Indikator Sukses:** Error rate < 1% dan tidak ada HTTP 502 Bad Gateway.

4. **`k6_spike.js` (Spike Test)**
   - **Tujuan:** Menguji ketahanan server saat terjadi lonjakan trafik tiba-tiba (*rush hour*).
   - **Beban:** Lonjakan instan dari 10 ke 250 VU dalam 5 detik.

---

## Cara Menjalankan

### Opsi 1: Menggunakan Batch Launcher (Paling Mudah)
Klik dua kali pada file **`run_tests.bat`**, lalu pilih nomor tes yang ingin dijalankan (1–4).

### Opsi 2: Menggunakan Command Prompt / PowerShell
Buka terminal di folder ini dan jalankan perintah:

```cmd
# Jalankan Smoke Test:
k6 run k6_smoke.js

# Jalankan Load Test Standar:
k6 run k6_load.js

# Jalankan Stress Test:
k6 run k6_stress.js

# Jalankan Spike Test:
k6 run k6_spike.js
```

### Menguji URL Jaringan Intranet (Beda Mesin)
Jika Anda menguji dari komputer client / laptop lain di jaringan LAN, gunakan variabel environment `TARGET_URL`:

```cmd
k6 run -e TARGET_URL=http://192.168.1.100:8084 k6_stress.js
```

