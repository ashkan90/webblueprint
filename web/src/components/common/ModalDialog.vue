<template>
  <div class="modal-backdrop" @click="handleBackdropClick">
    <div class="modal-container" @click.stop>
      <div class="modal-header">
        <h3 class="modal-title">{{ title }}</h3>
        <button class="modal-close-btn" @click="$emit('close')">×</button>
      </div>

      <div class="modal-body">
        <slot></slot>
      </div>

      <div class="modal-footer">
        <button class="modal-btn cancel-btn" @click="$emit('close')">Cancel</button>
        <button class="modal-btn confirm-btn" @click="$emit('confirm')">{{ confirmText }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'confirm'): void
}>()

function handleBackdropClick(event: MouseEvent) {
  // If closeOnBackdropClick is true or undefined (default to true)
  if (props.closeOnBackdropClick !== false) {
    event.target === event.currentTarget && emit('close')
  }
}

// Set default prop values
const props = withDefaults(defineProps<{
  title: string
  confirmText?: string
  closeOnBackdropClick?: boolean
}>(), {
  confirmText: 'Create',
  closeOnBackdropClick: true
})
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-container {
  background-color: #2d2d2d;
  border-radius: 6px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
  width: 450px;
  max-width: 95%;
  max-height: 90%;
  display: flex;
  flex-direction: column;
  animation: modal-fade-in 0.2s ease-out;
}

@keyframes modal-fade-in {
  from { opacity: 0; transform: translateY(-20px); }
  to { opacity: 1; transform: translateY(0); }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #444;
}

.modal-title {
  margin: 0;
  font-size: 1.1rem;
  color: #e0e0e0;
}

.modal-close-btn {
  background: none;
  border: none;
  color: #aaa;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
}

.modal-close-btn:hover {
  color: white;
}

.modal-body {
  padding: 16px;
  overflow-y: auto;
  flex: 1;
}

.modal-footer {
  padding: 12px 16px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  border-top: 1px solid #444;
}

.modal-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  transition: background-color 0.2s;
}

.cancel-btn {
  background-color: #555;
  color: white;
}

.cancel-btn:hover {
  background-color: #666;
}

.confirm-btn {
  background-color: var(--accent-blue);
  color: white;
}

.confirm-btn:hover {
  background-color: #2980b9;
}
</style>