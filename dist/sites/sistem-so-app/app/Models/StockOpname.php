<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

class StockOpname extends Model
{
    use HasFactory;

    protected $fillable = ['opname_date', 'status', 'total_items'];

    public function pages()
    {
        return $this->hasMany(StockOpnamePage::class);
    }
}
