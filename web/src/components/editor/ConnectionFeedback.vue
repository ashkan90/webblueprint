<template>
  <div
      v-if="show"
      class="connection-feedback"
      :style="{ left: position.x + 'px', top: position.y + 'px' }"
      :class="{ 'valid': result.valid, 'invalid': !result.valid }"
  >
    <div class="feedback-icon">
      <span v-if="result.valid">✓</span>
      <span v-else>✗</span>
    </div>
    <div class="feedback-message">
      <template v-if="result.valid">
        <div class="feedback-title">Valid Connection</div>
      </template>
      <template v-else>
        <div class="feedback-title">Invalid Connection</div>
        <div class="feedback-reason">{{ result.reason }}</div>
        <div v-if="result.suggestedFix" class="feedback-suggestion">
          Suggestion: {{ result.suggestedFix }}
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps } from 'vue';

interface ValidationResult {
  valid: boolean;
  reason?: string;
  suggestedFix?: string;
}

interface Position {
  x: number;
  y: number;
}

const props = defineProps<{
  show: boolean;
  position: Position;
  result: ValidationResult;
}>();
</script>

<style scoped>
.connection-feedback {
  position: fixed;
  background-color: #333;
  border-radius: 4px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
  padding: 8px 12px;
  max-width: 300px;
  z-index: 1000;
  pointer-events: none;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  transform: translate(-50%, -100%);
  margin-top: -10px;
  transition: opacity 0.2s, transform 0.2s;
  opacity: 0.9;
}

.connection-feedback.valid {
  border-left: 4px solid #4caf50;
}

.connection-feedback.invalid {
  border-left: 4px solid #f44336;
}

.feedback-icon {
  font-size: 16px;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.valid .feedback-icon {
  background-color: rgba(76, 175, 80, 0.2);
  color: #4caf50;
}

.invalid .feedback-icon {
  background-color: rgba(244, 67, 54, 0.2);
  color: #f44336;
}

.feedback-message {
  flex: 1;
}

.feedback-title {
  font-weight: 500;
  margin-bottom: 4px;
}

.feedback-reason {
  font-size: 0.9rem;
  color: #f44336;
  margin-bottom: 4px;
}

.feedback-suggestion {
  font-size: 0.8rem;
  color: #aaa;
  font-style: italic;
}
</style>