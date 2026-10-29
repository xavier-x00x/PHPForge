# Product Requirements Document (PRD): Sistem Stock Opname (SO) Cerdas Berbasis Pemindaian

## 1. Latar Belakang & Pernyataan Masalah
Proses *Stock Opname* (SO) saat ini dilakukan di dalam ruang pendingin (*cold storage*). Kondisi lingkungan yang ekstrem (suhu sangat rendah, kelembapan) membuat penggunaan perangkat elektronik (tablet/smartphone) menjadi tidak memungkinkan karena risiko kerusakan perangkat dan baterai yang cepat habis.

**Kondisi Saat Ini (Pain Points):**
*   **Ketergantungan pada Kertas:** Pencatatan SO harus menggunakan formulir cetak.
*   **Redundansi Kerja (Double Data Entry):** Staf harus menyalin ulang data dari kertas ke dalam sistem ERP/komputer secara manual satu per satu.
*   **Rentan *Human Error*:** Proses entri ulang sangat rentan terhadap kesalahan ketik (salah baca tulisan tangan atau kelewatan baris).
*   **Inefisiensi Waktu:** Proses rekapitulasi memakan waktu lama, memperlambat penyelesaian laporan SO.

## 2. Tujuan & Sasaran (Objectives & Goals)
*   Mengeliminasi proses *double data entry* manual secara keseluruhan.
*   Meningkatkan akurasi dan kecepatan rekapitulasi data SO hingga 80%.
*   Menyediakan sistem validasi visual yang presisi, di mana pengguna dapat membandingkan hasil bacaan sistem (OCR) dengan potongan gambar asli dari kertas pada baris yang sama.

## 3. Alur Kerja yang Diusulkan (Proposed User Journey)

1.  **Pembuatan & Pencetakan Form:** Sistem Web menghasilkan form SO yang dilengkapi dengan *QR Code* (sebagai ID Dokumen/Batch) dan *Fiducial Markers* (titik jangkar di sudut kertas untuk kalibrasi kemiringan scan). Baris kuantitas (Qty) disediakan dalam format kotak-kotak (*segmented boxes*) untuk memaksimalkan akurasi pembacaan tulisan tangan.
2.  **Pelaksanaan SO (Manual):** Staf gudang membawa kertas ke ruang pendingin dan mengisi jumlah fisik barang di dalam kotak yang disediakan menggunakan pulpen/spidol.
3.  **Pemindaian Dokumen (Scanning):** Setelah selesai, kertas di-scan menggunakan *scanner* dokumen atau difoto secara tegak lurus. File (PDF/JPG) diunggah ke dalam Aplikasi Web SO.
4.  **Ekstraksi Data Otomatis (OCR/ICR):** Sistem membaca QR Code untuk mengetahui daftar barang apa saja yang ada di form tersebut. Kemudian, sistem melakukan *cropping* pada setiap baris Qty dan menjalankan mesin ICR (*Intelligent Character Recognition*) untuk mengekstrak angka tulisan tangan.
5.  **Validasi & Koreksi (Review):** Sistem menampilkan antarmuka validasi. Untuk setiap baris barang, sistem menyandingkan kolom *input digital* dengan *potongan gambar (snippet)* tulisan tangan staf dari baris tersebut. Pengguna memverifikasi dan memperbaiki angka jika pembacaan sistem meleset.
6.  **Sinkronisasi ke Database:** Setelah semua baris divalidasi dan disetujui, data SO dikunci dan diperbarui ke dalam sistem inventaris/ERP utama.

---

## 4. Kebutuhan Fungsional (Functional Requirements)

| Fitur Utama | Deskripsi & Aturan Bisnis |
| :--- | :--- |
| **Form Generator Engine** | Menghasilkan PDF form SO. Wajib menyertakan:<br/>1. **QR Code** unik per halaman (menyimpan ID SO, ID Halaman).<br/>2. **Anchor Points** (kotak/lingkaran di 4 sudut halaman).<br/>3. **Tabel Item** (Kode Brg, Nama, UoM, dan kotak kosong untuk Qty aktual). |
| **Document Uploader** | Mendukung unggah multi-halaman (PDF, JPG, PNG). Memiliki indikator *loading* saat proses ekstraksi OCR sedang berjalan di *background*. |
| **OCR & Image Processing** | Sistem harus mampu:<br/>1. Melakukan *deskew* (meluruskan gambar miring) menggunakan *anchor points*.<br/>2. Memotong (*slice*) gambar per baris item.<br/>3. Mengenali angka tulisan tangan (ICR) pada area *bounding box* Qty. |
| **Split-View Validation UI** | UI menampilkan daftar item dalam bentuk tabel/grid. Wajib menampilkan potongan gambar asli berukuran kecil (*thumbnail*) tepat di sebelah *field input* angka hasil OCR. Jika di-klik, gambar bisa membesar (*zoom*). |
| **Confidence Level Indicator** | Sistem memberikan tanda warna pada *field input*. Merah/Kuning jika tingkat keyakinan OCR rendah (wajib di-cek teliti), Hijau jika OCR sangat yakin (misal >95%). |
| **Audit Trail & Save Draft** | Pengguna dapat menyimpan *draft* validasi dan melanjutkannya nanti. Gambar asli harus tetap tersimpan sebagai bukti audit historis jika ada selisih stok di kemudian hari. |

## 5. Pertimbangan Teknis & Arsitektur

*   **Desain Formulir (OMR/ICR Friendly):** Tulisan tangan bebas sangat sulit dibaca secara akurat. Desain form **harus** menggunakan format kotak terpisah untuk setiap digit (misal: `[ ] [ ] [ ] [ ]`) agar mesin pembaca tahu letak pasti setiap karakter.
*   **Pemilihan Mesin Ekstraksi (OCR/ICR API):** Karena ini membaca angka tulisan tangan (*Handwriting Recognition*), penggunaan *engine* standar kadang kurang memadai. Perlu dipertimbangkan layanan berbasis *Cloud Vision* (seperti Google Cloud Vision API, AWS Textract) atau melatih model *Machine Learning* kustom ringan (seperti Tesseract yang di-*fine-tune*, OpenCV, atau YOLO untuk deteksi *bounding box*).
*   **Frontend Performance:** Proses *slicing* gambar (memotong satu gambar utuh menjadi puluhan *snippet* per baris) sebaiknya diproses di *backend* dan disajikan ke antarmuka web dalam bentuk koordinat atau URL gambar kecil terpisah agar web tidak *lagging* saat merender ratusan baris di halaman *review*.

> **Tindak Lanjut:** Tentukan Engine Ekstraksi Visual. Analisis lebih lanjut diperlukan untuk merekomendasikan OCR atau Computer Vision engine yang paling cocok untuk membaca angka tulisan tangan pada sistem ini (cloud maupun on-premise).
