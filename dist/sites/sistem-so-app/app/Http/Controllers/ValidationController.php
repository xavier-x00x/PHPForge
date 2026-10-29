<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Inertia\Inertia;
use App\Models\StockOpnamePage;
use App\Models\StockOpnameItem;

class ValidationController extends Controller
{
    public function show($pageId)
    {
        $page = StockOpnamePage::with('items')->findOrFail($pageId);
        
        return Inertia::render('Validation', [
            'page' => $page,
            'items' => $page->items
        ]);
    }

    public function update(Request $request, $itemId)
    {
        $request->validate([
            'actual_qty' => 'required|numeric',
        ]);

        $item = StockOpnameItem::findOrFail($itemId);
        $item->actual_qty = $request->actual_qty;
        $item->save();

        return redirect()->back()->with('message', 'Item updated successfully');
    }
}
