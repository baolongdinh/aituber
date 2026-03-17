const fs = require('fs');
const path = require('path');

const COMPOSABLE_TEMPLATE = `import { ref, shallowRef } from 'vue'
import { ethers } from 'ethers'

export function use{contractName}(contractAddress, providerOrSigner) {
  const contract = shallowRef(null)
  const isLoading = ref(false)
  const error = ref(null)

  const abi = {abiJson}

  if (contractAddress && providerOrSigner) {
    contract.value = new ethers.Contract(contractAddress, abi, providerOrSigner)
  }

  {methods}

  return {
    contract,
    isLoading,
    error,
    {returnMethods}
  }
}
`;

function generateMethod(item) {
    const isRead = item.stateMutability === 'view' || item.stateMutability === 'pure';
    const inputs = item.inputs.map((input, i) => input.name || `arg${i}`).join(', ');

    if (isRead) {
        return `  async function ${item.name}(${inputs}) {
    if (!contract.value) return
    isLoading.value = true
    try {
      return await contract.value.${item.name}(${inputs})
    } catch (err) {
      error.value = err
      throw err
    } finally {
      isLoading.value = false
    }
  }`;
    } else {
        return `  async function ${item.name}(${inputs}, overrides = {}) {
    if (!contract.value) return
    isLoading.value = true
    try {
      const tx = await contract.value.${item.name}(${inputs}, overrides)
      return await tx.wait()
    } catch (err) {
      error.value = err
      throw err
    } finally {
      isLoading.value = false
    }
  }`;
    }
}

const args = process.argv.slice(2);
const abiPathArg = args.find(a => a.startsWith('--abi='));
const nameArg = args.find(a => a.startsWith('--name='));

if (!abiPathArg || !nameArg) {
    console.log('Usage: node web3-client-gen.js --abi=path/to/abi.json --name=MyToken');
    process.exit(1);
}

const abiPath = abiPathArg.split('=')[1];
const contractName = nameArg.split('=')[1];

try {
    const abi = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
    const functions = abi.filter(item => item.type === 'function');

    const methods = functions.map(generateMethod).join('\n\n');
    const returnMethods = functions.map(f => f.name).join(',\n    ');

    const content = COMPOSABLE_TEMPLATE
        .replace('{contractName}', contractName)
        .replace('{abiJson}', JSON.stringify(abi, null, 2))
        .replace('{methods}', methods)
        .replace('{returnMethods}', returnMethods);

    const outputPath = `src/composables/use${contractName}.js`;
    const outputDir = path.dirname(outputPath);
    if (!fs.existsSync(outputDir)) fs.mkdirSync(outputDir, { recursive: true });

    fs.writeFileSync(outputPath, content);
    console.log(`✅ Generated Web3 Composable: ${outputPath}`);
} catch (err) {
    console.error('❌ Error generating Web3 client:', err.message);
    process.exit(1);
}
