<?php
header('Content-Type: application/json');

// Catat waktu mulai
$start = microtime(true);

echo json_encode([
    'status' => 'ok',
    'endpoint' => 'api_simple',
    'timestamp' => microtime(true),
    'php_version' => phpversion(),
    'duration_ms' => round((microtime(true) - $start) * 1000, 3)
]);

