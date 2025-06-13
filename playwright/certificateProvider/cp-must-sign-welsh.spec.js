import { test, expect } from '@playwright/test';
import {screenshot} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";

test('certificate provider must sign in Welsh', async ({ page }) => {
  await page.goto('/fixtures');
  await page.getByRole('link', { name: 'Certificate provider' }).click();
  await page.getByRole('radio', { name: 'Welsh' }).check();
  await page.getByRole('radio', { name: 'Confirm your identity' }).check();
  await page.getByRole('button', { name: /Start|Start now/ }).click();
  await page.getByRole('tab', { name: 'I’m a certificate provider' }).click();
  await page.getByRole('link', { name: 'Go to task list' }).click();
  await page.getByRole('link', { name: 'Provide your certificate' }).click();

  await expect(page.getByLabel('Important')).toContainText('Sam Smith has chosen to have their LPA registered in Welsh');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
      await page.getByRole('link', { name: 'English' }).click();
          await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('What happens next');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Continue' }).click();

  await expect(page.getByLabel('Important')).toContainText('To provide Sam Smith’s certificate, you must view and sign it in Welsh.');
  await page.getByRole('checkbox', { name: 'I, Charlie Cooper, agree' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Submit signature' }).click();

  await expect(page.getByLabel('There is a problem')).toContainText('There is a problem To sign the certificate, you must view it in Welsh');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'View this page in Welsh' }).click();

  await expect(page.locator('h1')).toContainText('Darparu’r dystysgrif ar gyfer yr LPA hon');
  await page.getByRole('checkbox', { name: 'Yr wyf i, Charlie Cooper, yn' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Cyflwyno llofnod' }).click();
  await page.getByRole('link', { name: 'Change language to   English' }).click();

  await expect(page.locator('h1')).toContainText('You’ve provided the certificate for this LPA');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
});
