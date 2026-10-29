<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <title>Stock Opname Form</title>
    <style>
        @page { margin: 0px; }
        body { margin: 0px; font-family: sans-serif; padding: 40px; }
        /* Anchor Points (Fiducial Markers) */
        .anchor {
            position: absolute;
            width: 30px;
            height: 30px;
            background-color: black;
        }
        .anchor-tl { top: 20px; left: 20px; }
        .anchor-tr { top: 20px; right: 20px; }
        .anchor-bl { bottom: 20px; left: 20px; }
        .anchor-br { bottom: 20px; right: 20px; }

        .header {
            text-align: center;
            margin-top: 30px;
            margin-bottom: 30px;
        }
        .qr-code {
            position: absolute;
            top: 50px;
            right: 60px;
            width: 100px;
            height: 100px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }
        th, td {
            border: 1px solid #000;
            padding: 4px;
            text-align: left;
        }
        th {
            background-color: #f2f2f2;
        }
        /* Segmented Box for QTY */
        .qty-box {
            display: inline-block;
            width: 20px;
            height: 22px;
            border: 1px solid #000;
            margin-right: 5px;
        }
    </style>
</head>
<body>
    <!-- 4 Anchor Points -->
    <div class="anchor anchor-tl"></div>
    <div class="anchor anchor-tr"></div>
    <div class="anchor anchor-bl"></div>
    <div class="anchor anchor-br"></div>

    <div class="qr-code">
        <img src="data:image/svg+xml;base64,{{ $qrCode }}" alt="QR Code">
    </div>

    <div class="header">
        <h2>FORM STOCK OPNAME</h2>
        <p>Tanggal: {{ $opname->opname_date }} | Halaman: {{ $page->page_number }}</p>
        <p>ID: {{ $page->qr_code_id }}</p>
    </div>

    <table>
        <thead>
            <tr>
                <th>No</th>
                <th>Kode Barang</th>
                <th>Nama Barang</th>
                <th>UoM</th>
                <th style="width: 210px; text-align: center;">Qty Aktual</th>
            </tr>
        </thead>
        <tbody>
            @foreach($items as $index => $item)
            <tr>
                <td>{{ $index + 1 }}</td>
                <td>{{ $item->item_code }}</td>
                <td>{{ $item->item_name }}</td>
                <td>{{ $item->uom }}</td>
                <td style="white-space: nowrap;">
                    <!-- Bagian ribuan/ratusan (5 digit max) -->
                    <div class="qty-box"></div>
                    <div class="qty-box"></div>
                    <div class="qty-box"></div>
                    <div class="qty-box"></div>
                    <div class="qty-box"></div>
                    
                    <!-- Separator desimal -->
                    <span style="font-size: 20px; font-weight: bold; margin-left: 2px; margin-right: 2px;">,</span>
                    
                    <!-- Bagian desimal (3 digit max) -->
                    <div class="qty-box"></div>
                    <div class="qty-box"></div>
                    <div class="qty-box"></div>
                </td>
            </tr>
            @endforeach
        </tbody>
    </table>
</body>
</html>
