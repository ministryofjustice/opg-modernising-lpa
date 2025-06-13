import { test, expect } from '@playwright/test';
import {screenshot} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";

test('certificate provider provides certificate (personal)', async ({ page }) => {
  await page.goto('/fixtures');
  await page.getByRole('link', { name: 'Certificate provider' }).click();
  await page.getByRole('radio', { name: 'Signed by donor' }).check();
  await page.getByRole('button', { name: /Start|Start now/ }).click();
  await page.getByRole('tab', { name: 'I’m a certificate provider' }).click();
  await page.getByRole('link', { name: 'Go to task list' }).click();

  await expect(page.locator('h1')).toContainText('Your task list');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Confirm your details' }).click();

  await expect(page.locator('h1')).toContainText('Enter your date of birth');
  await page.getByRole('textbox', { name: 'Day' }).fill('02');
  await page.getByRole('textbox', { name: 'Month' }).fill('01');
  await page.getByRole('textbox', { name: 'Year' }).fill('1990');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('Your preferred language');
  await page.getByRole('radio', { name: 'English' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('Confirm your details');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('link', { name: 'Confirm your identity' }).click();

  await expect(page.locator('h1')).toContainText('Confirm your identity');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('radio', { name: 'Charlie Cooper (certificate' }).check();
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.getByLabel('Success').locator('h2')).toContainText('You have successfully confirmed your identity');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Return to task list' }).click();
  await page.getByRole('link', { name: 'Provide your certificate' }).click();

  await expect(page.locator('h1')).toContainText('Read the LPA');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('What happens next');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Provide the certificate for this LPA');
  await page.getByRole('checkbox', { name: 'I, Charlie Cooper, agree' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Submit signature' }).click();

  await expect(page.locator('h1')).toContainText('You’ve provided the certificate for this LPA');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Go to your dashboard' }).click();
});
