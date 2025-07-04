import { expect, test } from '@playwright/test';
import { randomAccessCode, screenshot, TestEmail } from '../e2e.js';
import { extractTextFromMainAndSave } from "../textExtractor.js";

test('certificate provider opts out of being a certificate provider (from email link)', async ({ page }) => {
    const accessCode = randomAccessCode()

    await page.goto(`/fixtures/certificate-provider?redirect=&lpa-type=property-and-affairs&lpa-language=en&progress=paid&withAccessCode=${accessCode}&email=${TestEmail}`);

    await page.goto('/certificate-provider-enter-access-code-opt-out');

    await expect(page.locator('h1')).toContainText('Add an LPA');
    await page.getByRole('textbox', { name: 'Donor’s last name' }).fill('Smith');
    await page.getByRole('textbox', { name: 'Access code' }).fill(accessCode);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('textbox', { name: 'Enter your access code' });
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page.locator('h1')).toContainText('Confirm you do not want to be the certificate provider');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Confirm' }).click();

    await expect(page.locator('h1')).toContainText('You have confirmed that you do not want to be Sam Smith’s certificate provider.');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
});
