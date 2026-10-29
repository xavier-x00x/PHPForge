<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

class StockOpnamePage extends Model
{
    use HasFactory;

    protected $fillable = ['stock_opname_id', 'page_number', 'qr_code_id', 'image_path', 'status'];

    public function opname()
    {
        return $this->belongsTo(StockOpname::class, 'stock_opname_id');
    }

    public function items()
    {
        return $this->hasMany(StockOpnameItem::class);
    }
}
