const fs = require('fs');
const path = require('path');

const E2E_TEMPLATE = `import { test, expect } from '@playwright/test';

test.describe('{domain_title} Flow', () => {{
  test.beforeEach(async ({{ page }}) => {{
    // Setup blockchain wallet state here if needed
    await page.goto('/');
  }});

  test('should show {domain_lower} list', async ({{ page }}) => {{
    await page.click('nav >> text={domain_title}s');
    await expect(page).toHaveURL(/.*{domain_lower}/);
    await expect(page.locator('h1')).toContainText('{domain_title}s');
  }});

  test('should allow interacting with {domain_lower}', async ({{ page }}) => {{
    // Add specific interaction test for {domain_lower}
  }});
}});
`;

const args = process.argv.slice(2);
const domainArg = args.find(a => a.startsWith('--domain='));

if (!domainArg) {
    console.log('Usage: node e2e-scaffold.js --domain=product');
    process.exit(1);
}

const domain = domainArg.split('=')[1];
const domain_title = domain.charAt(0).toUpperCase() + domain.slice(1);
const domain_lower = domain.toLowerCase();

const outputPath = `frontend/tests/e2e/${domain_lower}.spec.js`;
const outputDir = path.dirname(outputPath);
if (!fs.existsSync(outputDir)) fs.mkdirSync(outputDir, { recursive: true });

fs.writeFileSync(outputPath, E2E_TEMPLATE.replace(/{domain_title}/g, domain_title).replace(/{domain_lower}/g, domain_lower));
console.log(`✅ Generated E2E Test Scaffold: ${outputPath}`);
