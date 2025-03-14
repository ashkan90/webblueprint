<template>
  <div class="home-view">
    <HomeContent />
  </div>
</template>

<script setup lang="ts">
import HomeContent from '../components/workspace/HomeContent.vue';
</script>

<style scoped>
.home-view {
  flex: 1;
  height: calc(100vh - 50px);
  overflow-y: auto;
  color: var(--text-color);
}
</style>