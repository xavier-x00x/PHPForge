<script setup>
import { ref, onMounted } from 'vue'
import { GetServerConfig, SaveServerConfig, OpenPathInExplorer, GetServiceInfo } from '../../wailsjs/go/main/App'
import { useToast } from '../composables/useToast'

const toast = useToast()

const config = ref({
  api_port: 8081,
  php_workers: 16,
  data_dir: '',
  api_key: 'phforge-secret-key'
})

const loading = ref(true)
const saving = ref(false)
const showApiKey = ref(false)
const serviceInfo = ref({ installed: false, running: false, raw: '' })

const fetchConfig = async () => {
  loading.value = true
  try {
    const res = await GetServerConfig()
    if (res) {
      config.value = res
    }
    const svc = await GetServiceInfo()
    if (svc) {
      serviceInfo.value = svc
    }
  } catch (err) {
    toast.error("Failed to load settings: " + err)
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  saving.value = true
  try {
    await SaveServerConfig({
      api_port: parseInt(config.value.api_port),
      php_workers: parseInt(config.value.php_workers),
      data_dir: config.value.data_dir,
      api_key: config.value.api_key
    })
    toast.success("Settings saved successfully! A restart of Core Service is required for port/worker changes to take effect.")
  } catch (err) {
    toast.error("Failed to save settings: " + err)
  } finally {
    saving.value = false
  }
}

const openDataDir = async () => {
  if (!config.value.data_dir) return
  try {
    await OpenPathInExplorer(config.value.data_dir)
  } catch (err) {
    toast.error("Failed to open data directory: " + err)
  }
}

const copyApiKey = () => {
  navigator.clipboard.writeText(config.value.api_key)
  toast.success("API Secret Key copied to clipboard!")
}

onMounted(() => {
  fetchConfig()
})
</script>

<template>
  <div class="flex flex-col gap-6 animate-fade-in max-w-4xl mx-auto">
    <!-- Header -->
    <div class="pb-2 border-b border-border-subtle/80">
      <h1 class="text-2xl font-bold text-white tracking-tight">Global Engine Settings</h1>
      <p class="text-xs text-text-muted mt-1">Configure internal daemon ports, FastCGI worker concurrency, and security tokens</p>
    </div>

    <div v-if="loading" class="card p-12 text-center text-text-muted text-sm">
      <svg class="animate-spin w-6 h-6 mx-auto mb-2 text-primary" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
      Loading settings...
    </div>

    <form v-else @submit.prevent="saveConfig" class="flex flex-col gap-6">
      <!-- Section 1: Network & Concurrency -->
      <div class="card p-6 rounded-2xl border-border-subtle flex flex-col gap-5">
        <div>
          <h2 class="text-base font-bold text-white flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-primary"><rect x="2" y="2" width="20" height="8" rx="2" ry="2"/><rect x="2" y="14" width="20" height="8" rx="2" ry="2"/><line x1="6" y1="6" x2="6.01" y2="6"/><line x1="6" y1="18" x2="6.01" y2="18"/></svg>
            Network & Worker Concurrency
          </h2>
          <p class="text-xs text-text-muted mt-0.5">Control internal REST API communication and PHP FastCGI capacity</p>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
          <!-- API Port -->
          <div>
            <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">
              Internal REST API Port *
            </label>
            <input 
              v-model="config.api_port" 
              type="number" 
              class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main font-mono focus:outline-none focus:border-primary transition-all"
              required
            >
            <p class="text-[11px] text-text-muted mt-1">Default <code>8081</code>. Used by this GUI dashboard to communicate with the Core Service.</p>
          </div>

          <!-- PHP Workers -->
          <div>
            <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">
              Worker Pool Size (Concurrency) *
            </label>
            <input 
              v-model="config.php_workers" 
              type="number" 
              min="2"
              max="128"
              class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main font-mono focus:outline-none focus:border-primary transition-all"
              required
            >
            <p class="text-[11px] text-text-muted mt-1">Number of dedicated <code>php-cgi.exe</code> processes per PHP version (default 16 or 32).</p>
          </div>
        </div>

        <!-- API Secret Key -->
        <div>
          <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">
            API Secret Key (Authentication) *
          </label>
          <div class="flex gap-2">
            <div class="relative flex-grow">
              <input 
                v-model="config.api_key" 
                :type="showApiKey ? 'text' : 'password'" 
                class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main font-mono focus:outline-none focus:border-primary transition-all"
                required
              >
              <button 
                type="button" 
                @click="showApiKey = !showApiKey" 
                class="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-white text-xs"
              >
                {{ showApiKey ? 'Hide' : 'Show' }}
              </button>
            </div>
            <button 
              type="button" 
              @click="copyApiKey" 
              class="bg-white/5 hover:bg-white/10 text-white text-xs font-semibold px-3 py-2 rounded-xl border border-border-subtle transition-colors shrink-0"
            >
              Copy Key
            </button>
          </div>
          <p class="text-[11px] text-text-muted mt-1">Sent in <code>X-API-Key</code> header to authorize reload and telemetry requests.</p>
        </div>
      </div>

      <!-- Section 2: Storage & Filesystem -->
      <div class="card p-6 rounded-2xl border-border-subtle flex flex-col gap-4">
        <div>
          <h2 class="text-base font-bold text-white flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-amber-400"><path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"/></svg>
            System Distribution Path
          </h2>
          <p class="text-xs text-text-muted mt-0.5">Base location where configurations, logs, and PHP runtimes reside</p>
        </div>

        <div class="flex items-center gap-2">
          <input 
            v-model="config.data_dir" 
            type="text" 
            class="w-full bg-[#111923]/60 border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-muted font-mono cursor-not-allowed select-all"
            readonly
          >
          <button 
            type="button" 
            @click="openDataDir" 
            class="bg-white/10 hover:bg-white/20 text-white text-xs font-semibold px-3.5 py-2 rounded-xl transition-colors shrink-0 flex items-center gap-1.5"
            title="Open folder in Windows Explorer"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
            <span>Open in Explorer</span>
          </button>
        </div>
      </div>

      <!-- Section 3: Windows Service Status & Guide -->
      <div class="card p-6 rounded-2xl border-border-subtle flex flex-col gap-4">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-base font-bold text-white flex items-center gap-2">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-emerald-400"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="m9 12 2 2 4-4"/></svg>
              Windows Service Status
            </h2>
            <p class="text-xs text-text-muted mt-0.5">Windows Service Manager registration (Service: <code>PHForgeCoreSvc</code>)</p>
          </div>
          <span 
            class="text-xs px-2.5 py-1 rounded-full font-semibold"
            :class="serviceInfo.running ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : serviceInfo.installed ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20' : 'bg-white/5 text-text-muted border border-border-subtle'"
          >
            {{ serviceInfo.running ? 'Running as Service' : serviceInfo.installed ? 'Installed (Stopped)' : 'Not Installed' }}
          </span>
        </div>

        <div class="bg-[#111923] p-4 rounded-xl border border-border-subtle text-xs flex flex-col gap-2 font-mono">
          <div class="text-text-muted text-[11px] uppercase tracking-wider font-sans font-bold">Quick Command Reference (Run as Admin):</div>
          <div class="flex items-center justify-between bg-black/30 p-2 rounded-lg">
            <span>phforge-core.exe --install</span>
            <span class="text-text-muted text-[10px] font-sans">Install service</span>
          </div>
          <div class="flex items-center justify-between bg-black/30 p-2 rounded-lg">
            <span>phforge-core.exe --uninstall</span>
            <span class="text-text-muted text-[10px] font-sans">Uninstall service</span>
          </div>
        </div>
      </div>

      <!-- Action Button -->
      <div class="flex items-center justify-end gap-3 pt-2">
        <button 
          type="submit" 
          :disabled="saving"
          class="bg-primary hover:bg-primary-hover disabled:opacity-50 text-white text-xs font-semibold px-6 py-2.5 rounded-xl shadow-lg shadow-primary/20 transition-all cursor-pointer"
        >
          {{ saving ? 'Saving Changes...' : 'Save Global Settings' }}
        </button>
      </div>
    </form>
  </div>
</template>
