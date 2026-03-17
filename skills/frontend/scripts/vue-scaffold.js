const fs = require('fs');
const path = require('path');

const COMPONENT_TEMPLATE = `<script setup>
import { ref, onMounted } from 'vue'

const props = defineProps({
  id: {
    type: [String, Number],
    required: false
  }
})

const isLoading = ref(false)
const error = ref(null)

onMounted(() => {
  console.log('{name} mounted')
})
</script>

<template>
  <div class="{name_lower}-container">
    <h1>{name} Component</h1>
  </div>
</template>

<style scoped>
.{name_lower}-container {
  padding: 1rem;
}
</style>
`;

const STORE_TEMPLATE = `import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const use{name}Store = defineStore('{name_lower}', () => {
  const data = ref([])
  const isLoading = ref(false)

  async function fetchData() {
    isLoading.value = true
    try {
      // API call here
    } finally {
      isLoading.value = false
    }
  }

  return { data, isLoading, fetchData }
})
`;

function scaffold(name) {
  const name_lower = name.toLowerCase();

  const componentPath = `src/components/${name}.vue`;
  const storePath = `src/stores/${name_lower}.store.js`;

  // Create component
  const componentDir = path.dirname(componentPath);
  if (!fs.existsSync(componentDir)) fs.mkdirSync(componentDir, { recursive: true });
  fs.writeFileSync(componentPath, COMPONENT_TEMPLATE.replace(/{name}/g, name).replace(/{name_lower}/g, name_lower));
  console.log(`  ✅ Created ${componentPath}`);

  // Create store
  const storeDir = path.dirname(storePath);
  if (!fs.existsSync(storeDir)) fs.mkdirSync(storeDir, { recursive: true });
  fs.writeFileSync(storePath, STORE_TEMPLATE.replace(/{name}/g, name).replace(/{name_lower}/g, name_lower));
  console.log(`  ✅ Created ${storePath}`);
}

const args = process.argv.slice(2);
const nameArg = args.find(a => a.startsWith('--name='));
if (!nameArg) {
  console.log('Usage: node vue-scaffold.js --name=MyComponent');
  process.exit(1);
}

const name = nameArg.split('=')[1];
console.log(`🏗️  Scaffolding Vue component: ${name}...`);
scaffold(name);
console.log('\n🚀 Done!');
