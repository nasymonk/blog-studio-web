<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'

const props = defineProps<{
  isTyping?: boolean
  showPassword?: boolean
  passwordLength?: number
}>()

const mouseX = ref(0)
const mouseY = ref(0)

// Blink states
const coralBlinking = ref(false)
const slateBlinking = ref(false)
const amberBlinking = ref(false)
const sageBlinking = ref(false)

// Interaction states
const lookingAtEachOther = ref(false)
const coralPeeking = ref(false)

// Refs for position calculation
const coralEl = ref<HTMLElement | null>(null)
const slateEl = ref<HTMLElement | null>(null)
const amberEl = ref<HTMLElement | null>(null)
const sageEl = ref<HTMLElement | null>(null)

function onMouseMove(e: MouseEvent) {
  mouseX.value = e.clientX
  mouseY.value = e.clientY
}

onMounted(() => window.addEventListener('mousemove', onMouseMove))
onUnmounted(() => window.removeEventListener('mousemove', onMouseMove))

// Blink scheduler
const blinkTimers: ReturnType<typeof setTimeout>[] = []

function scheduleBlink(setBlinking: (v: boolean) => void) {
  const tick = () => {
    const delay = Math.random() * 4000 + 3000
    const t1 = setTimeout(() => {
      setBlinking(true)
      const t2 = setTimeout(() => {
        setBlinking(false)
        tick()
      }, 150)
      blinkTimers.push(t2)
    }, delay)
    blinkTimers.push(t1)
  }
  tick()
}

onUnmounted(() => {
  blinkTimers.forEach(clearTimeout)
})

onMounted(() => {
  scheduleBlink((v) => { coralBlinking.value = v })
  scheduleBlink((v) => { slateBlinking.value = v })
  scheduleBlink((v) => { amberBlinking.value = v })
  scheduleBlink((v) => { sageBlinking.value = v })
})

// Looking at each other on typing
watch(() => props.isTyping, (val) => {
  if (val) {
    lookingAtEachOther.value = true
    setTimeout(() => { lookingAtEachOther.value = false }, 800)
  }
})

// Coral peeking on visible password
watch([() => props.passwordLength, () => props.showPassword], ([len, show]) => {
  if (len && show) {
    const peek = () => {
      setTimeout(() => {
        coralPeeking.value = true
        setTimeout(() => { coralPeeking.value = false }, 800)
      }, Math.random() * 3000 + 2000)
    }
    peek()
  }
})

interface Position { faceX: number; faceY: number; bodySkew: number }

function calcPosition(el: HTMLElement | null): Position {
  if (!el) return { faceX: 0, faceY: 0, bodySkew: 0 }
  const rect = el.getBoundingClientRect()
  const cx = rect.left + rect.width / 2
  const cy = rect.top + rect.height / 3
  const dx = mouseX.value - cx
  const dy = mouseY.value - cy
  return {
    faceX: Math.max(-12, Math.min(12, dx / 25)),
    faceY: Math.max(-8, Math.min(8, dy / 35)),
    bodySkew: Math.max(-5, Math.min(5, -dx / 140)),
  }
}

const coralPos = computed(() => calcPosition(coralEl.value))
const slatePos = computed(() => calcPosition(slateEl.value))
const amberPos = computed(() => calcPosition(amberEl.value))
const sagePos = computed(() => calcPosition(sageEl.value))

const isHiding = computed(() => (props.passwordLength ?? 0) > 0 && !props.showPassword)

function eyeStyle(pos: Position, forceX?: number, forceY?: number) {
  const x = forceX ?? pos.faceX * 0.4
  const y = forceY ?? pos.faceY * 0.4
  return { transform: `translate(${x}px, ${y}px)`, transition: 'transform 0.1s ease-out' }
}
</script>

