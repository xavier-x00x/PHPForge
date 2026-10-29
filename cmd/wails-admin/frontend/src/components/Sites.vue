<script setup>
import { ref, onMounted, computed } from 'vue'
import { 
  GetSitesConfig, 
  SaveSitesConfig, 
  GetInstalledPHPVersions, 
  OpenSiteCLI, 
  SelectDirectory, 
  OpenBrowser, 
  OpenPathInExplorer,
  CheckPortAvailable
} from '../../wailsjs/go/main/App'
import { useToast } from '../composables/useToast'
import ConfirmModal from './ConfirmModal.vue'

const toast = useToast()

const sites = ref([])
const phpVersions = ref([])
const loading = ref(true)
const searchQuery = ref('')

const showAddModal = ref(false)
const editingIndex = ref(null)

const confirmDeleteOpen = ref(false)
const siteToDeleteIndex = ref(null)

const newSite = ref({
  name: '',
  domain: 'localhost',
  port: 8080,
  document_root: '',
  php_version: ''
})

const editSite = ref({
  name: '',
  domain: 'localhost',
  port: 8080,
  document_root: '',
  php_version: ''
})

const fetchSites = async () => {
  loading.value = true
  try {
    const res = await GetSitesConfig()
    sites.value = res.sites || []
    const phpRes = await GetInstalledPHPVersions()
    phpVersions.value = phpRes || []

    if (!newSite.value.php_version && phpVersions.value.length > 0) {
      newSite.value.php_version = phpVersions.value[0].version
    }
  } catch (err) {
    toast.error("Failed to load sites: " + err)
  } finally {
    loading.value = false
  }
}

// Browse Directory Dialog
const browseDirectory = async (target) => {
  try {
    const path = await SelectDirectory("Select Document Root (e.g. public folder)")
    if (path) {
      if (target === 'new') {
        newSite.value.document_root = path
      } else {
        editSite.value.document_root = path
      }
    }
  } catch (err) {
    toast.error("Failed to browse directory: " + err)
  }
}

// Open CLI
const handleOpenCLI = async (site) => {
  try {
    await OpenSiteCLI(site.document_root, site.php_version)
    toast.info(`Terminal opened for ${site.name} (PHP ${site.php_version})`)
  } catch (err) {
    toast.error("Failed to open terminal: " + err)
  }
}

// Open in Default Browser
const handleOpenBrowser = async (site) => {
  const url = `http://${site.domain || 'localhost'}:${site.port || 80}`
  try {
    await OpenBrowser(url)
  } catch (err) {
    toast.error("Failed to open browser: " + err)
  }
}

// Open Folder in Windows Explorer
const handleOpenExplorer = async (path) => {
  try {
    await OpenPathInExplorer(path)
  } catch (err) {
    toast.error("Failed to open folder: " + err)
  }
}

// Copy to Clipboard
const copyToClipboard = (text, label) => {
  navigator.clipboard.writeText(text)
  toast.success(`Copied ${label} to clipboard!`)
}

// Add Site
const addSite = async () => {
  if (!newSite.value.name.trim() || !newSite.value.domain.trim() || !newSite.value.document_root.trim()) {
    toast.error("Please fill in all required fields.")
    return
  }

  const port = parseInt(newSite.value.port)
  if (isNaN(port) || port < 1 || port > 65535) {
    toast.error("Port must be between 1 and 65535.")
    return
  }

  // Duplicate Check
  const isDuplicate = sites.value.some(s => s.domain === newSite.value.domain && s.port === port)
  if (isDuplicate) {
    toast.error(`Domain ${newSite.value.domain} on port ${port} already exists!`)
    return
  }

  const updatedSites = [...sites.value, {
    name: newSite.value.name.trim(),
    domain: newSite.value.domain.trim(),
    port: port,
    document_root: newSite.value.document_root.trim(),
    php_version: newSite.value.php_version
  }]

  try {
    await SaveSitesConfig({ sites: updatedSites })
    sites.value = updatedSites
    showAddModal.value = false
    toast.success(`Site "${newSite.value.name}" added successfully! Remember to Reload Service.`)
    newSite.value = {
      name: '',
      domain: 'localhost',
      port: 8080,
      document_root: '',
      php_version: phpVersions.value.length > 0 ? phpVersions.value[0].version : ''
    }
  } catch (err) {
    toast.error("Failed to save site: " + err)
  }
}

// Edit Site
const startEdit = (idx) => {
  editingIndex.value = idx
  editSite.value = { ...sites.value[idx] }
}

