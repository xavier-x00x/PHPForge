<?php
header('Content-Type: application/json');

$startTime = microtime(true);

if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    http_response_code(405);
    echo json_encode([
        'status' => 'error',
        'message' => 'Method Not Allowed. Silakan gunakan method POST.'
    ]);
    exit;
}

$rawInput = file_get_contents('php://input');
$payload = json_decode($rawInput, true);

if (!$payload) {
    // Fallback ke $_POST jika form-data biasa
    $payload = $_POST;
}

$receivedKeys = is_array($payload) ? array_keys($payload) : [];
$payloadSize = strlen($rawInput);

$endTime = microtime(true);

echo json_encode([
    'status' => 'success',
    'message' => 'Payload POST berhasil diproses oleh FastCGI worker',
    'payload_bytes' => $payloadSize,
    'fields_received' => $receivedKeys,
    'echo_sample' => is_array($payload) ? array_slice($payload, 0, 3) : null,
    'execution_ms' => round(($endTime - $startTime) * 1000, 2),
    'worker_pid' => getmypid(),
    'php_version' => phpversion()
]);

