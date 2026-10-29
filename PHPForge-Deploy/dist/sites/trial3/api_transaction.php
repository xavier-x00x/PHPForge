<?php
header('Content-Type: application/json');

$startTime = microtime(true);

// Tangkap parameter (jika ada) atau gunakan default
$storeId = isset($_GET['store_id']) ? htmlspecialchars($_GET['store_id']) : 'CABANG-' . rand(101, 110);
$cashierId = isset($_GET['cashier_id']) ? htmlspecialchars($_GET['cashier_id']) : 'KSR-' . rand(1, 20);

// Simulasi latensi database I/O (misal query SELECT promo & INSERT transaksi kasir)
// 25ms - 75ms adalah representasi realistis transaksi database intranet
$dbLatencyMs = rand(25, 75);
usleep($dbLatencyMs * 1000);

// Simulasi pembuatan item belanja
$itemsCount = rand(2, 8);
$items = [];
$subtotal = 0;

for ($i = 1; $i <= $itemsCount; $i++) {
    $qty = rand(1, 5);
    $price = rand(5000, 75000);
    $lineTotal = $qty * $price;
    $subtotal += $lineTotal;

    $items[] = [
        'item_sku' => 'SKU-' . rand(10000, 99999),
        'qty' => $qty,
        'unit_price' => $price,
        'line_total' => $lineTotal
    ];
}

$tax = round($subtotal * 0.11); // PPN 11%
$grandTotal = $subtotal + $tax;

$endTime = microtime(true);
$totalExecutionMs = round(($endTime - $startTime) * 1000, 2);

echo json_encode([
    'status' => 'success',
    'transaction_id' => 'TRX-' . date('YmdHis') . '-' . strtoupper(substr(bin2hex(random_bytes(3)), 0, 6)),
    'store_id' => $storeId,
    'cashier_id' => $cashierId,
    'total_items' => count($items),
    'subtotal' => $subtotal,
    'tax_ppn_11' => $tax,
    'grand_total' => $grandTotal,
    'simulated_db_latency_ms' => $dbLatencyMs,
    'server_execution_ms' => $totalExecutionMs,
    'server_time' => date('Y-m-d H:i:s'),
    'php_version' => phpversion()
]);

