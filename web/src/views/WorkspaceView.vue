<template>
  <div class="workspace-view">
    <WorkspacePanel v-if="!isEditorView" />
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import WorkspacePanel from '../components/workspace/WorkspacePanel.vue';

const route = useRoute();
const isEditorView = computed(() => {
  return route.name === 'editor' || route.name === 'editor2';
});
</script>

<style scoped>
.workspace-view {
  display: flex;
  height: 100%;
  width: 100%;
  overflow: hidden;
}
</style>