<template>
  <div class="json-tree" :style="{ paddingLeft: level > 1 ? '16px' : '0' }">
    <div v-if="isObject" class="object-container">
      <div
          class="expander"
          @click="toggleExpanded"
          :class="{ 'expanded': isExpanded }"
      >
        <span class="toggle-icon">{{ isExpanded ? '▼' : '▶' }}</span>
        <span v-if="label" class="key">{{ label }}:</span>
        <span class="type-label">{{ isArray ? 'Array' : 'Object' }}</span>
        <span class="item-count">{{ itemCount }}</span>
      </div>

      <div v-if="isExpanded" class="children">
        <div v-for="(value, key) in data" :key="key" class="property">
          <JsonTree
              :data="value"
              :level="level + 1"
              :label="key.toString()"
              :expanded="level < 2"
          />
        </div>

        <div v-if="isEmpty" class="empty-notice">
          Empty {{ isArray ? 'array' : 'object' }}
        </div>
      </div>
    </div>

    <div v-else class="primitive-container">
      <span v-if="label" class="key">{{ label }}:</span>
      <span :class="['value', valueType]">{{ displayValue }}</span>
      <span v-if="showType" class="type-hint">{{ valueType }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = withDefaults(defineProps<{
  data: any
  level?: number
  label?: string
  expanded?: boolean
  showType?: boolean
}>(), {
  level: 1,
  label: '',
  expanded: true,
  showType: true
})

// State
const isExpanded = ref(props.expanded)

// Computed
const isObject = computed(() => props.data !== null && typeof props.data === 'object')
const isArray = computed(() => Array.isArray(props.data))
const isEmpty = computed(() => isObject.value && Object.keys(props.data).length === 0)

const itemCount = computed(() => {
  if (!isObject.value) return ''
  const count = Object.keys(props.data).length
  return count === 1 ? '(1 item)' : `(${count} items)`
})

const valueType = computed(() => {
  if (props.data === null) return 'null'
  if (props.data === undefined) return 'undefined'
  if (Array.isArray(props.data)) return 'array'
  return typeof props.data
})

const displayValue = computed(() => {
  if (props.data === null) return 'null'
  if (props.data === undefined) return 'undefined'
  if (typeof props.data === 'string') return `"${props.data}"`
  if (typeof props.data === 'bigint') return `${props.data.toString()}n`
  return String(props.data)
})

// Methods
function toggleExpanded() {
  isExpanded.value = !isExpanded.value
}
</script>

<style scoped>
.json-tree {
  font-family: monospace;
  font-size: 13px;
  line-height: 1.5;
  word-break: break-word;
}

.object-container, .primitive-container {
  padding: 2px 0;
}

.expander {
  cursor: pointer;
  user-select: none;
  display: flex;
  align-items: center;
}

.expander:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.toggle-icon {
  display: inline-block;
  width: 14px;
  height: 14px;
  text-align: center;
  color: #aaa;
}

.key {
  color: #9876aa;
  margin-right: 4px;
}

.value {
  font-weight: 500;
}

.value.string {
  color: #6a8759;
}

.value.number {
  color: #6897bb;
}

.value.boolean {
  color: #cc7832;
}

.value.null, .value.undefined {
  color: #808080;
  font-style: italic;
}

.type-label, .type-hint {
  color: #808080;
  font-style: italic;
  margin-left: 4px;
  font-size: 11px;
}

.item-count {
  color: #808080;
  margin-left: 4px;
  font-size: 11px;
}

.children {
  padding-left: 14px;
}

.empty-notice {
  color: #808080;
  font-style: italic;
  padding: 2px 0;
}
</style>