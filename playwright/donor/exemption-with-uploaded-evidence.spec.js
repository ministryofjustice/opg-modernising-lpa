
import { test, expect } from '@playwright/test';
import {screenshot} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";
import path from 'path.js'

test('donor applies for a fee exemption and uploads evidence', async ({ page }) => {
  await page.goto('/fixtures');
  await page.getByRole('radio', { name: 'Check and send to your' }).check();
  await page.getByRole('button', { name: /Start|Start now/ }).click();
  await page.getByRole('link', { name: 'Go to task list' }).click();
  await page.getByRole('link', { name: 'Pay for the LPA' }).click();
  await page.getByRole('link', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Are you applying for any type of fee discount or exemption?');
  await page.getByRole('radio', { name: 'Yes' }).check();
  await screenshot(page)
  await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('What would you like to apply for?');
  await page.getByRole('radio', { name: 'No fee (an exemption)' }).check();
  await screenshot(page)
  await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('Evidence required to pay no fee');
  await screenshot(page)
  await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('How would you like to send us your evidence?');
  await page.getByRole('radio', { name: 'Upload it online' }).check();
  await screenshot(page)
  await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Upload your evidence');
  await page.getByRole('button', { name: 'Upload a file' }).setInputFiles([
    path.join(__dirname, 'upload-file-1.jpg'),
    path.join(__dirname, 'upload-file-2.jpg'),
  ]);
  await page.getByRole('button', { name: 'Upload files' }).click();
    await screenshot(page)
  await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.getByLabel('Important').getByRole('paragraph')).toContainText('We are reviewing the evidence you sent about your LPA fee');
  await screenshot(page)
  await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Return to task list' }).click();
});
