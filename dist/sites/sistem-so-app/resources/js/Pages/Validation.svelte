<script>
  import { Inertia } from '@inertiajs/inertia'

  export let page;
  export let items = [];

  // Initialize draft quantities
  $: items = items.map(i => ({ ...i, draft_qty: i.draft_qty !== undefined ? i.draft_qty : (i.actual_qty !== null ? i.actual_qty : '') }));

  let selectedItem = null;
  let savingIds = new Set();

  function selectItem(item) {
    selectedItem = item;
  }

  function saveItem(item) {
    if (savingIds.has(item.id)) return;
    
    // Don't save if the value hasn't actually changed
    if (item.draft_qty === item.actual_qty) return;
    
    savingIds.add(item.id);
    savingIds = savingIds; // trigger reactivity if needed

    Inertia.put(`/validation/item/${item.id}`, {
      actual_qty: item.draft_qty
    }, {
      preserveScroll: true,
      onFinish: () => {
        savingIds.delete(item.id);
        savingIds = savingIds;
      },
      onSuccess: () => {
        // Find index and update actual_qty to reflect saved state
        let index = items.findIndex(i => i.id === item.id);
        if (index !== -1) {
          items[index].actual_qty = item.draft_qty;
          items = [...items]; // trigger reactivity
        }
      }
    });
  }

  function handleKeydown(e, index, item) {
    if (e.key === 'Enter') {
      e.preventDefault();
      saveItem(item); // Force save on enter just in case change event doesn't fire
      
      const nextInput = document.getElementById(`qty-input-${index + 1}`);
      if (nextInput) {
        nextInput.focus();
      } else {
        e.target.blur();
      }
    }
  }
</script>

<div class="page">
  <div class="page-header d-print-none">
    <div class="container-xl">
      <div class="row g-2 align-items-center">
        <div class="col">
          <h2 class="page-title">
            Validation - Page {page.page_number}
          </h2>
        </div>
      </div>
    </div>
  </div>
  
  <div class="page-body">
    <div class="container-xl">
      


      <div class="row">
        <!-- Full Width Table -->
        <div class="col-12">
          <div class="card">
            <div class="table-responsive">
              <table class="table table-vcenter table-sm card-table">
                <thead>
                  <tr>
                    <th>Item Code</th>
                    <th>Item Name</th>
                    <th>Confidence</th>
                    <th>Actual Qty (OCR)</th>
                    <th>Snippet</th>
                    <th>Corrected Quantity</th>
                  </tr>
                </thead>
                <tbody>
                  {#each items as item, index}
                    <tr class={selectedItem && selectedItem.id === item.id ? 'bg-muted-lt' : ''}>
                      <td>{item.item_code}</td>
                      <td>{item.item_name}</td>
                      <td>
                        {#if item.confidence_level}
                          <div class="progress" style="width: 50px">
                            <div class="progress-bar {item.confidence_level < 80 ? 'bg-danger' : 'bg-success'}" style="width: {item.confidence_level}%"></div>
                          </div>
                        {/if}
                      </td>
                      <td>
                        {#if item.actual_qty !== null}
                          <span class="badge bg-success">{item.actual_qty}</span>
                        {:else}
                          <span class="badge bg-warning">Pending</span>
                        {/if}
                      </td>
                      <td class="text-center">
                        {#if item.snippet_image_path}
                          <img src="/storage/{item.snippet_image_path}" alt="Snippet" style="height: 36px; object-fit: contain; border: 1px solid #ccc; background: white;" />
                        {:else}
                          <span class="text-muted small">No Image</span>
                        {/if}
                      </td>
                      <td>
                        <div class="d-flex align-items-center gap-2">
                          <input type="number" step="0.001" class="form-control form-control-sm" style="width: 150px;"
                            id="qty-input-{index}"
                            bind:value={item.draft_qty} 
                            on:focus={() => selectItem(item)} 
                            on:change={() => saveItem(item)}
                            on:keydown={(e) => handleKeydown(e, index, item)} />
                          
                          {#if savingIds.has(item.id)}
                            <span class="spinner-border spinner-border-sm text-primary" role="status"></span>
                          {:else if item.actual_qty === item.draft_qty && item.actual_qty !== null}
                            <svg xmlns="http://www.w3.org/2000/svg" class="icon text-success" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M5 12l5 5l10 -10" /></svg>
                          {/if}
                        </div>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
