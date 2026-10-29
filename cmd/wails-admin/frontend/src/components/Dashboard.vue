<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { GetDashboardStats, GetAuditLogs, RestartCoreService, GetServiceInfo } from '../../wailsjs/go/main/App'
import { useToast } from '../composables/useToast'

const toast = useToast()

const emit = defineEmits(['navigate'])

const uptime = ref('0s')
const activeWorkers = ref(0)
const totalSites = ref(0)
const memoryAllocated = ref(0)
const memorySys = ref(0)
const status = ref('Checking...')
const apiPort = ref(8081)
const logs = ref([])
const interval = ref(null)
const isRestarting = ref(false)
const isReloading = ref(false)
const serviceInfo = ref({ installed: false, running: false })
const logFilter = ref('ALL')
const autoRefresh = ref(true)

const fetchServiceInfo = async () => {
  try {
    const svc = await GetServiceInfo()
    if (svc) serviceInfo.value = svc
  } catch (_) {}
}

const fetchData = async () => {
  try {
    const stats = await GetDashboardStats()
    uptime.value = stats.uptime || '0s'
    activeWorkers.value = stats.active_php_workers || 0
    totalSites.value = stats.total_sites || 0
    memoryAllocated.value = stats.memory_allocated_mb || 0
    memorySys.value = stats.memory_sys_mb || 0
    status.value = stats.status || 'Offline'
    if (stats.api_port) apiPort.value = stats.api_port

    const recentLogs = await GetAuditLogs()
    // Only update if logs changed to avoid unnecessary DOM reconciliations
    if (recentLogs && recentLogs.length > 0) {
      if (logs.value.length === 0 || logs.value[0].message !== recentLogs[0].message || logs.value.length !== recentLogs.length) {
        logs.value = recentLogs
      }
    }
  } catch (err) {
    status.value = 'Offline'
  }
}

const handleReload = async () => {
  isReloading.value = true
  try {
    await RestartCoreService()
    toast.success("Core Service configuration reloaded successfully!")
    await fetchData()
    await fetchServiceInfo()
  } catch (err) {
    toast.error("Failed to reload Core Service: " + err)
  } finally {
    isReloading.value = false
  }
}

const filteredLogs = computed(() => {
  if (logFilter.value === 'ALL') return logs.value
  return logs.value.filter(l => l.type.toUpperCase() === logFilter.value)
})

const toggleAutoRefresh = () => {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) {
    fetchData()
    interval.value = setInterval(fetchData, 4000)
    toast.info("Auto-refresh enabled (every 4s)")
  } else {
    if (interval.value) clearInterval(interval.value)
    interval.value = null
    toast.info("Auto-refresh paused")
  }
}

onMounted(() => {
  fetchData()
  fetchServiceInfo()
  if (autoRefresh.value) {
    interval.value = setInterval(fetchData, 4000)
  }
})

onUnmounted(() => {
  if (interval.value) clearInterval(interval.value)
})
</script>

