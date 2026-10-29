<?php
header('Content-Type: text/html; charset=UTF-8');
$startTime = microtime(true);
$phpVersion = phpversion();
$memoryUsage = round(memory_get_usage() / 1024 / 1024, 2);
$memoryPeak = round(memory_get_peak_usage() / 1024 / 1024, 2);
$opcacheEnabled = function_exists('opcache_get_status') && opcache_get_status() !== false;
$extensions = get_loaded_extensions();
?>
<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PHPForge Load Testing Suite - Trial 3</title>
    <style>
        :root {
            --bg: #0f172a;
            --card: #1e293b;
            --border: #334155;
            --primary: #38bdf8;
            --success: #34d399;
            --warning: #fbbf24;
            --text: #f8fafc;
            --text-muted: #94a3b8;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
            background-color: var(--bg);
            color: var(--text);
            line-height: 1.6;
            padding: 2rem;
        }
        .container { max-width: 1000px; margin: 0 auto; }
        .header {
            margin-bottom: 2rem;
            border-bottom: 1px solid var(--border);
            padding-bottom: 1.5rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .badge {
            background: rgba(56, 189, 248, 0.15);
            color: var(--primary);
            border: 1px solid rgba(56, 189, 248, 0.3);
            padding: 0.25rem 0.75rem;
            border-radius: 9999px;
            font-size: 0.85rem;
            font-family: monospace;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2rem;
        }
        .card {
            background: var(--card);
            border: 1px solid var(--border);
            border-radius: 12px;
            padding: 1.5rem;
            transition: transform 0.2s, border-color 0.2s;
        }
        .card:hover {
            transform: translateY(-2px);
            border-color: var(--primary);
        }
        h1 { font-size: 1.8rem; font-weight: 800; color: #fff; }
        h2 { font-size: 1.2rem; margin-bottom: 0.75rem; color: #fff; }
        p { color: var(--text-muted); font-size: 0.9rem; margin-bottom: 1rem; }
        .stats {
            display: flex;
            gap: 1rem;
            font-family: monospace;
            font-size: 0.85rem;
            color: var(--text-muted);
            margin-top: 0.5rem;
        }
        .btn {
            display: inline-block;
            background: rgba(56, 189, 248, 0.15);
            color: var(--primary);
            border: 1px solid rgba(56, 189, 248, 0.3);
            padding: 0.5rem 1rem;
            border-radius: 6px;
            text-decoration: none;
            font-size: 0.85rem;
            font-weight: 600;
            transition: all 0.2s;
        }
        .btn:hover {
            background: var(--primary);
            color: #000;
        }
        .tag {
            font-size: 0.75rem;
            padding: 0.15rem 0.5rem;
            border-radius: 4px;
            background: #334155;
            color: #e2e8f0;
            margin-right: 0.25rem;
            display: inline-block;
        }
        .code-box {
            background: #090d16;
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 1rem;
            font-family: monospace;
            font-size: 0.85rem;
            color: #38bdf8;
            overflow-x: auto;
            margin-top: 1rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div>
                <h1>PHPForge Target Server (Trial 3)</h1>
                <p>Proyek uji performa & stres untuk load test k6 pada PHPForge Web Server</p>
            </div>
            <div class="badge">PHP <?php echo $phpVersion; ?> | Port 8084</div>
        </div>

        <div class="grid">
            <div class="card">
                <h2>1. FastCGI Baseline (Simple)</h2>
                <p>Endpoint ringan tanpa query berat. Digunakan untuk menguji batas throughput maksimal server dan latensi murni FastCGI socket pool.</p>
                <div class="stats">
                    <span>Target: <code>api_simple.php</code></span>
                </div>
                <div style="margin-top: 1rem;">
                    <a href="api_simple.php" class="btn" target="_blank">Buka Endpoint &rarr;</a>
                </div>
            </div>

            <div class="card">
                <h2>2. Simulasi Transaksi POS (I/O Delay)</h2>
                <p>Mensimulasikan beban kasir/retail dengan latensi database (30ms - 80ms) dan pembuatan nomor transaksi acak.</p>
                <div class="stats">
                    <span>Target: <code>api_transaction.php</code></span>
                </div>
                <div style="margin-top: 1rem;">
                    <a href="api_transaction.php" class="btn" target="_blank">Buka Endpoint &rarr;</a>
                </div>
            </div>

            <div class="card">
                <h2>3. CPU & Memory Intensive</h2>
                <p>Mensimulasikan pemrosesan data berat: enkripsi SHA-256, komputasi matriks/array 2.000 data, dan serialisasi JSON.</p>
                <div class="stats">
                    <span>Target: <code>api_heavy.php</code></span>
                </div>
                <div style="margin-top: 1rem;">
                    <a href="api_heavy.php" class="btn" target="_blank">Buka Endpoint &rarr;</a>
                </div>
            </div>

            <div class="card">
                <h2>4. POST Request Payload</h2>
                <p>Menerima data JSON melalui HTTP POST (misal sync keranjang belanja dari cabang toko) dan mengembalikan respons validasi.</p>
                <div class="stats">
                    <span>Target: <code>api_post.php</code></span>
                </div>
                <div style="margin-top: 1rem;">
                    <a href="api_post.php" class="btn" target="_blank">Buka Endpoint &rarr;</a>
                </div>
            </div>
        </div>

        <div class="card" style="margin-bottom: 2rem;">
            <h2>Cara Menjalankan k6 Load Test</h2>
            <p>Buka Command Prompt atau PowerShell di folder ini, lalu jalankan script k6 yang sudah disediakan:</p>
            <div class="code-box">
# Menjalankan load test standar (50 Virtual Users):
k6 run k6_load.js

# Menjalankan high-concurrency stress test (200 - 300 Virtual Users):
k6 run k6_stress.js

# Atau cukup klik dua kali pada file run_tests.bat
            </div>
        </div>

        <div class="card">
            <h2>Status Runtime Engine</h2>
            <div class="stats" style="flex-wrap: wrap; gap: 1.5rem;">
                <div>Memory Saat Ini: <strong><?php echo $memoryUsage; ?> MB</strong></div>
                <div>Memory Peak: <strong><?php echo $memoryPeak; ?> MB</strong></div>
                <div>OPcache: <strong><?php echo $opcacheEnabled ? 'Aktif' : 'Non-aktif'; ?></strong></div>
                <div>Waktu Response: <strong><?php echo round((microtime(true) - $startTime) * 1000, 2); ?> ms</strong></div>
            </div>
        </div>
    </div>
</body>
</html>

