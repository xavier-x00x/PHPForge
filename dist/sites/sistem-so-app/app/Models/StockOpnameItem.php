<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

class StockOpnameItem extends Model
{
    use HasFactory;

    protected $fillable = [
        'stock_opname_page_id', 
        'item_code', 
        'item_name', 
        'uom', 
        'expected_qty', 
        'actual_qty', 
        'confidence_level', 
        'snippet_image_path'
    ];

    public function page()
    {
        return $this->belongsTo(StockOpnamePage::class, 'stock_opname_page_id');
    }
}
