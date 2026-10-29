<?php

use Illuminate\Support\Facades\Route;

/*
|--------------------------------------------------------------------------
| Web Routes
|--------------------------------------------------------------------------
|
| Here is where you can register web routes for your application. These
| routes are loaded by the RouteServiceProvider within a group which
| contains the "web" middleware group. Now create something great!
|
*/

use Inertia\Inertia;
use App\Http\Controllers\FormController;
use App\Http\Controllers\UploadController;
use App\Http\Controllers\ValidationController;

Route::get('/', function () {
    return Inertia::render('Index', [
        'name' => 'Admin'
    ]);
});

Route::get('/form/demo', [FormController::class, 'generateDemoForm']);

Route::get('/upload', [UploadController::class, 'show']);
Route::post('/upload', [UploadController::class, 'store']);

Route::get('/validation/{pageId}', [ValidationController::class, 'show']);
Route::put('/validation/item/{itemId}', [ValidationController::class, 'update']);
