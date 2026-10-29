<script setup>
import { ref } from 'vue'
import Sidebar from './components/Sidebar.vue'
import Dashboard from './components/Dashboard.vue'
import PHPVersions from './components/PHPVersions.vue'
import Sites from './components/Sites.vue'
import Settings from './components/Settings.vue'
import ToastContainer from './components/ToastContainer.vue'

const currentPage = ref('dashboard')

const handleNavigate = (page) => {
  currentPage.value = page
}
</script>

<template>
  <div class="flex h-screen w-screen bg-dark-bg text-text-main font-sans overflow-hidden selection:bg-primary/20">
    <Sidebar @navigate="handleNavigate" />
    
    <main class="flex-grow p-6 sm:p-8 overflow-y-auto">
      <Dashboard v-if="currentPage === 'dashboard'" @navigate="handleNavigate" />
      
      <Sites v-else-if="currentPage === 'sites'" />
      
      <PHPVersions v-else-if="currentPage === 'php'" />
      
      <Settings v-else-if="currentPage === 'settings'" />
    </main>

    <!-- Global Toasts -->
    <ToastContainer />
  </div>
</template>

<style>
@keyframes fadeIn {
  from { opacity: 0.85; }
  to { opacity: 1; }
}
.animate-fade-in {
  animation: fadeIn 0.15s ease-out;
}
</style>
