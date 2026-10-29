<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Inertia\Inertia;
use App\Jobs\ProcessScannedForm;

class UploadController extends Controller
{
    public function show()
    {
        return Inertia::render('Upload');
    }

    public function store(Request $request)
    {
        $request->validate([
            'file' => 'required|mimes:jpeg,png,jpg,pdf|max:10240',
        ]);

        if ($request->file('file')) {
            $path = $request->file('file')->store('public/scans');
            
            // Dispatch job to process the file in background
            ProcessScannedForm::dispatch(storage_path('app/' . $path));
            
            return redirect()->back()->with('message', 'File uploaded successfully and is being processed!');
        }

        return back()->withErrors(['file' => 'Upload failed.']);
    }
}
