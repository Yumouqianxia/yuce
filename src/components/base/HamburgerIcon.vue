<template>
  <div class="hamburger-icon" :class="{ 'is-active': isActive }" @click="toggleMenu">
    <span class="line"></span>
    <span class="line"></span>
    <span class="line"></span>
  </div>
</template>

<script lang="ts" setup>
defineProps({
  isActive: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['toggle'])

const toggleMenu = () => {
  emit('toggle')
}
</script>

<style scoped>
.hamburger-icon {
  display: none;
  flex-direction: column;
  justify-content: space-between;
  width: 24px;
  height: 20px;
  cursor: pointer;
  transition: all 0.3s;
  position: relative;
}

.line {
  display: block;
  width: 100%;
  height: 2px;
  background-color: var(--text-primary);
  border-radius: 2px;
  transition: all 0.3s ease;
}

.hamburger-icon.is-active .line:nth-child(1) {
  transform: translateY(9px) rotate(45deg);
}

.hamburger-icon.is-active .line:nth-child(2) {
  opacity: 0;
}

.hamburger-icon.is-active .line:nth-child(3) {
  transform: translateY(-9px) rotate(-45deg);
}

@media (max-width: 768px) {
  .hamburger-icon {
    display: flex;
  }
}
</style>
