<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use App\Models\StockOpname;
use App\Models\StockOpnamePage;
use App\Models\StockOpnameItem;
use Barryvdh\DomPDF\Facade\Pdf;
use SimpleSoftwareIO\QrCode\Facades\QrCode;
use Illuminate\Support\Str;

class FormController extends Controller
{
    public function generateDemoForm()
    {
        // 1. Create dummy session for demo
        $opname = StockOpname::create([
            'opname_date' => now(),
            'status' => 'draft',
            'total_items' => 5
        ]);

        $uuid = Str::uuid()->toString();
        
        $page = StockOpnamePage::create([
            'stock_opname_id' => $opname->id,
            'page_number' => 1,
            'qr_code_id' => $uuid,
        ]);

        $items = [
            ['code' => 'ITM-001', 'name' => 'Produk A', 'uom' => 'PCS'],
            ['code' => 'ITM-002', 'name' => 'Produk B', 'uom' => 'KG'],
            ['code' => 'ITM-003', 'name' => 'Produk C', 'uom' => 'BOX'],
            ['code' => 'ITM-004', 'name' => 'Produk D', 'uom' => 'PCS'],
            ['code' => 'ITM-005', 'name' => 'Produk E', 'uom' => 'LITER'],
        ];

        foreach ($items as $item) {
            StockOpnameItem::create([
                'stock_opname_page_id' => $page->id,
                'item_code' => $item['code'],
                'item_name' => $item['name'],
                'uom' => $item['uom'],
                'expected_qty' => rand(10, 100),
            ]);
        }

        // 2. Generate QR Code (Base64 for embedding in PDF)
        $qrCode = base64_encode(QrCode::format('svg')->size(100)->generate($uuid));

        // 3. Load View
        $pdf = Pdf::loadView('pdf.form', [
            'opname' => $opname,
            'page' => $page,
            'items' => $page->items,
            'qrCode' => $qrCode
        ])->setPaper('a4', 'portrait');

        return $pdf->stream('stock_opname_form.pdf');
    }
}
