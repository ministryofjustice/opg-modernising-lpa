import { test, expect } from '@playwright/test';
import {screenshot} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";


test('voucher completes their journey', async ({ page }) => {
  await page.goto('/fixtures');
  await page.getByRole('link', { name: 'Voucher' }).click();
  await page.getByRole('radio', { name: 'Confirm your name' }).check();
  await page.getByRole('button', { name: /Start|Start now/ }).click();
  await page.getByRole('tab', { name: 'I’m vouching for someone' }).click();
  await page.getByRole('link', { name: 'Go to task list' }).click();
  await page.getByRole('link', { name: 'Confirm your name' }).click();

  await expect(page.locator('h1')).toContainText('Confirm your name');
  await screenshot(page)
  await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Change   last name' }).click();

  await expect(page.locator('h1')).toContainText('Your name');
  await page.getByRole('textbox', { name: 'Last name' }).click();
  await page.getByRole('textbox', { name: 'Last name' }).fill('Smith');
  await page.getByRole('button', { name: 'Save and continue' }).click();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('#main-content')).toContainText('Confirm that you are allowed to vouch');
  await page.getByRole('radio', { name: 'Yes' }).check();
      await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('link', { name: 'Verify Sam Smith’s identity' }).click();

  await expect(page.locator('h1')).toContainText('Verify Sam Smith’s identity');
  await page.getByRole('radio', { name: 'Yes' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('link', { name: 'Confirm your identity' }).click();

  await expect(page.locator('h1')).toContainText('Confirm your identity');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('radio', { name: 'Custom' }).check();
  await page.getByRole('textbox', { name: 'First names' }).fill('Vivian');
  await page.getByRole('textbox', { name: 'Last name' }).fill('Smith');
  await page.getByRole('textbox', { name: 'Day' }).fill('01');
  await page.getByRole('textbox', { name: 'Month' }).fill('01');
  await page.getByRole('textbox', { name: 'Year' }).fill('2000');
  await page.getByRole('textbox', { name: 'Building number' }).fill('1');
  await page.getByRole('textbox', { name: 'Postcode' }).fill('LE12 6AL');
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('#main-content')).toContainText('Confirm that you are allowed to vouch');
  await page.getByRole('radio', { name: 'Yes' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('link', { name: 'Sign the declaration' }).click();

  await expect(page.locator('h1')).toContainText('Your declaration');
  await page.getByRole('checkbox', { name: 'To the best of my knowledge,' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Submit my signature' }).click();
  await expect(page.locator('h1')).toContainText('Thank you');
          await screenshot(page)
    await extractTextFromMainAndSave(page)

  await page.getByRole('link', { name: 'Manage your LPAs' }).click();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
});
