import { test, expect } from '@playwright/test';

test('Video generation flow works smoothly from UI to API polling', async ({ page }) => {
  // Mock API /api/generate
  await page.route('**/api/generate', async route => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ job_id: 'test-job-uuid', status: 'processing' }),
    });
  });

  // Mock API /api/status/:id
  let progress = 0;
  await page.route('**/api/status/test-job-uuid', async route => {
    progress += 25;
    if (progress > 100) progress = 100;
    
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        status: progress === 100 ? 'completed' : 'processing',
        progress: progress,
        current_step: `Step ${progress / 25}`,
        video_url: progress === 100 ? '/api/download/test-job-uuid' : null,
      }),
    });
  });

  // Navigate to frontend
  await page.goto('/');

  // 1. Fill topic and content name
  const topicInput = page.locator('textarea.topic-textarea');
  await topicInput.fill('How to Build an AI Tuber with Antigravity');

  // 2. Click Generate button
  const generateBtn = page.getByRole('button', { name: /Tạo Video/i });
  await expect(generateBtn).toBeEnabled();
  await generateBtn.click();

  // 3. Verify Progress Tracker appears
  await expect(page.getByText(/Đang tạo video/i)).toBeVisible();
  
  // 4. Wait for Completion (polling simulation via our mock)
  // 100% progress should be reached eventually
  await expect(page.getByText(/Video đã sẵn sàng/i)).toBeVisible({ timeout: 30000 });
  
  // 5. Verify Result Preview
  await expect(page.getByText(/Tải Video/i)).toBeVisible();
  
  // 6. Test Reset
  await page.getByRole('button', { name: /Tạo lại/i }).click();
  await expect(page.getByText(/Đang tạo video/i)).not.toBeVisible();
});