const saveEdit = async () => {
  if (editingIndex.value === null) return

  const port = parseInt(editSite.value.port)
  if (isNaN(port) || port < 1 || port > 65535) {
    toast.error("Port must be between 1 and 65535.")
    return
  }

  // Check duplicates
  const isDuplicate = sites.value.some((s, i) => i !== editingIndex.value && s.domain === editSite.value.domain && s.port === port)
  if (isDuplicate) {
    toast.error(`Domain ${editSite.value.domain} on port ${port} already exists!`)
    return
  }

  const updatedSites = [...sites.value]
  updatedSites[editingIndex.value] = {
    ...editSite.value,
    port: port
  }

  try {
    await SaveSitesConfig({ sites: updatedSites })
    sites.value = updatedSites
    editingIndex.value = null
    toast.success("Site configuration updated successfully!")
  } catch (err) {
    toast.error("Failed to update site: " + err)
  }
}

// Delete Confirmation
const requestDeleteSite = (idx) => {
  siteToDeleteIndex.value = idx
  confirmDeleteOpen.value = true
}

const confirmDelete = async () => {
  if (siteToDeleteIndex.value === null) return
  const deletedSite = sites.value[siteToDeleteIndex.value]
  const updatedSites = sites.value.filter((_, i) => i !== siteToDeleteIndex.value)

  try {
    await SaveSitesConfig({ sites: updatedSites })
    sites.value = updatedSites
    toast.success(`Site "${deletedSite.name}" deleted.`)
  } catch (err) {
    toast.error("Failed to remove site: " + err)
  } finally {
    confirmDeleteOpen.value = false
    siteToDeleteIndex.value = null
  }
}

const filteredSites = computed(() => {
  if (!searchQuery.value.trim()) return sites.value
  const q = searchQuery.value.toLowerCase()
  return sites.value.filter(s => 
    s.name.toLowerCase().includes(q) || 
    s.domain.toLowerCase().includes(q) || 
    String(s.port).includes(q) ||
    s.php_version.toLowerCase().includes(q)
  )
})

onMounted(() => {
  fetchSites()
})
</script>

