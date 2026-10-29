<script setup>
import { ref, onMounted, computed } from 'vue'
import { 
  GetInstalledPHPVersions, 
  GetPHPExtensions, 
  TogglePHPExtension,
  OpenPhpIni,
  OpenPathInExplorer,
  RestartCoreService,
  GetPHPTuning,
  SavePHPTuning
} from '../../wailsjs/go/main/App'
import { useToast } from '../composables/useToast'

const toast = useToast()

const versions = ref([])
const loading = ref(true)

// Extension Manager State
const showModal = ref(false)
const activeVersion = ref('')
const extensions = ref([])
const loadingExts = ref(false)
const searchQuery = ref('')
const extFilter = ref('ALL') // ALL, ENABLED, DISABLED
const extError = ref('')
const hasChanges = ref(false)

// Tuning & OPcache State
const showTuningModal = ref(false)
const tuningVersion = ref('')
const loadingTuning = ref(false)
const savingTuning = ref(false)
const tuning = ref({
  opcache_enabled: false,
  memory_limit: '256M',
  upload_max_filesize: '20M',
  post_max_size: '25M',
  max_execution_time: '60',
  opcache_memory: '128'
})

const fetchVersions = async () => {
  loading.value = true
  try {
    const res = await GetInstalledPHPVersions()
    versions.value = res || []
  } catch (err) {
    toast.error("Failed to scan PHP versions: " + err)
  } finally {
    loading.value = false
  }
}

// Open Extensions Modal
const openExtensions = async (version) => {
  activeVersion.value = version
  showModal.value = true
  loadingExts.value = true
  extError.value = ''
  searchQuery.value = ''
  extFilter.value = 'ALL'
  hasChanges.value = false
  
  try {
    const res = await GetPHPExtensions(version)
    extensions.value = (res || []).sort((a, b) => a.name.localeCompare(b.name))
  } catch (err) {
    extError.value = String(err)
  } finally {
    loadingExts.value = false
  }
}

// Open Tuning & OPcache Modal
const openTuning = async (version) => {
  tuningVersion.value = version
  showTuningModal.value = true
  loadingTuning.value = true
  try {
    const res = await GetPHPTuning(version)
    if (res) {
      tuning.value = {
        opcache_enabled: res.opcache_enabled ?? false,
        memory_limit: res.memory_limit || '256M',
        upload_max_filesize: res.upload_max_filesize || '20M',
        post_max_size: res.post_max_size || '25M',
        max_execution_time: res.max_execution_time || '60',
        opcache_memory: res.opcache_memory || '128'
      }
    }
  } catch (err) {
    toast.error("Gagal membaca konfigurasi tuning: " + err)
  } finally {
    loadingTuning.value = false
  }
}

// Apply Preset Suggestions
const applyPreset = (type) => {
  if (type === 'laravel') {
    tuning.value = {
      opcache_enabled: true,
      memory_limit: '256M',
      upload_max_filesize: '20M',
      post_max_size: '25M',
      max_execution_time: '60',
      opcache_memory: '128'
    }
    toast.info("Preset Rekomendasi Laravel diterapkan!")
  } else if (type === 'enterprise') {
    tuning.value = {
      opcache_enabled: true,
      memory_limit: '512M',
      upload_max_filesize: '50M',
      post_max_size: '60M',
      max_execution_time: '120',
      opcache_memory: '256'
    }
    toast.info("Preset Heavy Production diterapkan!")
  } else if (type === 'default') {
    tuning.value = {
      opcache_enabled: false,
      memory_limit: '128M',
      upload_max_filesize: '2M',
      post_max_size: '8M',
      max_execution_time: '30',
      opcache_memory: '128'
    }
    toast.info("Preset Default PHP diterapkan!")
  }
}

// Save Tuning
const saveTuningConfig = async () => {
  savingTuning.value = true
  try {
    await SavePHPTuning(tuningVersion.value, tuning.value)
    toast.success(`Konfigurasi tuning PHP ${tuningVersion.value} berhasil disimpan!`)
    hasChanges.value = true
    showTuningModal.value = false
  } catch (err) {
    toast.error("Gagal menyimpan konfigurasi tuning: " + err)
  } finally {
    savingTuning.value = false
  }
}