<template>
  <div class="relative" style="width: 420px; height: 340px;">
    <!-- Coral — tall rounded rectangle, warm -->
    <div
      ref="coralEl"
      class="absolute bottom-0 transition-all duration-700 ease-in-out"
      :style="{
        left: '40px', width: '64px', height: (isTyping || isHiding) ? '340px' : '300px',
        background: 'linear-gradient(180deg, #d4836a 0%, #c4705a 60%, #b5604d 100%)',
        borderRadius: '14px 14px 6px 6px',
        zIndex: 1,
        transformOrigin: 'bottom center',
        transform: (props.passwordLength && props.showPassword)
          ? 'skewX(0deg)'
          : (isTyping || isHiding)
            ? `skewX(${(coralPos.bodySkew || 0) - 10}deg) translateX(30px)`
            : `skewX(${coralPos.bodySkew || 0}deg)`,
      }"
    >
      <!-- Eyes -->
      <div
        class="absolute flex gap-3 transition-all duration-700 ease-in-out"
        :style="{
          left: (props.showPassword && props.passwordLength) ? '10px' : lookingAtEachOther ? '20px' : `${18 + coralPos.faceX}px`,
          top: (props.showPassword && props.passwordLength) ? '30px' : lookingAtEachOther ? '50px' : `${40 + coralPos.faceY}px`,
        }"
      >
        <div class="rounded-full bg-white flex items-center justify-center overflow-hidden" :style="{ width: '14px', height: coralBlinking ? '2px' : '14px', transition: 'height 0.15s' }">
          <div v-if="!coralBlinking" class="rounded-full bg-[#3a2520]" :style="{ width: '6px', height: '6px', ...eyeStyle(coralPos, (props.passwordLength && props.showPassword) ? (coralPeeking ? 3 : -3) : lookingAtEachOther ? 2 : undefined, (props.passwordLength && props.showPassword) ? (coralPeeking ? 3 : -3) : lookingAtEachOther ? 3 : undefined) }" />
        </div>
        <div class="rounded-full bg-white flex items-center justify-center overflow-hidden" :style="{ width: '14px', height: coralBlinking ? '2px' : '14px', transition: 'height 0.15s' }">
          <div v-if="!coralBlinking" class="rounded-full bg-[#3a2520]" :style="{ width: '6px', height: '6px', ...eyeStyle(coralPos, (props.passwordLength && props.showPassword) ? (coralPeeking ? 3 : -3) : lookingAtEachOther ? 2 : undefined, (props.passwordLength && props.showPassword) ? (coralPeeking ? 3 : -3) : lookingAtEachOther ? 3 : undefined) }" />
        </div>
      </div>
      <!-- Mouth -->
      <div
        class="absolute w-6 h-[2px] rounded-full transition-all duration-200 ease-out"
        :style="{
          left: (props.passwordLength && props.showPassword) ? '8px' : `${16 + coralPos.faceX}px`,
          top: (props.passwordLength && props.showPassword) ? '65px' : `${68 + coralPos.faceY}px`,
          background: 'rgba(58,37,32,0.35)',
        }"
      />
    </div>

    <!-- Slate — wide oval, cool-toned -->
    <div
      ref="slateEl"
      class="absolute bottom-0 transition-all duration-700 ease-in-out"
      :style="{
        left: '120px', width: '160px', height: '120px',
        background: 'linear-gradient(135deg, #7a9aae 0%, #6b8a9e 50%, #5c7a8e 100%)',
        borderRadius: '50% 50% 12px 12px',
        zIndex: 2,
        transformOrigin: 'bottom center',
        transform: (props.passwordLength && props.showPassword) ? 'skewX(0deg)' : `skewX(${slatePos.bodySkew || 0}deg)`,
      }"
    >
      <!-- Eyes -->
      <div
        class="absolute flex gap-6 transition-all duration-200 ease-out"
        :style="{
          left: (props.passwordLength && props.showPassword) ? '42px' : `${52 + slatePos.faceX}px`,
          top: (props.passwordLength && props.showPassword) ? '30px' : `${32 + slatePos.faceY}px`,
        }"
      >
        <div class="w-3 h-3 rounded-full bg-white/90" :style="eyeStyle(slatePos, (props.passwordLength && props.showPassword) ? -4 : undefined, (props.passwordLength && props.showPassword) ? -3 : undefined)" />
        <div class="w-3 h-3 rounded-full bg-white/90" :style="eyeStyle(slatePos, (props.passwordLength && props.showPassword) ? -4 : undefined, (props.passwordLength && props.showPassword) ? -3 : undefined)" />
      </div>
    </div>

    <!-- Amber — small rounded square, warm golden -->
    <div
      ref="amberEl"
      class="absolute bottom-0 transition-all duration-700 ease-in-out"
      :style="{
        left: '2px', width: '96px', height: '130px',
        background: 'linear-gradient(180deg, #d4b86a 0%, #c4a35a 100%)',
        borderRadius: '16px 16px 6px 6px',
        zIndex: 3,
        transformOrigin: 'bottom center',
        transform: (props.passwordLength && props.showPassword) ? 'skewX(0deg)' : `skewX(${amberPos.bodySkew || 0}deg)`,
      }"
    >
      <!-- Eyes -->
      <div
        class="absolute flex gap-4 transition-all duration-200 ease-out"
        :style="{
          left: (props.passwordLength && props.showPassword) ? '22px' : `${28 + amberPos.faceX}px`,
          top: (props.passwordLength && props.showPassword) ? '45px' : `${50 + amberPos.faceY}px`,
        }"
      >
        <div class="w-2.5 h-2.5 rounded-full bg-[#3a2a10]" :style="eyeStyle(amberPos, (props.passwordLength && props.showPassword) ? -4 : undefined, (props.passwordLength && props.showPassword) ? -3 : undefined)" />
        <div class="w-2.5 h-2.5 rounded-full bg-[#3a2a10]" :style="eyeStyle(amberPos, (props.passwordLength && props.showPassword) ? -4 : undefined, (props.passwordLength && props.showPassword) ? -3 : undefined)" />
      </div>
    </div>

    <!-- Sage — medium pill, soft green -->
    <div
      ref="sageEl"
      class="absolute bottom-0 transition-all duration-700 ease-in-out"
      :style="{
        left: '290px', width: '100px', height: '190px',
        background: 'linear-gradient(180deg, #9aba8e 0%, #8aaa7e 100%)',
        borderRadius: '50px 50px 8px 8px',
        zIndex: 4,
        transformOrigin: 'bottom center',
        transform: (props.passwordLength && props.showPassword) ? 'skewX(0deg)' : `skewX(${sagePos.bodySkew || 0}deg)`,
      }"
    >
      <!-- Eyes -->
      <div
        class="absolute flex gap-5 transition-all duration-200 ease-out"
        :style="{
          left: (props.passwordLength && props.showPassword) ? '18px' : `${28 + sagePos.faceX}px`,
          top: (props.passwordLength && props.showPassword) ? '65px' : `${70 + sagePos.faceY}px`,
        }"
      >
        <div class="rounded-full bg-[#2a3a20] flex items-center justify-center overflow-hidden" :style="{ width: '12px', height: sageBlinking ? '2px' : '12px', transition: 'height 0.15s' }">
          <div v-if="!sageBlinking" class="rounded-full bg-white/80" :style="{ width: '5px', height: '5px', ...eyeStyle(sagePos, (props.passwordLength && props.showPassword) ? -4 : undefined, (props.passwordLength && props.showPassword) ? -3 : undefined) }" />
        </div>
        <div class="rounded-full bg-[#2a3a20] flex items-center justify-center overflow-hidden" :style="{ width: '12px', height: sageBlinking ? '2px' : '12px', transition: 'height 0.15s' }">
          <div v-if="!sageBlinking" class="rounded-full bg-white/80" :style="{ width: '5px', height: '5px', ...eyeStyle(sagePos, (props.passwordLength && props.showPassword) ? -4 : undefined, (props.passwordLength && props.showPassword) ? -3 : undefined) }" />
        </div>
      </div>
      <!-- Mouth -->
      <div
        class="absolute w-8 h-[2px] rounded-full transition-all duration-200 ease-out"
        :style="{
          left: (props.passwordLength && props.showPassword) ? '14px' : `${26 + sagePos.faceX}px`,
          top: (props.passwordLength && props.showPassword) ? '95px' : `${98 + sagePos.faceY}px`,
          background: 'rgba(42,58,32,0.3)',
        }"
      />
    </div>
  </div>
</template>
