
import { test, expect } from '@playwright/test';
import {screenshot} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";

test('voucher confirms donors details do not match their identity', async ({ page }) => {
  await page.goto('/fixtures');
  await page.getByRole('link', { name: 'Voucher' }).click();
  await page.getByRole('radio', { name: 'Confirm your name' }).check();
  await page.getByRole('button', { name: /Start|Start now/ }).click();
  await page.getByRole('tab', { name: 'I’m vouching for someone' }).click();
  await page.getByRole('link', { name: 'Go to task list' }).click();
  await page.getByRole('link', { name: 'Verify Sam Smith’s identity' }).click();
  await page.getByRole('radio', { name: 'No' }).check();
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.getByLabel('Important').getByRole('paragraph')).toContainText('You have told us that the details do not match Sam Smith’s identity.');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
});