const handleOpenIni = async (version) => {
  try {
    await OpenPhpIni(version)
    toast.info(`Opened php.ini for PHP ${version} in Notepad.`)
  } catch (err) {
    toast.error("Failed to open php.ini: " + err)
  }
}

const handleOpenFolder = async (path) => {
  try {
    await OpenPathInExplorer(path)
  } catch (err) {
    toast.error("Failed to open directory: " + err)
  }
}

const toggleExt = async (extName, currentStatus) => {
  const newStatus = !currentStatus
  try {
    await TogglePHPExtension(activeVersion.value, extName, newStatus)
    const ext = extensions.value.find(e => e.name === extName)
    if (ext) ext.enabled = newStatus
    hasChanges.value = true
    toast.success(`Extension "${extName}" ${newStatus ? 'enabled' : 'disabled'}`)
  } catch (err) {
    toast.error("Failed to toggle extension: " + err)
  }
}

const reloadServiceAfterExt = async () => {
  try {
    await RestartCoreService()
    hasChanges.value = false
    toast.success("Core Service reloaded! Perubahan kini aktif di semua worker.")
  } catch (err) {
    toast.error("Failed to reload service: " + err)
  }
}

const filteredExtensions = computed(() => {
  let list = extensions.value.filter(e => !['opcache', 'php_opcache'].includes(e.name.toLowerCase()))

  if (extFilter.value === 'ENABLED') {
    list = list.filter(e => e.enabled)
  } else if (extFilter.value === 'DISABLED') {
    list = list.filter(e => !e.enabled)
  }

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim()
    list = list.filter(e => e.name.toLowerCase().includes(q))
  }

  return list
})

const enabledCount = computed(() => extensions.value.filter(e => !['opcache', 'php_opcache'].includes(e.name.toLowerCase()) && e.enabled).length)

onMounted(() => {
  fetchVersions()
})
</script>

