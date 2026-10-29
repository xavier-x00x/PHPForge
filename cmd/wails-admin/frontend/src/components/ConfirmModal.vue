<script setup>
defineProps({
  show: Boolean,
  title: {
    type: String,
    default: 'Confirm Action'
  },
  message: {
    type: String,
    default: 'Are you sure you want to proceed?'
  },
  confirmText: {
    type: String,
    default: 'Confirm'
  },
  cancelText: {
    type: String,
    default: 'Cancel'
  },
  danger: {
    type: Boolean,
    default: false
  }
})

defineEmits(['confirm', 'cancel'])
</script>

<template>
  <Transition name="fade">
    <div 
      v-if="show" 
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
      @click.self="$emit('cancel')"
    >
      <div class="bg-dark-card border border-border-subtle rounded-2xl shadow-2xl w-full max-w-md p-6 animate-scale-in">
        <div class="flex items-center gap-3.5 mb-3">
          <div 
            class="w-10 h-10 rounded-xl flex items-center justify-center shrink-0"
            :class="danger ? 'bg-red-500/15 text-danger' : 'bg-primary/15 text-primary'"
          >
            <svg v-if="danger" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
          </div>
          <div>
            <h3 class="text-base font-bold text-text-main m-0">{{ title }}</h3>
          </div>
        </div>

        <p class="text-sm text-text-muted mb-6 leading-relaxed pl-13">
          {{ message }}
        </p>

        <div class="flex items-center justify-end gap-3 pt-3 border-t border-border-subtle">
          <button 
            type="button" 
            @click="$emit('cancel')"
            class="px-4 py-2 rounded-lg text-sm font-medium bg-white/5 hover:bg-white/10 text-text-muted hover:text-text-main transition-colors"
          >
            {{ cancelText }}
          </button>
          <button 
            type="button" 
            @click="$emit('confirm')"
            class="px-5 py-2 rounded-lg text-sm font-semibold text-white shadow-lg transition-all"
            :class="danger ? 'bg-danger hover:bg-red-600 shadow-red-900/30' : 'bg-primary hover:bg-primary-hover shadow-blue-900/30'"
          >
            {{ confirmText }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
@keyframes scaleIn {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}
.animate-scale-in {
  animation: scaleIn 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}
</style>
