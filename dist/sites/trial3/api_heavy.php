<?php
header('Content-Type: application/json');

$startTime = microtime(true);

// 1. Pemrosesan data komputasi array (2,500 record)
$data = [];
for ($i = 0; $i < 2500; $i++) {
    $data[] = [
        'id' => $i,
        'hash' => hash('sha256', 'item_secret_key_' . $i . '_' . microtime()),
        'score' => rand(1, 10000) / 100
    ];
}

// 2. Sorting data berdasarkan score
usort($data, function($a, $b) {
    return $b['score'] <=> $a['score'];
});

// 3. Ambil top 10
$top10 = array_slice($data, 0, 10);

$endTime = microtime(true);
$executionMs = round(($endTime - $startTime) * 1000, 2);

echo json_encode([
    'status' => 'success',
    'endpoint' => 'api_heavy',
    'processed_records' => count($data),
    'top_sample' => $top10[0],
    'execution_time_ms' => $executionMs,
    'memory_peak_mb' => round(memory_get_peak_usage() / 1024 / 1024, 2),
    'php_version' => phpversion()
]);