<template>
  <div class="flex flex-col gap-6 animate-fade-in max-w-7xl mx-auto">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 pb-2 border-b border-border-subtle/80">
      <div>
        <h1 class="text-2xl font-bold text-white tracking-tight flex items-center gap-2.5">
          Sites & Virtual Hosts
          <span class="text-xs px-2.5 py-0.5 rounded-full font-mono font-semibold bg-primary/15 text-primary border border-primary/20">
            {{ sites.length }} Configured
          </span>
        </h1>
        <p class="text-xs text-text-muted mt-1">Manage web routes, isolated PHP runtime versions, and local docroots</p>
      </div>

      <div class="flex items-center gap-3 w-full sm:w-auto">
        <!-- Search bar -->
        <div class="relative flex-grow sm:w-64">
          <svg class="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted w-3.5 h-3.5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          <input 
            v-model="searchQuery" 
            type="text" 
            placeholder="Search sites or ports..." 
            class="w-full bg-[#111923] border border-border-subtle rounded-xl pl-9 pr-3 py-2 text-xs text-text-main focus:outline-none focus:border-primary transition-all placeholder:text-text-muted/60"
          >
        </div>

        <!-- Add Button -->
        <button 
          @click="showAddModal = true" 
          class="flex items-center gap-2 bg-primary hover:bg-primary-hover text-white text-xs font-semibold py-2 px-4 rounded-xl shadow-lg shadow-primary/20 hover:-translate-y-0.5 transition-all shrink-0 cursor-pointer"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          <span>Add Site</span>
        </button>
      </div>
    </div>

    <!-- Sites Table Card -->
    <div class="card p-0 rounded-2xl overflow-hidden border-border-subtle shadow-xl">
      <div v-if="loading" class="text-text-muted text-center py-12 text-sm">
        <svg class="animate-spin w-6 h-6 mx-auto mb-2 text-primary" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
        Loading registered sites...
      </div>
      
      <div v-else-if="filteredSites.length === 0" class="text-center py-16 px-4">
        <div class="w-14 h-14 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center mx-auto mb-3 text-text-muted">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" x2="22" y1="12" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
        </div>
        <h3 class="text-base font-bold text-white mb-1">No Sites Found</h3>
        <p class="text-xs text-text-muted max-w-sm mx-auto mb-4">
          {{ searchQuery ? 'No sites match your filter query. Try clearing the search.' : 'You haven\'t configured any sites yet. Click "Add Site" to bind a project folder to a port.' }}
        </p>
        <button 
          v-if="!searchQuery"
          @click="showAddModal = true" 
          class="bg-primary hover:bg-primary-hover text-white text-xs font-semibold py-2 px-5 rounded-xl shadow-lg transition-all"
        >
          + Add Your First Site
        </button>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-[#101824] border-b border-border-subtle text-text-muted text-[11px] font-bold uppercase tracking-wider">
              <th class="py-3.5 px-6">Site Name & URL</th>
              <th class="py-3.5 px-6">Port</th>
              <th class="py-3.5 px-6">PHP Runtime</th>
              <th class="py-3.5 px-6">Document Root</th>
              <th class="py-3.5 px-6 text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border-subtle/50">
            <tr 
              v-for="(site, idx) in filteredSites" 
              :key="idx" 
              class="hover:bg-white/[0.02] transition-colors group"
            >
              <!-- Name & URL Link -->
              <td class="py-4 px-6">
                <div class="font-bold text-white text-sm tracking-tight">{{ site.name }}</div>
                <button 
                  @click="handleOpenBrowser(site)" 
                  class="text-xs font-mono text-primary hover:text-blue-400 hover:underline flex items-center gap-1.5 mt-0.5"
                  title="Click to open site in default browser"
                >
                  <span>http://{{ site.domain || 'localhost' }}:{{ site.port }}</span>
                  <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
                </button>
              </td>

              <!-- Port -->
              <td class="py-4 px-6 font-mono text-xs font-bold text-emerald-400">
                :{{ site.port || 8080 }}
              </td>

              <!-- PHP Version Badge -->
              <td class="py-4 px-6">
                <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-semibold bg-blue-500/10 text-blue-300 border border-blue-500/20 font-mono">
                  PHP {{ site.php_version }}
                </span>
              </td>

              <!-- Document Root -->
              <td class="py-4 px-6">
                <div class="flex items-center gap-2 max-w-xs sm:max-w-md">
                  <span class="text-xs text-text-muted font-mono truncate" :title="site.document_root">
                    {{ site.document_root }}
                  </span>
                  <button 
                    @click="copyToClipboard(site.document_root, 'path')" 
                    class="text-text-muted hover:text-white p-1 rounded hover:bg-white/5 opacity-0 group-hover:opacity-100 transition-opacity shrink-0"
                    title="Copy full path"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>
                  </button>
                </div>
              </td>

              <!-- Action Buttons -->
              <td class="py-4 px-6 text-right">
                <div class="flex items-center justify-end gap-1.5">
                  <!-- Browse in Browser -->
                  <button 
                    @click="handleOpenBrowser(site)" 
                    class="p-2 rounded-lg text-text-muted hover:text-blue-400 hover:bg-blue-500/10 transition-colors" 
                    title="Open in Browser"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" x2="22" y1="12" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
                  </button>

                  <!-- Open in Explorer -->
                  <button 
                    @click="handleOpenExplorer(site.document_root)" 
                    class="p-2 rounded-lg text-text-muted hover:text-amber-400 hover:bg-amber-500/10 transition-colors" 
                    title="Open Folder in Windows Explorer"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"/></svg>
                  </button>

                  <!-- Open CLI -->
                  <button 
                    @click="handleOpenCLI(site)" 
                    class="p-2 rounded-lg text-text-muted hover:text-emerald-400 hover:bg-emerald-500/10 transition-colors" 
                    title="Open Command Prompt (with PHP & Composer in PATH)"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>
                  </button>

                  <!-- Edit Site -->
                  <button 
                    @click="startEdit(idx)" 
                    class="p-2 rounded-lg text-text-muted hover:text-primary hover:bg-primary/10 transition-colors" 
                    title="Edit Site Configuration"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"/></svg>
                  </button>

                  <!-- Delete Site -->
                  <button 
                    @click="requestDeleteSite(idx)" 
                    class="p-2 rounded-lg text-text-muted hover:text-danger hover:bg-red-500/10 transition-colors" 
                    title="Delete Site"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Add Site Modal -->
    <div v-if="showAddModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in" @click.self="showAddModal = false">
      <div class="bg-dark-card border border-border-subtle rounded-2xl shadow-2xl w-full max-w-xl p-6">
        <div class="flex justify-between items-center mb-5 pb-3 border-b border-border-subtle">
          <div>
            <h2 class="text-lg font-bold text-white">Add New Virtual Host</h2>
            <p class="text-xs text-text-muted mt-0.5">Configure route, port, and PHP runtime for a project</p>
          </div>
          <button @click="showAddModal = false" class="text-text-muted hover:text-white p-1 rounded-lg hover:bg-white/5">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
          </button>
        </div>

        <form @submit.prevent="addSite" class="flex flex-col gap-4">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">Site Name *</label>
              <input v-model="newSite.name" type="text" placeholder="e.g. Retail POS" class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main focus:outline-none focus:border-primary transition-all" required>
            </div>
            <div>
              <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">Domain *</label>
              <input v-model="newSite.domain" type="text" placeholder="localhost" class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main focus:outline-none focus:border-primary transition-all font-mono" required>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">Port *</label>
              <input v-model="newSite.port" type="number" placeholder="8080" class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main focus:outline-none focus:border-primary transition-all font-mono" required>
            </div>
            <div>
              <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">PHP Version *</label>
              <select v-model="newSite.php_version" class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main focus:outline-none focus:border-primary transition-all appearance-none cursor-pointer" required>
                <option v-for="v in phpVersions" :key="v.version" :value="v.version">PHP {{ v.version }}</option>
              </select>
            </div>
          </div>

          <div>
            <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">Document Root *</label>
            <div class="flex gap-2">
              <input v-model="newSite.document_root" type="text" placeholder="e.g. D:\projects\pos\public" class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main focus:outline-none focus:border-primary transition-all font-mono" required>
              <button 
                type="button" 
                @click="browseDirectory('new')" 
                class="bg-white/10 hover:bg-white/20 text-white text-xs font-semibold px-3.5 py-2 rounded-xl transition-colors flex items-center gap-1.5 shrink-0"
                title="Browse folder via Windows Explorer"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"/></svg>
                <span>Browse</span>
              </button>
            </div>
            <p class="text-[11px] text-text-muted mt-1">For Laravel apps, select the <code>public</code> subfolder.</p>
          </div>

          <div class="flex justify-end gap-3 pt-4 border-t border-border-subtle mt-2">
            <button type="button" @click="showAddModal = false" class="px-4 py-2 rounded-xl text-xs font-semibold bg-white/5 hover:bg-white/10 text-text-muted hover:text-white transition-colors">
              Cancel
            </button>
            <button type="submit" class="bg-primary hover:bg-primary-hover text-white text-xs font-semibold px-5 py-2 rounded-xl shadow-lg shadow-primary/20 transition-all">
              Save Site
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Edit Site Modal -->
    <div v-if="editingIndex !== null" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in" @click.self="editingIndex = null">
      <div class="bg-dark-card border border-border-subtle rounded-2xl shadow-2xl w-full max-w-xl p-6">
        <div class="flex justify-between items-center mb-5 pb-3 border-b border-border-subtle">
          <div>
            <h2 class="text-lg font-bold text-white">Edit Site: {{ editSite.name }}</h2>
            <p class="text-xs text-text-muted mt-0.5">Modify virtual host settings</p>
          </div>
          <button @click="editingIndex = null" class="text-text-muted hover:text-white p-1 rounded-lg hover:bg-white/5">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
          </button>
        </div>

        <form @submit.prevent="saveEdit" class="flex flex-col gap-4">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">Site Name *</label>
              <input v-model="editSite.name" type="text" class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main focus:outline-none focus:border-primary" required>
            </div>
            <div>
              <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">Domain *</label>
              <input v-model="editSite.domain" type="text" class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main focus:outline-none focus:border-primary font-mono" required>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">Port *</label>
              <input v-model="editSite.port" type="number" class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main focus:outline-none focus:border-primary font-mono" required>
            </div>
            <div>
              <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">PHP Version *</label>
              <select v-model="editSite.php_version" class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main focus:outline-none focus:border-primary appearance-none cursor-pointer" required>
                <option v-for="v in phpVersions" :key="v.version" :value="v.version">PHP {{ v.version }}</option>
              </select>
            </div>
          </div>

          <div>
            <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-1">Document Root *</label>
            <div class="flex gap-2">
              <input v-model="editSite.document_root" type="text" class="w-full bg-[#111923] border border-border-subtle rounded-xl px-3.5 py-2 text-xs text-text-main focus:outline-none focus:border-primary font-mono" required>
              <button 
                type="button" 
                @click="browseDirectory('edit')" 
                class="bg-white/10 hover:bg-white/20 text-white text-xs font-semibold px-3.5 py-2 rounded-xl transition-colors flex items-center gap-1.5 shrink-0"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"/></svg>
                <span>Browse</span>
              </button>
            </div>
          </div>

          <div class="flex justify-end gap-3 pt-4 border-t border-border-subtle mt-2">
            <button type="button" @click="editingIndex = null" class="px-4 py-2 rounded-xl text-xs font-semibold bg-white/5 hover:bg-white/10 text-text-muted hover:text-white transition-colors">
              Cancel
            </button>
            <button type="submit" class="bg-primary hover:bg-primary-hover text-white text-xs font-semibold px-5 py-2 rounded-xl shadow-lg transition-all">
              Save Changes
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Confirm Delete Modal -->
    <ConfirmModal 
      :show="confirmDeleteOpen"
      title="Delete Site Configuration"
      :message="`Are you sure you want to remove this site? The site will no longer be served by PHPForge, but its source files on disk will not be deleted.`"
      confirm-text="Delete Site"
      cancel-text="Keep Site"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="confirmDeleteOpen = false"
    />
  </div>
</template>
