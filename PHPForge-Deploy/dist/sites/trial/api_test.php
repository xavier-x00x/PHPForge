<?php
// Set header sebagai JSON (seolah-olah ini API dari aplikasi Kasir/POS)
header('Content-Type: application/json');

// Catat waktu mulai
$startTime = microtime(true);

// Mensimulasikan pemrosesan logika berat atau query database
// Misalnya: POS sedang mencocokkan kode barcode, menghitung diskon grosir, dan PPN.
// Kita gunakan usleep (micro-sleep) antara 50ms - 150ms untuk simulasi yang realistis.
$delay_ms = rand(50, 150);
usleep($delay_ms * 1000); 

// Mensimulasikan konsumsi CPU dan RAM untuk manipulasi array/data
$dummyData = [];
for ($i = 0; $i < 1000; $i++) {
    $dummyData[] = [
        'item_id' => rand(1000, 9999),
        'qty' => rand(1, 10),
        'price' => rand(1000, 100000)
    ];
}

$total = array_reduce($dummyData, function($carry, $item) {
    return $carry + ($item['qty'] * $item['price']);
}, 0);

// Catat waktu selesai
$endTime = microtime(true);
$executionTime = round(($endTime - $startTime) * 1000, 2);

// Kembalikan respons API
echo json_encode([
    'status' => 'success',
    'message' => 'Transaksi POS berhasil disinkronisasi',
    'simulated_delay_ms' => $delay_ms,
    'execution_time_ms' => $executionTime,
    'total_amount' => $total,
    'php_version' => phpversion()
]);
