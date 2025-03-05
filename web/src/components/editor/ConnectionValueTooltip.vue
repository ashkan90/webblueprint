<template>
  <div
      v-if="show"
      class="connection-value-tooltip"
      :style="{ left: position.x + 'px', top: position.y + 'px' }"
  >
    <div class="tooltip-header">
      <div class="pin-info">
        <div class="pin-name">{{ pinName }}</div>
        <div class="pin-type" :style="{ backgroundColor: getPinTypeColor(pinType) }">
          {{ pinTypeName }}
        </div>
      </div>
    </div>

    <div class="tooltip-content">
      <template v-if="displayValue !== undefined">
        <div class="value-preview" :class="`value-type-${pinType}`">
          <template v-if="pinType === 'string'">
            <span class="value-string">"{{ displayValue }}"</span>
          </template>
          <template v-else-if="pinType === 'number'">
            <span class="value-number">{{ displayValue }}</span>
          </template>
          <template v-else-if="pinType === 'boolean'">
            <span class="value-boolean">{{ displayValue }}</span>
          </template>
          <template v-else-if="pinType === 'object' || pinType === 'array'">
            <div class="value-complex">
              {{ displayValue }}
            </div>
          </template>
          <template v-else>
            <span class="value-any">{{ displayValue }}</span>
          </template>
        </div>
      </template>
      <div v-else class="no-value">
        No value available
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {computed, defineProps} from 'vue';
import { PinTypeColors } from '../../types/nodes';

interface Position {
  x: number;
  y: number;
}

const props = defineProps<{
  show: boolean;
  position: Position;
  pinName: string;
  pinType: string;
  pinTypeName: string;
  value: any;
}>();

// Computed property to format the display value
const displayValue = computed(() => {
  if (props.value === undefined || props.value === null) {
    return props.value === null ? 'null' : 'undefined';
  }

  if (props.pinType === 'object' || props.pinType === 'array') {
    try {
      return JSON.stringify(props.value, null, 2);
    } catch (e) {
      return String(props.value);
    }
  }

  return String(props.value);
});

// Helper function to get color for pin types
function getPinTypeColor(type: string): string {
  return PinTypeColors[type] || PinTypeColors['any'];
}
</script>

<style scoped>
.connection-value-tooltip {
  position: fixed;
  background-color: #333;
  border-radius: 4px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.5);
  padding: 0;
  min-width: 200px;
  max-width: 400px;
  z-index: 1000;
  pointer-events: none;
  transform: translate(-50%, -100%);
  margin-top: -10px;
  animation: tooltip-fade-in 0.2s ease-out;
  border: 1px solid #444;
  max-height: 300px;
  overflow-y: auto;
}

.tooltip-header {
  padding: 8px 12px;
  background-color: #444;
  border-bottom: 1px solid #555;
  display: flex;
  justify-content: space-between;
}

.pin-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pin-name {
  font-weight: 500;
  font-size: 0.9rem;
}

.pin-type {
  font-size: 0.7rem;
  padding: 2px 6px;
  border-radius: 10px;
  color: #333;
  font-weight: bold;
}

.tooltip-content {
  padding: 8px 12px;
}

.value-preview {
  font-family: monospace;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 0.9rem;
  max-height: 200px;
  overflow-y: auto;
}

.value-string {
  color: #f0883e;
}

.value-number {
  color: #6ed69a;
}

.value-boolean {
  color: #dc5050;
}

.value-complex {
  color: #8ab4f8;
}

.value-any {
  color: #aaaaaa;
}

.no-value {
  color: #888;
  font-style: italic;
  font-size: 0.9rem;
}

@keyframes tooltip-fade-in {
  from { opacity: 0; transform: translate(-50%, -90%); }
  to { opacity: 1; transform: translate(-50%, -100%); }
}
</style>