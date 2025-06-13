import {expect, test} from '@playwright/test';
import {randomShareCode, screenshot, TestEmail} from '../e2e.js'
import {extractTextFromMainAndSave} from "../textExtractor.js";

test('attorney opts out of being an attorney', async ({page}) => {
    const shareCode = randomShareCode()

    await page.goto(`/fixtures/attorney?redirect=&lpa-type=property-and-affairs&lpa-language=en&progress=signedByCertificateProvider&withShareCode=${shareCode}&email=${TestEmail}`);

    await page.goto('/attorney-enter-reference-number-opt-out');

    await expect(page.locator('h1')).toContainText('Enter your attorney access code');
    await page.getByRole('textbox', {name: 'Enter code'}).fill(shareCode);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Confirm you do not want to be an attorney');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Confirm'}).click();

    await expect(page.locator('h1')).toContainText('You have confirmed that you do not want to be Sam Smith’s attorney.');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
});
