<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { GetServerStatus } from '../../wailsjs/go/main/App'

const activeMenu = ref('dashboard')
const coreStatus = ref('Checking...')
const isOnline = ref(false)
let timer = null

const emit = defineEmits(['navigate'])

const navigate = (page) => {
  activeMenu.value = page
  emit('navigate', page)
}

const checkStatus = async () => {
  try {
    const res = await GetServerStatus()
    coreStatus.value = res || 'Online'
    isOnline.value = true
  } catch (err) {
    coreStatus.value = 'Offline'
    isOnline.value = false
  }
}

onMounted(() => {
  checkStatus()
  timer = setInterval(checkStatus, 5000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <aside class="w-[260px] h-screen bg-[#0e1722] border-r border-border-subtle flex flex-col box-border shadow-2xl z-20 shrink-0 select-none">
    <!-- Brand Header -->
    <div class="px-6 py-6 border-b border-border-subtle/80 flex items-center justify-between">
      <div>
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 rounded-lg bg-gradient-to-tr from-primary to-blue-400 flex items-center justify-center shadow-lg shadow-primary/20 text-white font-black text-sm">
            PF
          </div>
          <div>
            <h2 class="text-xl font-extrabold tracking-tight text-white leading-none">
              <span class="text-primary">PHP</span>Forge
            </h2>
            <span class="text-[10px] text-text-muted font-semibold tracking-wider uppercase opacity-80">Server Engine</span>
          </div>
        </div>
      </div>
      <span class="text-[10px] font-mono font-semibold px-2 py-0.5 rounded-full bg-white/5 border border-white/10 text-text-muted">
        v1.0
      </span>
    </div>
    
    <!-- Navigation Links -->
    <nav class="flex flex-col gap-1.5 p-4 flex-grow">
      <button 
        :class="[
          'group relative flex items-center gap-3 px-3.5 py-2.5 rounded-xl cursor-pointer text-sm font-semibold transition-all duration-200 text-left',
          activeMenu === 'dashboard' 
            ? 'bg-primary text-white shadow-lg shadow-primary/25' 
            : 'text-text-muted hover:bg-white/5 hover:text-white'
        ]" 
        @click="navigate('dashboard')"
      >
        <svg class="w-4 h-4 transition-transform group-hover:scale-110" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="7" height="9" x="3" y="3" rx="1"/><rect width="7" height="5" x="14" y="3" rx="1"/><rect width="7" height="9" x="14" y="12" rx="1"/><rect width="7" height="5" x="3" y="16" rx="1"/></svg>
        <span>Overview</span>
      </button>

      <button 
        :class="[
          'group relative flex items-center gap-3 px-3.5 py-2.5 rounded-xl cursor-pointer text-sm font-semibold transition-all duration-200 text-left',
          activeMenu === 'sites' 
            ? 'bg-primary text-white shadow-lg shadow-primary/25' 
            : 'text-text-muted hover:bg-white/5 hover:text-white'
        ]" 
        @click="navigate('sites')"
      >
        <svg class="w-4 h-4 transition-transform group-hover:scale-110" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" x2="22" y1="12" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
        <span>Sites & VHosts</span>
      </button>

      <button 
        :class="[
          'group relative flex items-center gap-3 px-3.5 py-2.5 rounded-xl cursor-pointer text-sm font-semibold transition-all duration-200 text-left',
          activeMenu === 'php' 
            ? 'bg-primary text-white shadow-lg shadow-primary/25' 
            : 'text-text-muted hover:bg-white/5 hover:text-white'
        ]" 
        @click="navigate('php')"
      >
        <svg class="w-4 h-4 transition-transform group-hover:scale-110" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m18 16 4-4-4-4"/><path d="m6 8-4 4 4 4"/><path d="m14.5 4-5 16"/></svg>
        <span>PHP Versions</span>
      </button>

      <button 
        :class="[
          'group relative flex items-center gap-3 px-3.5 py-2.5 rounded-xl cursor-pointer text-sm font-semibold transition-all duration-200 text-left mt-auto',
          activeMenu === 'settings' 
            ? 'bg-primary text-white shadow-lg shadow-primary/25' 
            : 'text-text-muted hover:bg-white/5 hover:text-white'
        ]" 
        @click="navigate('settings')"
      >
        <svg class="w-4 h-4 transition-transform group-hover:scale-110" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>
        <span>Settings</span>
      </button>
    </nav>

    <!-- Bottom Status Card -->
    <div class="p-4 border-t border-border-subtle/80 bg-[#0a111a]">
      <div class="p-3 rounded-xl bg-dark-card/60 border border-border-subtle flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <div class="relative flex items-center justify-center">
            <span 
              class="w-2.5 h-2.5 rounded-full shadow-sm"
              :class="isOnline ? 'bg-emerald-400 shadow-emerald-400/50' : 'bg-red-400 shadow-red-400/50'"
            ></span>
          </div>
          <div>
            <div class="text-xs font-bold text-text-main leading-tight">Core Service</div>
            <div class="text-[10px] text-text-muted font-mono">{{ coreStatus }}</div>
          </div>
        </div>
        <button 
          @click="navigate('dashboard')" 
          class="text-text-muted hover:text-white p-1 rounded transition-colors"
          title="View in Dashboard"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
        </button>
      </div>
    </div>
  </aside>
</template>
