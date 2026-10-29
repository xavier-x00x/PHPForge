<?php

namespace App\Jobs;

use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldBeUnique;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Log;

class ProcessScannedForm implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    protected $filePath;

    /**
     * Create a new job instance.
     *
     * @return void
     */
    public function __construct($filePath)
    {
        $this->filePath = $filePath;
    }

    /**
     * Execute the job.
     *
     * @return void
     */
    public function handle()
    {
        Log::info("Processing scanned form: " . $this->filePath);
        
        $scriptPath = storage_path('app/scripts/process_image.py');
        $pythonExe = base_path('pyenv/Scripts/python.exe');
        $command = escapeshellarg($pythonExe) . " " . escapeshellarg($scriptPath) . " " . escapeshellarg($this->filePath);
        
        $output = shell_exec($command);
        $result = json_decode($output, true);

        if (!$result || !isset($result['status']) || $result['status'] !== 'success') {
            Log::error("Failed to process image: " . $output);
            return;
        }

        Log::info("Python script success: " . $result['message']);

        // Since we are mocking the QR Code ID for demonstration, 
        // we will just take the first pending StockOpnamePage from the database.
        // In a real scenario, you would use:
        // $page = \App\Models\StockOpnamePage::where('qr_code_id', $result['qr_code_id'])->first();
        
        $page = \App\Models\StockOpnamePage::where('status', 'pending')->first();
        
        if ($page) {
            $items = $page->items()->orderBy('id')->get();
            
            // Generate fallback snippets using PHP GD if possible
            $extension = strtolower(pathinfo($this->filePath, PATHINFO_EXTENSION));
            $useGDSnippets = false;
            $gdSnippets = [];
            
            if (in_array($extension, ['jpg', 'jpeg', 'png'])) {
                try {
                    $img = $extension === 'png' ? imagecreatefrompng($this->filePath) : imagecreatefromjpeg($this->filePath);
                    if ($img) {
                        $width = imagesx($img);
                        $height = imagesy($img);
                        $sliceHeight = floor($height / 5);
                        
                        $baseName = basename($this->filePath, '.' . $extension);
                        
                        for ($i = 0; $i < 5; $i++) {
                            $snippet = imagecrop($img, ['x' => 0, 'y' => $i * $sliceHeight, 'width' => $width, 'height' => $sliceHeight]);
                            $snippetPathAbsolute = storage_path('app/public/scans/' . $baseName . '_snippet_' . $i . '.jpg');
                            imagejpeg($snippet, $snippetPathAbsolute, 90);
                            $gdSnippets[$i] = 'scans/' . $baseName . '_snippet_' . $i . '.jpg';
                            imagedestroy($snippet);
                        }
                        imagedestroy($img);
                        $useGDSnippets = true;
                    }
                } catch (\Exception $e) {
                    Log::error("GD Cropping failed: " . $e->getMessage());
                }
            }
            
            $relativePath = 'scans/' . basename($this->filePath);
            
            // Loop through the results from the Python script
            foreach ($result['items'] as $extractedItem) {
                $index = $extractedItem['item_index'];
                if (isset($items[$index])) {
                    $itemModel = $items[$index];
                    $itemModel->actual_qty = $extractedItem['actual_qty'];
                    $itemModel->confidence_level = $extractedItem['confidence'];
                    
                    // Use Python snippet, or GD snippet, or fallback to full image
                    if (isset($extractedItem['snippet_path'])) {
                        $itemModel->snippet_image_path = $extractedItem['snippet_path'];
                    } elseif ($useGDSnippets && isset($gdSnippets[$index])) {
                        $itemModel->snippet_image_path = $gdSnippets[$index];
                    } else {
                        $itemModel->snippet_image_path = $relativePath;
                    }
                    
                    $itemModel->save();
                }
            }
            
            $page->status = 'processed';
            $page->save();
            Log::info("Page ID {$page->id} processed successfully.");
        } else {
            Log::warning("No pending StockOpnamePage found to update.");
        }
    }
}
