import {expect, test} from '@playwright/test';
import {randomShareCode, screenshot, TestEmail} from '../e2e.js'
import {extractTextFromMainAndSave} from "../textExtractor.js";

test('certificate provider opts out of being a certificate provider (from email link)', async ({page}) => {
    const shareCode = randomShareCode()

    await page.goto(`/fixtures/certificate-provider?redirect=&lpa-type=property-and-affairs&lpa-language=en&progress=paid&withShareCode=${shareCode}&email=${TestEmail}`);

    await page.goto('/certificate-provider-enter-reference-number-opt-out');

    await expect(page.locator('h1')).toContainText('Add an LPA');
    await page.getByRole('textbox', {name: 'Enter your access code'}).fill(shareCode);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('textbox', {name: 'Enter your access code'});
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Confirm you do not want to be the certificate provider');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Confirm'}).click();

    await expect(page.locator('h1')).toContainText('You have confirmed that you do not want to be Sam Smith’s certificate provider.');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
});
