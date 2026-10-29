import { ref } from 'vue'

const toasts = ref([])
let nextId = 1

export function useToast() {
  const show = (message, type = 'info', duration = 4000) => {
    const id = nextId++
    toasts.value.push({ id, message, type })
    if (duration > 0) {
      setTimeout(() => {
        remove(id)
      }, duration)
    }
    return id
  }

  const remove = (id) => {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  const success = (msg, duration) => show(msg, 'success', duration)
  const error = (msg, duration) => show(msg, 'error', duration)
  const info = (msg, duration) => show(msg, 'info', duration)

  return {
    toasts,
    show,
    remove,
    success,
    error,
    info
  }
}
