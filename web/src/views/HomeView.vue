<template>
  <div class="home-view">
    <div class="hero">
      <h1>WebBlueprint</h1>
      <p class="subtitle">Visual Programming for the Web</p>

      <div class="actions">
        <router-link to="/content" class="btn primary">Open Content Browser</router-link>
        <router-link to="/editor" class="btn">Create New Blueprint</router-link>
        <button class="btn" @click="openLoadModal">Load Blueprint</button>
      </div>
    </div>

    <div class="features">
      <div class="feature-card">
        <div class="feature-icon">🧩</div>
        <h3>Visual Programming</h3>
        <p>Create web applications by connecting nodes visually without writing code</p>
      </div>

      <div class="feature-card">
        <div class="feature-icon">🔄</div>
        <h3>Real-time Execution</h3>
        <p>See your application execute in real-time with visual feedback</p>
      </div>

      <div class="feature-card">
        <div class="feature-icon">🔍</div>
        <h3>Interactive Debugging</h3>
        <p>Inspect data flow and execution with powerful debugging tools</p>
      </div>
    </div>

    <!-- Load Blueprint Modal -->
    <div v-if="showLoadModal" class="modal-backdrop">
      <div class="modal">
        <div class="modal-header">
          <h3>Load Blueprint</h3>
          <button class="close-btn" @click="showLoadModal = false">×</button>
        </div>
        <div class="modal-body">
          <div v-if="isLoading">
            <p>Loading blueprints...</p>
          </div>
          <div v-else-if="error">
            <p class="error">{{ error }}</p>
          </div>
          <div v-else-if="blueprints.length === 0">
            <p>No saved blueprints found.</p>
          </div>
          <div v-else class="blueprints-list">
            <div
                v-for="blueprint in blueprints"
                :key="blueprint.id"
                class="blueprint-item"
                @click="loadBlueprint(blueprint)"
            >
              <div class="blueprint-name">{{ blueprint.name }}</div>
              <div class="blueprint-info">
                <span class="blueprint-date">Last modified: {{ formatDate(blueprint.updatedAt) }}</span>
                <span class="blueprint-nodes">{{ blueprint.nodes.length }} nodes</span>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="showLoadModal = false">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

// Router
const router = useRouter()

// State
const showLoadModal = ref(false)
const blueprints = ref<any[]>([])
const isLoading = ref(false)
const error = ref<string | null>(null)

// Methods
function openLoadModal() {
  showLoadModal.value = true
  loadBlueprints()
}

async function loadBlueprints() {
  isLoading.value = true
  error.value = null

  try {
    const response = await fetch('/api/blueprints')
    if (!response.ok) {
      throw new Error('Failed to load blueprints')
    }

    blueprints.value = await response.json()
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err)
    console.error('Error loading blueprints:', err)
  } finally {
    isLoading.value = false
  }
}

function loadBlueprint(blueprint: any) {
  showLoadModal.value = false
  router.push(`/editor/${blueprint.id}`)
}

function formatDate(dateString: string | null): string {
  if (!dateString) return 'Unknown date'

  const date = new Date(dateString)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
}

// Initialize
onMounted(() => {
  // Any initialization logic
})
</script>

<style scoped>
.home-view {
  height: calc(100vh - 50px);
  overflow-y: auto;
  color: var(--text-color);
}

.hero {
  padding: 60px 20px;
  text-align: center;
  background-color: #1a1a1a;
}

h1 {
  font-size: 3rem;
  margin-bottom: 10px;
  color: var(--accent-blue);
}

.subtitle {
  font-size: 1.5rem;
  margin-bottom: 30px;
  color: #aaa;
}

.actions {
  display: flex;
  justify-content: center;
  gap: 20px;
}

.btn {
  background-color: #444;
  border: none;
  color: white;
  padding: 12px 24px;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  display: inline-block;
  text-decoration: none;
  transition: background-color 0.2s;
}

.btn:hover {
  background-color: #555;
}

.btn.primary {
  background-color: var(--accent-blue);
}

.btn.primary:hover {
  background-color: #2980b9;
}

.features {
  display: flex;
  justify-content: center;
  gap: 30px;
  padding: 60px 20px;
}

.feature-card {
  background-color: #2d2d2d;
  border-radius: 8px;
  padding: 20px;
  width: 300px;
  text-align: center;
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.3);
}

.feature-icon {
  font-size: 3rem;
  margin-bottom: 20px;
}

.feature-card h3 {
  font-size: 1.5rem;
  margin-bottom: 10px;
}

.feature-card p {
  color: #bbb;
  line-height: 1.5;
}

/* Modal styles */
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

.modal {
  background-color: #2d2d2d;
  border-radius: 8px;
  width: 500px;
  max-width: 90%;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid #3d3d3d;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.2rem;
}

.close-btn {
  background: none;
  border: none;
  color: #aaa;
  font-size: 1.5rem;
  cursor: pointer;
}

.close-btn:hover {
  color: white;
}

.modal-body {
  padding: 16px;
  max-height: 400px;
  overflow-y: auto;
}

.modal-footer {
  padding: 16px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  border-top: 1px solid #3d3d3d;
}

.error {
  color: var(--accent-red);
}

.blueprints-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.blueprint-item {
  background-color: #333;
  border-radius: 4px;
  padding: 12px;
  cursor: pointer;
  transition: transform 0.1s, background-color 0.2s;
}

.blueprint-item:hover {
  background-color: #444;
  transform: translateY(-2px);
}

.blueprint-name {
  font-weight: 500;
  margin-bottom: 4px;
}

.blueprint-info {
  display: flex;
  justify-content: space-between;
  font-size: 0.8rem;
  color: #aaa;
}
</style>