<template>
  <div class="flex flex-col gap-6 animate-fade-in max-w-7xl mx-auto">
    <!-- Header with Live Status & Controls -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 pb-2 border-b border-border-subtle/80">
      <div>
        <h1 class="text-2xl font-bold text-white tracking-tight flex items-center gap-2.5">
          Server Overview
          <span 
            class="text-xs px-2.5 py-0.5 rounded-full font-mono font-semibold"
            :class="status === 'OK' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20'"
          >
            REST API :{{ apiPort }}
          </span>
        </h1>
        <p class="text-xs text-text-muted mt-1">Real-time telemetry and process metrics for PHPForge Web Engine</p>
      </div>

      <div class="flex items-center gap-3">
        <!-- Live Status Pill -->
        <div 
          class="inline-flex items-center gap-2 px-3 py-1.5 rounded-xl text-xs font-semibold border backdrop-blur-sm shadow-sm"
          :class="status === 'OK' ? 'bg-emerald-500/10 text-emerald-300 border-emerald-500/30' : 'bg-red-500/10 text-red-300 border-red-500/30'"
        >
          <span 
            class="w-2 h-2 rounded-full shadow-sm" 
            :class="status === 'OK' ? 'bg-emerald-400 shadow-emerald-400/50' : 'bg-red-400 shadow-red-400/50'"
          ></span>
          Core: {{ status === 'OK' ? 'Operational' : status }}
        </div>

        <!-- Reload Button -->
        <button 
          @click="handleReload" 
          :disabled="isReloading"
          class="flex items-center gap-2 bg-primary hover:bg-primary-hover disabled:opacity-50 text-white text-xs font-semibold py-2 px-4 rounded-xl shadow-lg shadow-primary/20 hover:-translate-y-0.5 transition-all cursor-pointer"
          title="Reload configuration and apply changes without dropping connections"
        >
          <svg 
            xmlns="http://www.w3.org/2000/svg" 
            width="14" 
            height="14" 
            viewBox="0 0 24 24" 
            fill="none" 
            stroke="currentColor" 
            stroke-width="2.5" 
            stroke-linecap="round" 
            stroke-linejoin="round"
            :class="{ 'animate-spin': isReloading }"
          >
            <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/>
            <path d="M3 3v5h5"/>
            <path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16"/>
            <path d="M16 21h5v-5"/>
          </svg>
          <span>{{ isReloading ? 'Reloading...' : 'Reload Sites' }}</span>
        </button>
      </div>
    </div>

    <!-- System Deployment Status Banner -->
    <div class="card p-4 rounded-2xl bg-gradient-to-r from-dark-card to-[#192435] border-border-subtle flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
      <div class="flex items-center gap-3.5">
        <div class="w-10 h-10 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-400 flex items-center justify-center shrink-0">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="8" rx="2" ry="2"/><rect x="2" y="14" width="20" height="8" rx="2" ry="2"/><line x1="6" y1="6" x2="6.01" y2="6"/><line x1="6" y1="18" x2="6.01" y2="18"/></svg>
        </div>
        <div>
          <div class="text-sm font-bold text-white flex items-center gap-2">
            Execution Mode: 
            <span v-if="serviceInfo.running" class="text-emerald-400 font-semibold">Windows Service (Auto-Start)</span>
            <span v-else-if="status === 'OK'" class="text-amber-400 font-semibold">Console Mode (Standalone)</span>
            <span v-else class="text-text-muted font-semibold">Service Stopped</span>
          </div>
          <p class="text-xs text-text-muted mt-0.5">
            {{ serviceInfo.installed ? 'Installed as Windows Service (PHForgeCoreSvc)' : 'Running directly via console process or start-console.bat' }}
          </p>
        </div>
      </div>
      <div class="text-xs text-text-muted font-mono flex items-center gap-2">
        <span class="w-2 h-2 rounded-full" :class="status === 'OK' ? 'bg-emerald-400' : 'bg-red-400'"></span>
        Localhost FastCGI Pool Ready
      </div>
    </div>

    <!-- 4 Key Stat Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- Card 1: Uptime -->
      <div class="card p-5 rounded-2xl relative overflow-hidden group hover:border-emerald-500/40 transition-all duration-300">
        <div class="flex items-start justify-between">
          <div>
            <span class="text-[11px] font-bold uppercase tracking-wider text-text-muted">Uptime</span>
            <h3 class="text-2xl font-extrabold text-white mt-1 font-mono tracking-tight">{{ uptime }}</h3>
          </div>
          <div class="w-11 h-11 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center justify-center shrink-0">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          </div>
        </div>
        <div class="mt-4 flex items-center gap-1.5 text-xs text-emerald-400 font-medium">
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
          <span>Core service active</span>
        </div>
      </div>

      <!-- Card 2: Sites -->
      <div class="card p-5 rounded-2xl relative overflow-hidden group hover:border-blue-500/40 transition-all duration-300">
        <div class="flex items-start justify-between">
          <div>
            <span class="text-[11px] font-bold uppercase tracking-wider text-text-muted">Active Sites</span>
            <h3 class="text-2xl font-extrabold text-white mt-1 font-mono tracking-tight">{{ totalSites }}</h3>
          </div>
          <div class="w-11 h-11 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-400 flex items-center justify-center shrink-0">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" x2="22" y1="12" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
          </div>
        </div>
        <div class="mt-4 flex items-center gap-1.5 text-xs text-blue-400 font-medium">
          <button @click="$emit('navigate', 'sites')" class="hover:underline flex items-center gap-1">
            <span>Manage virtual hosts</span> &rarr;
          </button>
        </div>
      </div>

      <!-- Card 3: PHP Workers -->
      <div class="card p-5 rounded-2xl relative overflow-hidden group hover:border-amber-500/40 transition-all duration-300">
        <div class="flex items-start justify-between">
          <div>
            <span class="text-[11px] font-bold uppercase tracking-wider text-text-muted">FastCGI Workers</span>
            <h3 class="text-2xl font-extrabold text-white mt-1 font-mono tracking-tight">{{ activeWorkers }}</h3>
          </div>
          <div class="w-11 h-11 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-400 flex items-center justify-center shrink-0">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/><path d="M12 18v-6"/><path d="M8 18v-1"/><path d="M16 18v-3"/></svg>
          </div>
        </div>
        <div class="mt-4 flex items-center gap-1.5 text-xs text-amber-400 font-medium">
          <span class="w-1.5 h-1.5 rounded-full bg-amber-400"></span>
          <span>Concurrent php-cgi pool</span>
        </div>
      </div>

      <!-- Card 4: Memory -->
      <div class="card p-5 rounded-2xl relative overflow-hidden group hover:border-purple-500/40 transition-all duration-300">
        <div class="flex items-start justify-between">
          <div>
            <span class="text-[11px] font-bold uppercase tracking-wider text-text-muted">RAM Usage</span>
            <h3 class="text-2xl font-extrabold text-white mt-1 font-mono tracking-tight">{{ memoryAllocated }} <span class="text-xs font-normal text-text-muted">MB</span></h3>
          </div>
          <div class="w-11 h-11 rounded-xl bg-purple-500/10 border border-purple-500/20 text-purple-400 flex items-center justify-center shrink-0">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
          </div>
        </div>
        <div class="mt-4 flex items-center justify-between text-xs text-text-muted">
          <span>Sys Reserved:</span>
          <span class="font-mono text-white font-medium">{{ memorySys }} MB</span>
        </div>
      </div>
    </div>

    <!-- Audit Log & Event Stream -->
    <div class="card p-6 rounded-2xl">
      <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 pb-4 mb-4 border-b border-border-subtle">
        <div class="flex items-center gap-3">
          <h2 class="text-lg font-bold text-white tracking-tight flex items-center gap-2">
            Activity Stream
            <span class="text-xs font-mono font-normal px-2 py-0.5 rounded-full bg-white/5 border border-border-subtle text-text-muted">
              audit.log
            </span>
          </h2>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <!-- Filters -->
          <div class="flex rounded-lg bg-[#111923] p-1 border border-border-subtle text-xs">
            <button 
              v-for="filter in ['ALL', 'SYSTEM', 'CONFIG', 'ERROR', 'ACCESS']"
              :key="filter"
              @click="logFilter = filter"
              class="px-2.5 py-1 rounded-md font-semibold transition-colors"
              :class="logFilter === filter ? 'bg-primary text-white shadow-sm' : 'text-text-muted hover:text-white'"
            >
              {{ filter }}
            </button>
          </div>

          <!-- Auto-refresh button -->
          <button 
            @click="toggleAutoRefresh"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border text-xs font-medium transition-colors cursor-pointer"
            :class="autoRefresh ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300' : 'bg-white/5 border-border-subtle text-text-muted'"
            :title="autoRefresh ? 'Auto-refreshing every 4s' : 'Auto-refresh paused'"
          >
            <span class="w-1.5 h-1.5 rounded-full" :class="autoRefresh ? 'bg-emerald-400' : 'bg-slate-500'"></span>
            <span>{{ autoRefresh ? 'Live (4s)' : 'Paused' }}</span>
          </button>
        </div>
      </div>

      <!-- Logs List Container -->
      <div v-if="filteredLogs.length === 0" class="text-center py-12 text-text-muted text-sm">
        No {{ logFilter === 'ALL' ? '' : logFilter }} events recorded in this session.
      </div>
      <div v-else class="flex flex-col gap-2 max-h-80 overflow-y-auto pr-2 custom-scrollbar">
        <div 
          v-for="(log, idx) in filteredLogs" 
          :key="log.time + '-' + idx" 
          class="flex items-start gap-3 p-3 rounded-xl bg-[#111923]/60 border border-border-subtle/50 hover:border-border-subtle hover:bg-[#111923]"
        >
          <!-- Badge -->
          <span 
            class="shrink-0 text-[10px] font-mono font-bold px-2 py-0.5 rounded-md uppercase"
            :class="{
              'bg-red-500/15 text-red-400 border border-red-500/20': log.type === 'ERROR' || log.type === 'FATAL',
              'bg-emerald-500/15 text-emerald-400 border border-emerald-500/20': log.type === 'SYSTEM',
              'bg-blue-500/15 text-blue-400 border border-blue-500/20': log.type === 'CONFIG',
              'bg-purple-500/15 text-purple-400 border border-purple-500/20': log.type === 'ACCESS',
              'bg-white/10 text-text-muted': log.type === 'INFO'
            }"
          >
            {{ log.type }}
          </span>

          <!-- Message -->
          <div class="flex-grow min-w-0">
            <p class="text-xs text-text-main font-mono break-all m-0 leading-relaxed">{{ log.message }}</p>
          </div>

          <!-- Timestamp -->
          <span class="shrink-0 text-[11px] font-mono text-text-muted opacity-60">
            {{ log.time }}
          </span>
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
