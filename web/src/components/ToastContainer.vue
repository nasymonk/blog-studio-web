<script setup lang="ts">
import { useToast } from '@/composables/useToast'
import ToastComponent from './Toast.vue'

const { toasts, remove } = useToast()
</script>

<template>
  <Teleport to="body">
    <div class="fixed top-4 right-4 z-[100] flex flex-col gap-2 pointer-events-none">
      <div class="pointer-events-auto">
        <TransitionGroup name="toast">
          <ToastComponent
            v-for="toast in toasts"
            :key="toast.id"
            :toast="toast"
            @dismiss="remove"
          />
        </TransitionGroup>
      </div>
    </div>
  </Teleport>
</template>

<style>
.toast-enter-active {
  animation: toast-in 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.toast-leave-active {
  animation: toast-out 0.2s ease-in forwards;
}
.toast-move {
  transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes toast-in {
  from {
    opacity: 0;
    transform: translateX(100%);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

@keyframes toast-out {
  from {
    opacity: 1;
    transform: translateX(0);
  }
  to {
    opacity: 0;
    transform: translateX(100%);
  }
}
</style>
