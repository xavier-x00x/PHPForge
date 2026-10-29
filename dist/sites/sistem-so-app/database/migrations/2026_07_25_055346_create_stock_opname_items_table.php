<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

class CreateStockOpnameItemsTable extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('stock_opname_items', function (Blueprint $table) {
            $table->id();
            $table->foreignId('stock_opname_page_id')->constrained('stock_opname_pages')->onDelete('cascade');
            $table->string('item_code');
            $table->string('item_name');
            $table->string('uom');
            $table->integer('expected_qty');
            $table->integer('actual_qty')->nullable();
            $table->decimal('confidence_level', 5, 2)->nullable();
            $table->string('snippet_image_path')->nullable();
            $table->timestamps();
        });
    }

    /**
     * Reverse the migrations.
     *
     * @return void
     */
    public function down()
    {
        Schema::dropIfExists('stock_opname_items');
    }
}
