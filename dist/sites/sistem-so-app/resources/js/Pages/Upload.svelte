<script>
  import { useForm } from '@inertiajs/inertia-svelte'

  let form = useForm({
    file: null,
  })

  function submit() {
    $form.post('/upload', {
      forceFormData: true,
      onSuccess: () => alert('File uploaded successfully and is being processed in background!'),
    })
  }

  function handleFileChange(event) {
    $form.file = event.target.files[0]
  }
</script>

<div class="page">
  <div class="page-header d-print-none">
    <div class="container-xl">
      <div class="row g-2 align-items-center">
        <div class="col">
          <h2 class="page-title">
            Upload Scanned Form
          </h2>
        </div>
      </div>
    </div>
  </div>
  <div class="page-body">
    <div class="container-xl">
      <div class="card">
        <div class="card-header">
          <h3 class="card-title">Upload Stock Opname Form (Image/PDF)</h3>
        </div>
        <div class="card-body">
          <form on:submit|preventDefault={submit}>
            <div class="mb-3">
              <label class="form-label" for="fileInput">Select File</label>
              <input type="file" id="fileInput" class="form-control" on:change={handleFileChange} accept="image/jpeg,image/png,application/pdf" required />
            </div>
            
            {#if $form.progress}
              <div class="progress mb-3">
                <div class="progress-bar" style="width: {$form.progress.percentage}%"></div>
              </div>
            {/if}

            <div class="form-footer">
              <button type="submit" class="btn btn-primary" disabled={$form.processing}>
                {#if $form.processing}
                  <span class="spinner-border spinner-border-sm me-2" role="status"></span>
                  Uploading...
                {:else}
                  Upload & Process
                {/if}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</div>
