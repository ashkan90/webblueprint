<template>
  <div class="about-view">
    <div class="about-content">
      <h1>About WebBlueprint</h1>

      <section>
        <h2>What is WebBlueprint?</h2>
        <p>
          WebBlueprint is a visual programming system that allows users to create functional web applications by connecting nodes in a graphical interface rather than writing traditional code. It's inspired by Unreal Engine's Blueprint system but adapted specifically for web development.
        </p>
        <p>
          The key idea is to enable both developers and non-developers to build web applications through a visual, flow-based interface where:
        </p>
        <ul>
          <li>Nodes represent different functions (HTTP requests, data processing, DOM manipulation)</li>
          <li>Connections between nodes define how data and execution flow</li>
          <li>Execution happens in real-time on a backend server</li>
        </ul>
      </section>

      <section>
        <h2>Inspiration: Unreal Engine Blueprint</h2>
        <p>
          Unreal Engine's Blueprint system is a visual scripting language integrated into the Unreal game engine. It allows game designers and non-programmers to create gameplay mechanics and systems without writing C++ code. Key aspects that inspired WebBlueprint include:
        </p>
        <ol>
          <li><strong>Node-based visual programming:</strong> Logic represented as nodes with inputs/outputs</li>
          <li><strong>Execution flow:</strong> White triangular arrows show the sequence of operations</li>
          <li><strong>Data flow:</strong> Colored circular connections show how data passes between nodes</li>
          <li><strong>Real-time feedback:</strong> Visual debugging and execution highlighting</li>
          <li><strong>Type system:</strong> Strong typing with validation between connections</li>
        </ol>
        <p>
          Blueprint revolutionized game development by making programming more accessible to artists, designers, and non-technical team members. WebBlueprint aims to bring these same benefits to web development.
        </p>
      </section>

      <section>
        <h2>Value of Visual Programming for Web</h2>
        <p>
          Unlike traditional coding, visual programming in WebBlueprint offers:
        </p>
        <ul>
          <li><strong>Lower barrier to entry:</strong> Non-developers can create functional web applications</li>
          <li><strong>Immediate feedback:</strong> See results as you build</li>
          <li><strong>Visual debugging:</strong> Observe execution flow and data values in real-time</li>
          <li><strong>Reduced syntax errors:</strong> Connection validation prevents many common mistakes</li>
          <li><strong>Component reuse:</strong> Easily reuse patterns across different projects</li>
        </ul>
        <p>
          This approach bridges the gap between no-code tools (which are limited but accessible) and traditional coding (which is powerful but has a steep learning curve).
        </p>
      </section>

      <section>
        <h2>Technical Architecture</h2>
        <p>
          WebBlueprint consists of:
        </p>
        <ul>
          <li><strong>Frontend:</strong> Vue.js-based visual editor for creating and connecting nodes</li>
          <li><strong>Backend:</strong> Go server with an actor-based system for executing node operations</li>
          <li><strong>Communication:</strong> WebSockets for real-time feedback during execution</li>
          <li><strong>Type System:</strong> Strong typing with validation between node connections</li>
          <li><strong>Debugging:</strong> Real-time visualization of data flow and execution state</li>
        </ul>
      </section>

      <section>
        <h2>Use Cases</h2>
        <p>
          WebBlueprint is ideal for:
        </p>
        <ul>
          <li><strong>API-driven applications:</strong> Connect to APIs and transform/visualize data</li>
          <li><strong>Interactive prototypes:</strong> Quickly build working prototypes with real functionality</li>
          <li><strong>Education:</strong> Teach programming concepts visually before moving to text-based code</li>
          <li><strong>Cross-functional teams:</strong> Empower designers and product managers to build simple applications</li>
          <li><strong>Process automation:</strong> Create workflows connecting different web services</li>
        </ul>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
// No script needed for this view
</script>

<style scoped>
.about-view {
  height: calc(100vh - 50px);
  overflow-y: auto;
  padding: 20px;
  display: flex;
  justify-content: center;
}

.about-content {
  max-width: 800px;
  color: var(--text-color);
}

h1 {
  font-size: 2.5rem;
  margin-bottom: 2rem;
  color: var(--accent-blue);
}

h2 {
  font-size: 1.5rem;
  margin-top: 2rem;
  margin-bottom: 1rem;
  color: var(--accent-blue);
}

section {
  margin-bottom: 2rem;
}

p {
  margin-bottom: 1rem;
  line-height: 1.6;
}

ul, ol {
  margin-bottom: 1rem;
  padding-left: 2rem;
}

li {
  margin-bottom: 0.5rem;
  line-height: 1.6;
}

strong {
  color: #ddd;
}
</style>