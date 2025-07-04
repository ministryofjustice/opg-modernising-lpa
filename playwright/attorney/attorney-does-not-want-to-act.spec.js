import { expect, test } from '@playwright/test';
import { randomAccessCode, screenshot, TestEmail } from '../e2e.js';
import { extractTextFromMainAndSave } from "../textExtractor.js";

test('attorney opts out of being an attorney', async ({ page }) => {
    const accessCode = randomAccessCode()

    await page.goto(`/fixtures/attorney?redirect=&lpa-type=property-and-affairs&lpa-language=en&progress=signedByCertificateProvider&withAccessCode=${accessCode}&email=${TestEmail}`);

    await page.goto('/attorney-enter-access-code-opt-out');

    await expect(page.locator('h1')).toContainText('Enter your attorney access code');
    await page.getByRole('textbox', { name: 'Donor’s last name' }).fill('Smith');
    await page.getByRole('textbox', { name: 'Access code' }).fill(accessCode);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page.locator('h1')).toContainText('Confirm you do not want to be an attorney');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Confirm' }).click();

    await expect(page.locator('h1')).toContainText('You have confirmed that you do not want to be Sam Smith’s attorney.');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
});