<template>
  <div class="flex flex-col gap-6 animate-fade-in max-w-7xl mx-auto">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 pb-2 border-b border-border-subtle/80">
      <div>
        <h1 class="text-2xl font-bold text-white tracking-tight flex items-center gap-2.5">
          PHP Runtime Roster
          <span class="text-xs px-2.5 py-0.5 rounded-full font-mono font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/20">
            {{ versions.length }} Versions Found
          </span>
        </h1>
        <p class="text-xs text-text-muted mt-1">Configured PHP binary environments residing in <code>dist/php-versions/</code></p>
      </div>

      <button 
        @click="fetchVersions" 
        class="flex items-center gap-2 bg-white/5 hover:bg-white/10 text-white text-xs font-semibold py-2 px-4 rounded-xl border border-border-subtle transition-all cursor-pointer"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/><path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16"/><path d="M16 21h5v-5"/></svg>
        <span>Rescan Directory</span>
      </button>
    </div>

    <!-- Notification if changed -->
    <div v-if="hasChanges" class="bg-amber-500/10 border border-amber-500/25 rounded-2xl p-4 flex items-center justify-between">
      <span class="text-xs text-amber-300 font-medium flex items-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        Konfigurasi PHP telah diperbarui! Muat ulang (Reload) Core Service untuk menerapkan ke worker aktif.
      </span>
      <button 
        @click="reloadServiceAfterExt"
        class="bg-amber-500 hover:bg-amber-400 text-black text-xs font-bold px-4 py-1.5 rounded-xl shadow transition-all cursor-pointer shrink-0"
      >
        Reload Service Sekarang
      </button>
    </div>

    <!-- Versions Table -->
    <div class="card p-0 rounded-2xl overflow-hidden border-border-subtle shadow-xl">
      <div v-if="loading" class="text-text-muted text-center py-12 text-sm">
        <svg class="animate-spin w-6 h-6 mx-auto mb-2 text-primary" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
        Scanning for PHP distributions...
      </div>

      <div v-else-if="versions.length === 0" class="text-center py-16 px-4">
        <div class="w-14 h-14 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center mx-auto mb-3 text-text-muted">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m18 16 4-4-4-4"/><path d="m6 8-4 4 4 4"/><path d="m14.5 4-5 16"/></svg>
        </div>
        <h3 class="text-base font-bold text-white mb-1">No PHP Versions Detected</h3>
        <p class="text-xs text-text-muted max-w-md mx-auto">
          Extract standard Windows NTS/TS PHP binaries into folders like <code>dist/php-versions/php-8.4</code> or <code>dist/php-versions/php-7.4</code>.
        </p>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-[#101824] border-b border-border-subtle text-text-muted text-[11px] font-bold uppercase tracking-wider">
              <th class="py-3.5 px-6">PHP Version</th>
              <th class="py-3.5 px-6">Binary Location</th>
              <th class="py-3.5 px-6">Engine Readiness</th>
              <th class="py-3.5 px-6 text-right">Quick Tools</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border-subtle/50">
            <tr v-for="v in versions" :key="v.version" class="hover:bg-white/[0.02] transition-colors group">
              <!-- Version -->
              <td class="py-4 px-6">
                <div class="flex items-center gap-2.5">
                  <div class="w-9 h-9 rounded-xl bg-gradient-to-tr from-primary to-blue-500 text-white font-extrabold text-xs flex items-center justify-center shadow-md">
                    PHP
                  </div>
                  <div>
                    <div class="font-bold text-white text-sm font-mono">{{ v.version }}</div>
                    <div class="text-[10px] text-text-muted font-medium">Windows x64 FastCGI</div>
                  </div>
                </div>
              </td>

              <!-- Path -->
              <td class="py-4 px-6">
                <span class="text-xs text-text-muted font-mono truncate block max-w-xs sm:max-w-md" :title="v.path">
                  {{ v.path }}
                </span>
              </td>

              <!-- Status -->
              <td class="py-4 px-6">
                <span 
                  v-if="v.status === 'Installed'" 
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
                  Ready (CGI & CLI)
                </span>
                <span 
                  v-else 
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-semibold bg-red-500/10 text-red-400 border border-red-500/20"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-red-400"></span>
                  {{ v.status }}
                </span>
              </td>

              <!-- Action buttons -->
              <td class="py-4 px-6 text-right">
                <div class="flex items-center justify-end gap-2">
                  <!-- Tuning & OPcache Button -->
                  <button 
                    @click="openTuning(v.version)" 
                    class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold bg-emerald-500/10 text-emerald-300 hover:bg-emerald-500 hover:text-black border border-emerald-500/20 transition-all cursor-pointer shadow-sm"
                    title="Optimasi OPcache, memory limit, dan upload limit"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
                    <span>Tuning & OPcache</span>
                  </button>

                  <!-- Extensions Manager -->
                  <button 
                    @click="openExtensions(v.version)" 
                    class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold bg-primary/10 text-primary hover:bg-primary hover:text-white transition-all cursor-pointer"
                    title="Toggle php.ini extensions"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20.94c1.5 0 2.75 1.06 4 1.06 3 0 6-8 6-12.22A4.91 4.91 0 0 0 17 5c-2.22 0-4 1.44-5 2-1-.56-2.78-2-5-2a4.9 4.9 0 0 0-5 4.78C2 14 5 22 8 22c1.25 0 2.5-1.06 4-1.06Z"/><polygon points="10 2 10 10 14 10 14 2"/></svg>
                    <span>Extensions</span>
                  </button>

                  <!-- Edit php.ini in Notepad -->
                  <button 
                    @click="handleOpenIni(v.version)" 
                    class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold bg-white/5 hover:bg-white/10 text-text-main border border-border-subtle transition-all cursor-pointer"
                    title="Open php.ini in Notepad"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path></svg>
                    <span>php.ini</span>
                  </button>

                  <!-- Open Folder -->
                  <button 
                    @click="handleOpenFolder(v.path)" 
                    class="p-2 rounded-xl text-text-muted hover:text-amber-400 hover:bg-amber-500/10 transition-colors" 
                    title="Open PHP directory in Windows Explorer"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"/></svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Tuning & OPcache Modal -->
    <div 
      v-if="showTuningModal" 
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in"
      @click.self="showTuningModal = false"
    >
      <div class="bg-dark-card border border-border-subtle rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] flex flex-col overflow-hidden">
        <!-- Header -->
        <div class="p-6 border-b border-border-subtle flex justify-between items-center bg-[#101824]">
          <div>
            <h2 class="text-lg font-bold text-white flex items-center gap-2">
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-emerald-400"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
              PHP {{ tuningVersion }} Tuning & Optimization
            </h2>
            <p class="text-xs text-text-muted mt-0.5">Atur Zend OPcache, memory limit, dan batas upload file</p>
          </div>
          <button @click="showTuningModal = false" class="text-text-muted hover:text-white p-1 rounded-lg hover:bg-white/5">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
          </button>
        </div>

        <div v-if="loadingTuning" class="p-12 text-center text-text-muted text-xs">
          Membaca konfigurasi tuning dari php.ini...
        </div>

        <div v-else class="p-6 overflow-y-auto flex-grow flex flex-col gap-6 custom-scrollbar">
          <!-- Quick Presets -->
          <div>
            <div class="text-[11px] font-bold uppercase tracking-wider text-text-muted mb-2.5">
              Pilih Rekomendasi Siap Pakai (Preset):
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
              <!-- Preset 1: Laravel -->
              <button 
                type="button" 
                @click="applyPreset('laravel')"
                class="p-3 rounded-xl border border-emerald-500/30 bg-emerald-500/5 hover:bg-emerald-500/10 text-left transition-all cursor-pointer group"
              >
                <div class="text-xs font-bold text-emerald-400 flex items-center justify-between">
                  <span>🚀 Rekomendasi Laravel</span>
                </div>
                <div class="text-[11px] text-text-muted mt-1 leading-snug">
                  OPcache ON (128M)<br>
                  RAM 256M • Upload 20M
                </div>
              </button>

              <!-- Preset 2: Enterprise -->
              <button 
                type="button" 
                @click="applyPreset('enterprise')"
                class="p-3 rounded-xl border border-blue-500/30 bg-blue-500/5 hover:bg-blue-500/10 text-left transition-all cursor-pointer group"
              >
                <div class="text-xs font-bold text-blue-400 flex items-center justify-between">
                  <span>🏢 Heavy / Enterprise</span>
                </div>
                <div class="text-[11px] text-text-muted mt-1 leading-snug">
                  OPcache ON (256M)<br>
                  RAM 512M • Upload 50M
                </div>
              </button>

              <!-- Preset 3: Default -->
              <button 
                type="button" 
                @click="applyPreset('default')"
                class="p-3 rounded-xl border border-border-subtle bg-white/5 hover:bg-white/10 text-left transition-all cursor-pointer group"
              >
                <div class="text-xs font-bold text-text-muted flex items-center justify-between">
                  <span>🔄 Default PHP</span>
                </div>
                <div class="text-[11px] text-text-muted mt-1 leading-snug">
                  OPcache OFF<br>
                  RAM 128M • Upload 2M
                </div>
              </button>
            </div>
          </div>

          <!-- Section 1: Zend OPcache -->
          <div class="p-4 rounded-xl border border-border-subtle bg-[#121c29]/60 flex flex-col gap-3">
            <div class="flex items-center justify-between">
              <div>
                <div class="text-sm font-bold text-white flex items-center gap-2">
                  <span>Zend OPcache Engine</span>
                  <span 
                    class="text-[10px] font-mono px-2 py-0.5 rounded-full font-bold uppercase"
                    :class="tuning.opcache_enabled ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30' : 'bg-slate-700/50 text-text-muted'"
                  >
                    {{ tuning.opcache_enabled ? 'Aktif' : 'Nonaktif' }}
                  </span>
                </div>
                <p class="text-xs text-text-muted mt-1 leading-relaxed">
                  Menyimpan bytecode terkompilasi Laravel di RAM. Mencegah PHP membaca ulang ribuan file <code>vendor/</code> dari disk pada setiap request (Boost 3x–5x).
                </p>
              </div>

              <!-- Switch -->
              <button 
                type="button"
                @click="tuning.opcache_enabled = !tuning.opcache_enabled"
                class="relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                :class="tuning.opcache_enabled ? 'bg-emerald-500' : 'bg-slate-700'"
              >
                <span 
                  class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                  :class="tuning.opcache_enabled ? 'translate-x-5' : 'translate-x-0'"
                ></span>
              </button>
            </div>

            <!-- OPcache RAM allocation -->
            <div v-if="tuning.opcache_enabled" class="pt-3 border-t border-border-subtle/50 flex flex-col gap-1.5">
              <label class="text-xs font-semibold text-text-muted">Alokasi RAM OPcache (opcache.memory_consumption)</label>
              <div class="flex items-center gap-2">
                <input 
                  v-model="tuning.opcache_memory" 
                  type="text" 
                  class="w-32 bg-[#111923] border border-border-subtle rounded-xl px-3 py-1.5 text-xs text-text-main font-mono focus:outline-none focus:border-primary"
                >
                <div class="flex items-center gap-1.5">
                  <button 
                    v-for="preset in ['64', '128', '256', '512']" 
                    :key="preset"
                    type="button"
                    @click="tuning.opcache_memory = preset"
                    class="text-[11px] font-mono px-2 py-1 rounded-lg border transition-colors cursor-pointer"
                    :class="tuning.opcache_memory === preset ? 'bg-emerald-500/20 border-emerald-500/40 text-emerald-300' : 'bg-white/5 border-border-subtle text-text-muted hover:text-white'"
                  >
                    {{ preset }}MB
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Section 2: Memory & Execution Limits -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <!-- Memory Limit -->
            <div>
              <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">
                memory_limit *
              </label>
              <input 
                v-model="tuning.memory_limit" 
                type="text" 
                class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main font-mono focus:outline-none focus:border-primary transition-all"
                required
              >
              <div class="flex items-center gap-1.5 mt-1.5">
                <span class="text-[10px] text-text-muted">Saran:</span>
                <button 
                  v-for="p in ['128M', '256M', '512M', '1024M']" 
                  :key="p"
                  type="button"
                  @click="tuning.memory_limit = p"
                  class="text-[10px] font-mono px-1.5 py-0.5 rounded border transition-colors"
                  :class="tuning.memory_limit === p ? 'bg-primary/20 border-primary text-primary' : 'bg-white/5 border-border-subtle text-text-muted hover:text-white'"
                >
                  {{ p }}
                </button>
              </div>
            </div>

            <!-- Max Execution Time -->
            <div>
              <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">
                max_execution_time (detik) *
              </label>
              <input 
                v-model="tuning.max_execution_time" 
                type="number" 
                class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main font-mono focus:outline-none focus:border-primary transition-all"
                required
              >
              <div class="flex items-center gap-1.5 mt-1.5">
                <span class="text-[10px] text-text-muted">Saran:</span>
                <button 
                  v-for="p in ['30', '60', '120', '300']" 
                  :key="p"
                  type="button"
                  @click="tuning.max_execution_time = p"
                  class="text-[10px] font-mono px-1.5 py-0.5 rounded border transition-colors"
                  :class="tuning.max_execution_time === p ? 'bg-primary/20 border-primary text-primary' : 'bg-white/5 border-border-subtle text-text-muted hover:text-white'"
                >
                  {{ p }}s
                </button>
              </div>
            </div>
          </div>

          <!-- Section 3: Upload Limits -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <!-- Upload Max Filesize -->
            <div>
              <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">
                upload_max_filesize *
              </label>
              <input 
                v-model="tuning.upload_max_filesize" 
                type="text" 
                class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main font-mono focus:outline-none focus:border-primary transition-all"
                required
              >
              <div class="flex items-center gap-1.5 mt-1.5">
                <span class="text-[10px] text-text-muted">Saran:</span>
                <button 
                  v-for="p in ['2M', '20M', '50M', '100M']" 
                  :key="p"
                  type="button"
                  @click="tuning.upload_max_filesize = p"
                  class="text-[10px] font-mono px-1.5 py-0.5 rounded border transition-colors"
                  :class="tuning.upload_max_filesize === p ? 'bg-primary/20 border-primary text-primary' : 'bg-white/5 border-border-subtle text-text-muted hover:text-white'"
                >
                  {{ p }}
                </button>
              </div>
            </div>

            <!-- Post Max Size -->
            <div>
              <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">
                post_max_size *
              </label>
              <input 
                v-model="tuning.post_max_size" 
                type="text" 
                class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main font-mono focus:outline-none focus:border-primary transition-all"
                required
              >
              <div class="flex items-center gap-1.5 mt-1.5">
                <span class="text-[10px] text-text-muted">Saran:</span>
                <button 
                  v-for="p in ['8M', '25M', '60M', '120M']" 
                  :key="p"
                  type="button"
                  @click="tuning.post_max_size = p"
                  class="text-[10px] font-mono px-1.5 py-0.5 rounded border transition-colors"
                  :class="tuning.post_max_size === p ? 'bg-primary/20 border-primary text-primary' : 'bg-white/5 border-border-subtle text-text-muted hover:text-white'"
                >
                  {{ p }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-4 border-t border-border-subtle bg-[#101824] flex items-center justify-between">
          <button 
            type="button"
            @click="handleOpenIni(tuningVersion)"
            class="text-xs text-text-muted hover:text-white flex items-center gap-1.5 transition-colors"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path></svg>
            <span>Buka Notepad</span>
          </button>
          
          <div class="flex items-center gap-2.5">
            <button 
              type="button"
              @click="showTuningModal = false" 
              class="bg-white/5 hover:bg-white/10 text-text-muted hover:text-white text-xs font-semibold px-4 py-2 rounded-xl transition-colors cursor-pointer"
            >
              Batal
            </button>
            <button 
              type="button"
              @click="saveTuningConfig"
              :disabled="savingTuning"
              class="bg-emerald-500 hover:bg-emerald-400 text-black text-xs font-bold px-5 py-2 rounded-xl shadow-lg shadow-emerald-950/40 transition-all cursor-pointer disabled:opacity-50"
            >
              {{ savingTuning ? 'Menyimpan...' : 'Simpan Pengaturan Tuning' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Extensions Manager Modal -->
    <div 
      v-if="showModal" 
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in"
      @click.self="showModal = false"
    >
      <div class="bg-dark-card border border-border-subtle rounded-2xl shadow-2xl w-full max-w-2xl max-h-[85vh] flex flex-col overflow-hidden">
        <!-- Header -->
        <div class="p-6 border-b border-border-subtle flex justify-between items-center bg-[#101824]">
          <div>
            <h2 class="text-lg font-bold text-white flex items-center gap-2">
              PHP {{ activeVersion }} Extensions
              <span class="text-xs font-mono px-2 py-0.5 rounded-full bg-emerald-500/15 text-emerald-400 border border-emerald-500/20">
                {{ enabledCount }} of {{ extensions.length }} Enabled
              </span>
            </h2>
            <p class="text-xs text-text-muted mt-0.5">Toggle modules instantly in <code>php.ini</code></p>
          </div>
          <button @click="showModal = false" class="text-text-muted hover:text-white p-1 rounded-lg hover:bg-white/5">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
          </button>
        </div>

        <!-- OPcache Info Notice -->
        <div class="px-6 py-2.5 bg-emerald-500/10 border-b border-border-subtle flex items-center justify-between text-xs">
          <div class="flex items-center gap-2 text-emerald-300">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-emerald-400 shrink-0"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
            <span>Zend OPcache kini dikelola secara terpusat pada menu <strong>Tuning & OPcache</strong>.</span>
          </div>
          <button 
            type="button"
            @click="showModal = false; openTuning(activeVersion)" 
            class="text-emerald-400 hover:text-emerald-200 font-semibold underline cursor-pointer shrink-0 ml-2"
          >
            Buka Menu Tuning &rarr;
          </button>
        </div>

        <!-- Search & Filter Controls -->
        <div class="p-4 border-b border-border-subtle bg-[#121c29] flex flex-col sm:flex-row gap-3 items-center justify-between">
          <!-- Search -->
          <div class="relative w-full sm:w-64">
            <svg class="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted w-3.5 h-3.5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            <input 
              v-model="searchQuery" 
              type="text" 
              placeholder="Search extensions (curl, gd, mbstring)..." 
              class="w-full bg-[#111923] border border-border-subtle rounded-xl pl-9 pr-3 py-1.5 text-xs text-text-main focus:outline-none focus:border-primary transition-all placeholder:text-text-muted/60"
            >
          </div>

          <!-- Filter Pills -->
          <div class="flex rounded-lg bg-[#111923] p-1 border border-border-subtle text-xs w-full sm:w-auto justify-center">
            <button 
              v-for="filter in ['ALL', 'ENABLED', 'DISABLED']"
              :key="filter"
              @click="extFilter = filter"
              class="px-3 py-1 rounded-md font-semibold text-[11px] transition-colors"
              :class="extFilter === filter ? 'bg-primary text-white shadow-sm' : 'text-text-muted hover:text-white'"
            >
              {{ filter }}
            </button>
          </div>
        </div>

        <!-- Extensions Grid -->
        <div class="p-6 overflow-y-auto flex-grow bg-dark-card custom-scrollbar">
          <div v-if="loadingExts" class="text-center py-12 text-text-muted text-xs">
            Reading php.ini configuration...
          </div>
          <div v-else-if="extError" class="text-center py-12 text-danger text-xs">
            {{ extError }}
          </div>
          <div v-else-if="filteredExtensions.length === 0" class="text-center py-12 text-text-muted text-xs">
            No extensions match your filter.
          </div>

          <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div 
              v-for="ext in filteredExtensions" 
              :key="ext.name"
              class="flex items-center justify-between p-3 rounded-xl border transition-all"
              :class="ext.enabled ? 'bg-emerald-500/[0.04] border-emerald-500/25 hover:border-emerald-500/40' : 'bg-white/[0.02] border-border-subtle/50 hover:border-border-subtle'"
            >
              <div class="flex items-center gap-2.5">
                <span 
                  class="w-2 h-2 rounded-full"
                  :class="ext.enabled ? 'bg-emerald-400' : 'bg-slate-600'"
                ></span>
                <span 
                  class="text-xs font-mono font-medium"
                  :class="ext.enabled ? 'text-white font-bold' : 'text-text-muted'"
                >
                  {{ ext.name }}
                </span>
              </div>

              <!-- Switch -->
              <button 
                @click="toggleExt(ext.name, ext.enabled)"
                class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                :class="ext.enabled ? 'bg-emerald-500' : 'bg-slate-700'"
              >
                <span 
                  class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                  :class="ext.enabled ? 'translate-x-4' : 'translate-x-0'"
                ></span>
              </button>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-4 border-t border-border-subtle bg-[#101824] flex items-center justify-between">
          <button 
            @click="handleOpenIni(activeVersion)"
            class="text-xs text-text-muted hover:text-white flex items-center gap-1.5 transition-colors"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path></svg>
            <span>Open in Notepad for raw edits</span>
          </button>
          
          <button 
            @click="showModal = false" 
            class="bg-white/10 hover:bg-white/20 text-white text-xs font-semibold px-5 py-2 rounded-xl transition-colors cursor-pointer"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.02);
  border-radius: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.2);
}
</style>
